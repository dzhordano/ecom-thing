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

// TODO to finish
// клиенты. метрики. трейсы. тесты. деплой. очередь.
// улучшить логгер.

func main() {

	ctx := context.Background()

	cfg := config.MustNew()

	log := logger.NewZapLogger(cfg.LogLevel, []string{"stdout"}, []string{"stderr"}) // FIXME опять хардкод

	db := pg.MustNewPGXPool(ctx, cfg.PG.DSN())

	repo := pg.NewOrderRepository(db)

	svc := service.NewOrderService(log, repo)

	handler := grpc_server.NewItemHandler(svc)

	srv := grpc_server.MustNew(log, handler,
		grpc_server.WithAddr(cfg.GRPC.Addr()),
	)

	if err := srv.Run(); err != nil {
		log.Panic("errors running grpc server", zap.Error(err))
	}
	// Shutdown
}
