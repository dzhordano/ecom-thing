package main

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/order/internal/config"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/grpc/inventory"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/grpc/product"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/order/internal/interfaces/grpc_server"
	"github.com/dzhordano/ecom-thing/services/order/pkg/logger"
	"go.uber.org/zap"
)

// TODO to finish
// трейсы. тесты. деплой. очередь (после payment apiшки).
// улучшить логгер.

func main() {

	ctx := context.Background()

	cfg := config.MustNew()

	// WARNING.
	// When specifying file path for logs to save, logger WONT create a file.
	log := logger.NewZapLogger(
		cfg.Logger.Level,
		cfg.Logger.OutputPaths,
		cfg.Logger.ErrorOutputPaths,
	)

	db := pg.MustNewPGXPool(ctx, cfg.PG.DSN())

	repo := pg.NewOrderRepository(db)

	ps := product.NewProductClient(cfg.GRPCProduct.Addr())

	is := inventory.NewInventoryClient(cfg.GRPCInventory.Addr())

	svc := service.NewOrderService(log, ps, is, repo)

	handler := grpc_server.NewItemHandler(svc)

	srv := grpc_server.MustNew(log, handler,
		grpc_server.WithAddr(cfg.GRPC.Addr()),
	)

	if err := srv.Run(); err != nil {
		log.Panic("errors running grpc server", zap.Error(err))
	}
	// Shutdown
}
