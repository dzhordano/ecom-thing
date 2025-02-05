package main

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/product/internal/config"
	grpcServer "github.com/dzhordano/ecom-thing/services/product/internal/infrastructure/grpc"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure/repository/pg"
	"log/slog"
	"os"
)

// 04.02:
// Unit Тесты на сервис (мокать репозиторий)
// Intergration + Load Тесты.
// Если успею:
// Rate-Limiter. Circ	uit Breaker.
// (После метрик) Запустить профилирование + Нагрузочное.

// TODO:
// Метрики. (Prom. Grafana.) [Думаю через интерцептор]
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

	handler := grpcServer.NewProductHandler(productService)

	server := grpcServer.MustNew(log, cfg.GRPC.Addr(), handler)

	if err := server.Run(); err != nil {
		log.Error("failed to run grpc server", slog.String("error", err.Error()))
	}

	// TODO graceful
}
