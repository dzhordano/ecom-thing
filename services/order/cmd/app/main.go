package main

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/order/internal/config"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/order/internal/interfaces/grpc_server"
	"github.com/dzhordano/ecom-thing/services/order/pkg/logger"
	"go.uber.org/zap"
)

func main() {

	ctx := context.Background()

	cfg := config.MustNew()
	// Logger
	log := logger.NewZapLogger(cfg.LogLevel, []string{"stdout"}, []string{"stderr"}) // FIXME опять хардкод
	// Deps
	db := pg.MustNewPGXPool(ctx, cfg.PG.DSN())

	repo := pg.NewOrderRepository(db)

	svc := service.NewOrderService(log, repo)

	handler := grpc_server.NewItemHandler(svc)

	srv := grpc_server.MustNew(log, handler,
		grpc_server.WithAddr(cfg.GRPC.Addr()))
	// Run
	if err := srv.Run(); err != nil {
		log.Panic("errors running grpc server", zap.Error(err))
	}
	// Shutdown
}
