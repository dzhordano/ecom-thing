package main

import (
	"context"
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
	grpc_server "github.com/dzhordano/ecom-thing/services/payment/internal/interfaces/grpc"
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

	srv := grpc_server.MustNew(
		log,
		grpc_server.NewPaymentHandler(svc),
		grpc_server.WithAddr(cfg.GRPC.Addr()),
	)

	// TODO поменять, чтобы я тут не импоритровал саму сараму.
	kafkaProducer := kafka.NewPaymentsSyncProducer(cfg.Kafka.Brokers)
	defer kafkaProducer.Close()

	go func() {
		c, err := kafka.NewConsumerGroup(
			ctx,
			cfg.Kafka.Brokers,
			cfg.Kafka.GroupID,
			svc,
		)
		if err != nil {
			log.Error("error starting consumer group", "error", err)
			return
		}
		if err := c.Start(ctx, cfg.Kafka.Topics); err != nil {
			log.Error("error starting consumer group", "error", err)
		}
	}()

	// TODO тута хардкод
	outboxWorker := outbox.NewOutboxProcessor(log, db, kafkaProducer, 5*time.Second, billingSvc)
	go outboxWorker.Start(ctx)

	q := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	go srv.Run(ctx)

	<-q

	shutdownWG.Add(1)
	go func() {
		defer shutdownWG.Done()
		srv.GracefulStop()
	}()

	shutdownWG.Wait()

	log.Info("graceful shutdown completed")
}
