package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/dzhordano/ecom-thing/services/product/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/product/internal/config"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/product/internal/interfaces/grpc_server"
	"github.com/dzhordano/ecom-thing/services/product/pkg/logger"
	"go.uber.org/zap"
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

	cfg := config.MustNew()

	log := logger.NewZapLogger(
		cfg.Logger.Level,
		logger.WithEncoding(cfg.Logger.Encoding),
		logger.WithOutputPaths(cfg.Logger.OutputPaths),
		logger.WithErrorOutputPaths(cfg.Logger.ErrorOutputPaths),
		logger.WithFileOutput(cfg.Logger.OutputFilePath),
		logger.WithFileErrorsOutput(cfg.Logger.ErrorOutputFilePath),
	)

	ctx := context.Background()

	db := pg.MustNewPGXPool(ctx, cfg.PG.DSN())

	// migrate.MustMigrateUpWithNoChange(cfg.PG.URL())

	repo := pg.NewProductRepository(db)

	productService := service.NewProductService(log, repo)

	handler := grpc_server.NewProductHandler(productService)

	server := grpc_server.MustNew(log, handler,
		grpc_server.WithAddr(net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port)),
		grpc_server.WithRateLimiter(cfg.RateLimiter.Limit, cfg.RateLimiter.Burst),
		grpc_server.WithGoBreakerSettings(
			cfg.CircuitBreaker.MaxRequests,
			cfg.CircuitBreaker.Interval,
			cfg.CircuitBreaker.Timeout),
		grpc_server.WithProfiling(cfg.ProfilingEnabled),
	)

	q := make(chan os.Signal, 1)

	signal.Notify(q, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := server.Run(); err != nil {
			log.Error("failed to run grpc server", zap.Error(err))
		}
	}()

	<-q

	log.Info("stopping grpc server")
	server.GracefulStop()

	// TODO graceful
}
