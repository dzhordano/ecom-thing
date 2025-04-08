package main

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/outbox"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/tracing/tracer"
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

	tp, err := tracer.NewTracerProvider(cfg.Tracing.URL, "order")
	if err != nil {
		log.Error("error creating tracer provider", "error", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Error("error shutting down tracer provider", "error", err)
		}
	}()
	tracer.SetGlobalTracerProvider(tp)

	repo := pg.NewOrderRepository(db)

	ps := product.NewProductClient(cfg.GRPCProduct.Addr(), product.WithTracing(tp))

	is := inventory.NewInventoryClient(cfg.GRPCInventory.Addr(), inventory.WithTracing(tp))

	kafkaProducer, err := kafka.NewOrdersSyncProducer(cfg.Kafka.Brokers)
	if err == nil {
		outboxWorker := outbox.NewOutboxProcessor(log, db, kafkaProducer, 5*time.Second)
		go outboxWorker.Start(ctx)
		defer kafkaProducer.Close()
	} else {
		log.Error("error creating kafka producer", "error", err)
	}

	// TODO тута хардкод

	svc := service.NewOrderService(log, ps, is, repo)

	srv := grpc_server.MustNew(
		log,
		grpc_server.NewItemHandler(svc),
		grpc_server.WithAddr(cfg.GRPC.Addr()),
		grpc_server.WithTracerProvider(tp),
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
