package main

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/order/internal/config"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/grpc/inventory"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/grpc/product"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/kafka"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/outbox"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/order/internal/interfaces/grpc_server"
	"github.com/dzhordano/ecom-thing/services/order/pkg/logger"
	"go.uber.org/zap"
)

// TODO to finish
// тесты. деплой. очередь (после payment apiшки).
// улучшить логгер (просто избаться от зависимости zap в сервисах для начала).

func main() {

	ctx := context.Background()

	cfg := config.MustNew()

	// WARNING.
	log := logger.NewZapLogger(
		cfg.Logger.Level,
		logger.WithEncoding(cfg.Logger.Encoding),
		logger.WithOutputPaths(cfg.Logger.OutputPaths),
		logger.WithErrorOutputPaths(cfg.Logger.ErrorOutputPaths),
	)

	db := pg.MustNewPGXPool(ctx, cfg.PG.DSN())
	defer db.Close()

	repo := pg.NewOrderRepository(db)

	ps := product.NewProductClient(cfg.GRPCProduct.Addr())

	is := inventory.NewInventoryClient(cfg.GRPCInventory.Addr())

	// TODO поменять, чтобы я тут не импоритровал саму сараму.
	kafkaProducer := kafka.NewOrdersSyncProducer(
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

	// TODO тута хардкод
	outboxWorker := outbox.NewOutboxProcessor(log, db, kafkaProducer, 5*time.Second)
	go outboxWorker.Start(ctx)

	svc := service.NewOrderService(log, ps, is, repo)

	srv := grpc_server.MustNew(
		log,
		grpc_server.NewItemHandler(svc),
		grpc_server.WithAddr(cfg.GRPC.Addr()),
		// FIXME ещо
	)

	if err := srv.Run(); err != nil {
		log.Panic("errors running grpc server", zap.Error(err))
	}

	// Shutdown
}
