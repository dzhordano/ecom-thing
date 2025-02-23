package main

import (
	"context"

	_ "embed"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/config"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/interfaces/grpc_server"
	"github.com/dzhordano/ecom-thing/services/inventory/pkg/migrate"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	cfg := config.MustNew()

	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	pool := pg.MustNewPGXPool(ctx, cfg.PG.DSN())

	migrate.MustMigrateUpWithNoChange(cfg.PG.URL())

	repo := pg.NewPGRepository(ctx, pool)

	svc := service.NewItemService(log, repo)

	handlers := grpc_server.NewItemHandler(svc)

	srv := grpc_server.MustNew(log, handlers, grpc_server.WithAddr(cfg.GRPC.Addr()))

	if err := srv.Run(); err != nil {
		log.Error(err.Error())
	}
}
