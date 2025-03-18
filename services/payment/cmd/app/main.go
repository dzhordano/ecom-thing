package main

import (
	"context"
	"time"

	"github.com/IBM/sarama"
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

	cfg := config.MustNew()

	log := logger.NewZapLogger(
		cfg.Logger.Level,
		logger.WithEncoding(cfg.Logger.Encoding),
		logger.WithOutputPaths(cfg.Logger.OutputPaths),
		logger.WithErrorOutputPaths(cfg.Logger.ErrorOutputPaths),
	)

	db := pg.MustNewPGXPool(ctx, cfg.PG.DSN())
	defer db.Close()

	repo := pg.NewPaymentRepository(db)

	billingSvc := billing.NewStubBilling()

	svc := service.NewPaymerService(log, repo)

	srv := grpc_server.MustNew(
		log,
		grpc_server.NewPaymentHandler(svc),
		grpc_server.WithAddr(cfg.GRPC.Addr()),
		// FIXME ещо
	)

	// TODO поменять, чтобы я тут не импоритровал саму сараму.
	kafkaProducer := kafka.NewPaymentsSyncProducer(
		[]string{"localhost:19092"},
		func() *sarama.Config {
			producerConfig := sarama.NewConfig()

			producerConfig.Net.MaxOpenRequests = 1
			producerConfig.Producer.RequiredAcks = sarama.WaitForAll
			producerConfig.Producer.Return.Successes = true

			return producerConfig
		},
	)
	defer kafkaProducer.Close()

	c, err := kafka.NewConsumerGroup(
		[]string{"localhost:19092"},
		"payment-group",
		svc,
	)
	if err != nil {
		log.Error("error starting consumer group", "error", err)
	}

	go c.Start(ctx, []string{"order-events"})

	// TODO тута хардкод
	outboxWorker := outbox.NewOutboxProcessor(log, db, kafkaProducer, 5*time.Second, billingSvc)
	go outboxWorker.Start(ctx)

	if err := srv.Run(); err != nil {
		log.Error("run grpc server error", "error", err)
	}

	//run server & outbox cron
	//graceful
}
