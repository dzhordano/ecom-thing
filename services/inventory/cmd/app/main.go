package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/infrastructure/tracing/tracer"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/config"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/infrastructure/kafka"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/interfaces/grpc_server"
	"github.com/dzhordano/ecom-thing/services/inventory/pkg/logger"
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

	pool := pg.MustNewPGXPool(ctx, cfg.PG.DSN())
	defer pool.Close()

	repo := pg.NewInventoryRepository(ctx, pool)

	svc := service.NewItemService(log, repo)

	tp, err := tracer.NewTracerProvider(cfg.Tracing.URL, "inventory")
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

	srv := grpc_server.MustNew(
		log,
		grpc_server.NewItemHandler(svc),
		grpc_server.WithAddr(cfg.GRPC.Addr()),
		grpc_server.WithTracerProvider(tp),
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
		if err := c.Start(ctx, cfg.Kafka.Topics); err != nil {
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
