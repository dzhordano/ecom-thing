package main

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/product/internal/config"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/product/internal/interfaces/grpc"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// 04.02:
// Unit Тесты на домен.
// Load Тесты.
// Rate-Limiter. Circuit Breaker.
// Запустить профилирование + Нагрузочное.

// TODO:
// gRPC-Gateway. OpenAPI.
// JWT. [Тоже логика в интерцепторе]
// TLS.
// Redis. [Мб сейвить количество продуктов, чтобы нагрузка на минус + другие сервисы (inv) получали быстрее ответ]

func main() {
	cfg := config.MustNew()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()

	db := pg.MustNewPGXPool(ctx, cfg.PG.DSN())

	repo := pg.NewProductRepository(db)

	productService := service.NewProductService(log, repo)

	handler := grpc.NewProductHandler(productService)

	server := grpc.MustNew(log, cfg.GRPC.Addr(), handler)

	go infrastructure.RunMetricsServer(net.JoinHostPort(cfg.GRPC.Host, cfg.Prometheus.Port))

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
