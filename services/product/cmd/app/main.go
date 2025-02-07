package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/dzhordano/ecom-thing/services/product/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/product/internal/config"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure/profiling"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/product/internal/interfaces/grpc"
)

// 04.02:
// Unit Тесты на домен.
// Запустить профилирование + Нагрузочное.
// sync.Pool for objects? [Мб для конвертации выделить как-то пулы, иначе оч много alloc_objects]

// TODO:
// gRPC-Gateway. OpenAPI.
// JWT. [Тоже логика в интерцепторе]
// TLS.
// Redis. [Мб сейвить количество продуктов, чтобы нагрузка на минус + другие сервисы (inv) получали быстрее ответ]

var (
	pprof bool
)

func main() {
	flag.BoolVar(&pprof, "pprof", false, "specify to run profiling server on PPROF_PORT")
	flag.Parse()

	cfg := config.MustNew()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()

	db := pg.MustNewPGXPool(ctx, cfg.PG.DSN())

	repo := pg.NewProductRepository(db)

	productService := service.NewProductService(log, repo)

	handler := grpc.NewProductHandler(productService)

	server := grpc.MustNew(log, handler,
		grpc.WithAddr(net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port)),
		grpc.WithRateLimiter(cfg.RateLimiter.Limit, cfg.RateLimiter.Burst),
		grpc.WithCircuitBreakerSettings(
			cfg.CircuitBreaker.MaxRequests,
			cfg.CircuitBreaker.Interval,
			cfg.CircuitBreaker.Timeout),
	)

	// Run Metrics Server
	go infrastructure.RunMetricsServer(net.JoinHostPort(cfg.GRPC.Host, cfg.Prometheus.Port))

	// Run Profiling Server. TODO: Run this only with a specific flag
	if pprof {
		go profiling.Run(net.JoinHostPort(cfg.GRPC.Host, cfg.Pprof.Port))
	}

	q := make(chan os.Signal, 1)

	signal.Notify(q, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info("starting grpc server")
		if err := server.Run(); err != nil {
			log.Error("failed to run grpc server", slog.String("error", err.Error()))
		}
	}()

	<-q

	log.Info("stopping grpc server")
	server.GracefulStop()

	// TODO graceful
}
