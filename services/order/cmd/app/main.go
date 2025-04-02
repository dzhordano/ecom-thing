package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/order/internal/config"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/grpc/inventory"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/grpc/product"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/kafka"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/outbox"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/order/internal/interfaces/grpc_server"
	"github.com/dzhordano/ecom-thing/services/order/pkg/logger"
)

// TODO to finish
// тесты. деплой. очередь (после payment apiшки).
// улучшить логгер (просто избаться от зависимости zap в сервисах для начала).

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

	repo := pg.NewOrderRepository(db)

	ps := product.NewProductClient(cfg.GRPCProduct.Addr())

	is := inventory.NewInventoryClient(cfg.GRPCInventory.Addr())

	// TODO поменять, чтобы я тут не импоритровал саму сараму.
	kafkaProducer := kafka.NewOrdersSyncProducer(cfg.Kafka.Brokers)
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
		if err := c.Start(ctx, []string{"payment-events"}); err != nil {
			log.Error("error starting consumer group", "error", err)
		}
	}()

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
