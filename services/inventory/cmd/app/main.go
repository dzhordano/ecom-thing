package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	repo := pg.NewPGRepository(ctx, pool)

	svc := service.NewItemService(log, repo)

	srv := grpc_server.MustNew(
		log,
		grpc_server.NewItemHandler(svc),
		grpc_server.WithAddr(cfg.GRPC.Addr()),
		grpc_server.WithTracing(cfg.Tracing.URL),
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
