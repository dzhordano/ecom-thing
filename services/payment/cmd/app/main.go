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
	ctx := context.Background()
	shutdownWG := &sync.WaitGroup{}

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

	shutdownWG.Add(1)
	tp, err := tracer.NewTracerProvider(cfg.Tracing.URL, "payment")
	if err != nil {
		log.Error("error creating tracer provider", "error", err)
	}
	defer func() {
		defer shutdownWG.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Error("error shutting down tracer provider", "error", err)
			return
		}
		log.Debug("tracer provider closed")
	}()
	tracer.SetGlobalTracerProvider(tp)

	var kp *kafka.PaymentsSyncProducer
	go func() {
		kp, err = kafka.NewPaymentsSyncProducer(cfg.Kafka.Brokers)
		if err != nil {
			log.Error("error creating kafka producer", "error", err)
			return
		}

		outboxWorker := outbox.NewOutboxProcessor(log, db, kp, 5*time.Second, billingSvc)
		go outboxWorker.Start(ctx)
	}()
	defer func() {
		if err := kp.Close(); err != nil {
			log.Error("error closing kafka producer", "error", err)
		}
	}()

	var cg *kafka.Consumer
	go func() {
		cg, err = kafka.NewConsumerGroup(
			ctx,
			cfg.Kafka.Brokers,
			cfg.Kafka.GroupID,
			svc,
		)
		if err != nil {
			log.Error("error starting consumer group", "error", err)
			return
		}
		if err := cg.Start(ctx, cfg.Kafka.Topics); err != nil {
			log.Error("error starting consumer group", "error", err)
		}
	}()
	defer func() {
		if err := cg.Close(); err != nil {
			log.Error("error closing consumer group", "error", err)
		}
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

	<-q

	shutdownWG.Add(1)
	go func() {
		defer shutdownWG.Done()
		srv.GracefulStop()
	}()

	shutdownWG.Wait()

	log.Info("graceful shutdown completed")
}
