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
	ctx, cancel := context.WithCancel(context.Background())

	// TODO думаю сюда норм errgroup залетит

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

	repo := pg.NewInventoryRepository(pool)

	svc := service.NewItemService(log, repo)

	tp, err := tracer.NewTracerProvider(cfg.Tracing.URL, "inventory")
	if err != nil {
		log.Panic("error creating tracer provider", "error", err)
	}
	tracer.SetGlobalTracerProvider(tp)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Error("error shutting down tracer provider", "error", err)
		}
	}()

	srv := grpc_server.MustNew(
		log,
		grpc_server.NewItemHandler(svc),
		grpc_server.WithAddr(cfg.GRPC.Addr()),
		grpc_server.WithTracerProvider(tp),
	)

	wg := sync.WaitGroup{}

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

	go func() {
		if err := srv.Run(ctx); err != nil {
			log.Panic("error running server", "error", err)
		}
	}()

	// TODO вроде можно объединить?
	<-q
	cancel()

	shutdownWG := sync.WaitGroup{}

	shutdownWG.Add(1)
	go func() {
		defer shutdownWG.Done()
		// Wait till resources are freed (closed)
		wg.Wait()
		srv.GracefulStop()
	}()

	// Wait till everything's shut down
	shutdownWG.Wait()

	log.Info("graceful shutdown completed")
}
