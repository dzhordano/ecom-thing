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

	log := logger.NewZapLogger(
		cfg.Logger.Level,
		logger.WithEncoding(cfg.Logger.Encoding),
		logger.WithOutputPaths(cfg.Logger.OutputPaths),
		logger.WithErrorOutputPaths(cfg.Logger.ErrorOutputPaths),
		logger.WithFileOutput(cfg.Logger.OutputFilePath),
		logger.WithFileErrorsOutput(cfg.Logger.ErrorOutputFilePath),
	)

	pool := pg.MustNewPGXPool(ctx, cfg.PG.DSN())

	repo := pg.NewPGRepository(ctx, pool)

	svc := service.NewItemService(log, repo)

	handlers := grpc_server.NewItemHandler(svc)

	srv := grpc_server.MustNew(log, handlers, grpc_server.WithAddr(cfg.GRPC.Addr()))

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

	log.Info("graceful shutdown completed")

	shutdownWG.Add(1)
	go func() {
		defer shutdownWG.Done()
		pool.Close()
	}()

	shutdownWG.Wait()

	log.Info("graceful shutdown completed")
}

// TODO Не понимаю, почему после выключения:
//		failed to serve HTTP: mux: server closed
