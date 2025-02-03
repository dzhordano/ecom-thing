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

func main() {
	cfg := config.MustNew()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()

	db := pg.MustNewPGXPool(ctx, cfg.PG.DSN())

	repo := pg.NewProductRepository(db)

	productService := service.NewProductService(repo)

	handler := grpcServer.NewProductHandler(productService)

	server := grpcServer.MustNew(log, cfg.GRPC.Addr(), handler)

	if err := server.Run(); err != nil {
		log.Error("failed to run grpc server", err)
	}

}
