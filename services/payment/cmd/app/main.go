package main

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/payment/internal/infrastructure/tracing/tracer"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dzhordano/ecom-thing/services/payment/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/payment/internal/config"
	"github.com/dzhordano/ecom-thing/services/payment/internal/infrastructure/billing"
	"github.com/dzhordano/ecom-thing/services/payment/internal/infrastructure/kafka"
	"github.com/dzhordano/ecom-thing/services/payment/internal/infrastructure/outbox"
	"github.com/dzhordano/ecom-thing/services/payment/internal/infrastructure/repository/pg"
	grpc_server "github.com/dzhordano/ecom-thing/services/payment/internal/interfaces/grpc_server"
	"github.com/dzhordano/ecom-thing/services/payment/pkg/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}

	cfg := config.MustNew()

	log := logger.MustInit(
		cfg.Logger.Level,
		cfg.Logger.LogFile,
		cfg.Logger.Encoding,
		cfg.Logger.Development,
	)
	defer log.Sync()

	db := pg.MustNewPGXPool(ctx, cfg.PG.DSN())
	defer db.Close()

	repo := pg.NewPaymentRepository(db)

	billingSvc := billing.NewStubBilling()

	svc := service.NewPaymerService(log, repo)

	wg.Add(1)
	tp, err := tracer.NewTracerProvider(cfg.Tracing.URL, "payment")
	if err != nil {
		log.Panic("error creating tracer provider", "error", err)
	}
	tracer.SetGlobalTracerProvider(tp)
	defer func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Error("error shutting down tracer provider", "error", err)
			return
		}
		log.Debug("tracer provider closed")
	}()

	kp := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.TopicsToProduce, time.Second, 0)
	defer kp.Close()
	if err != nil {
		log.Error("error creating kafka producer", "error", err)
		return
	}
	outboxWorker := outbox.NewOutboxProcessor(log, db, kp, 5*time.Second, billingSvc)
	go outboxWorker.Start(ctx)

	// TODO хардкод
	go func() {
		cg, err := kafka.NewConsumerGroup(
			ctx,
			cfg.Kafka.Brokers,
			cfg.Kafka.TopicsToConsume,
			svc,
			time.Second,
			uint(100),
		)
		if err != nil {
			log.Error("error starting consumer group", "error", err)
			return
		}
		if err := cg.CreateTopics(ctx, 8, 2); err != nil {
			log.Error("error creating kafka topics", "error", err)
			return
		}
		log.Info("topics created", "topics", cfg.Kafka.TopicsToConsume)
		cg.RunConsumers(ctx, 2, &wg)
	}()

	q := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	srv := grpc_server.MustNew(
		log,
		grpc_server.NewPaymentHandler(svc),
		grpc_server.WithAddr(cfg.GRPC.Addr()),
		grpc_server.WithTracerProvider(tp),
	)

	go func() {
		if err := srv.Run(ctx); err != nil {
			log.Panic("error running server", "error", err)
		}
	}()

	// TODO вроде можно объединить?
	<-q
	cancel()

	shutdownWG := &sync.WaitGroup{}
	shutdownWG.Add(1)
	go func() {
		defer shutdownWG.Done()
		// Wait till resources are freed (closed)
		wg.Wait()
		srv.GracefulStop()
	}()

	shutdownWG.Wait()

	log.Info("graceful shutdown completed")
}
