package main

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure/tracing/tracer"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dzhordano/ecom-thing/services/product/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/product/internal/config"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/product/internal/interfaces/grpc_server"
	"github.com/dzhordano/ecom-thing/services/product/pkg/logger"
)

// Unit Тесты на домен.
// Запустить профилирование + Нагрузочное.
// sync.Pool for objects? [Мб для конвертации выделить как-то пулы, иначе оч много alloc_objects]
// Деплой локально (миникуб там манифесты написать).

// TODO:
// Redis. [Мб сейвить количество продуктов, чтобы нагрузка на минус + другие сервисы получали быстрее ответ]
// TLS.
// JWT. [Тоже логика в интерцепторе]

func main() {
	ctx := context.Background()

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

	repo := pg.NewProductRepository(db)

	productService := service.NewProductService(log, repo)

	tp, err := tracer.NewTracerProvider(cfg.Tracing.URL, "product")
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
		grpc_server.NewProductHandler(productService),
		grpc_server.WithAddr(net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port)),
		grpc_server.WithRateLimiter(cfg.RateLimiter.Limit, cfg.RateLimiter.Burst),
		grpc_server.WithGoBreakerSettings(
			cfg.CircuitBreaker.MaxRequests,
			cfg.CircuitBreaker.Interval,
			cfg.CircuitBreaker.Timeout),
		grpc_server.WithTracerProvider(tp),
		grpc_server.WithProfiling(),
	)

	q := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	go func() {
		if err := srv.Run(ctx); err != nil {
			log.Error("error running server", "error", err)
			panic(err)
		}
	}()

	<-q

	shutdownWG := &sync.WaitGroup{}

	shutdownWG.Add(1)
	go func() {
		defer shutdownWG.Done()
		srv.GracefulStop()
	}()

	shutdownWG.Wait()

	log.Info("graceful shutdown completed")
}
