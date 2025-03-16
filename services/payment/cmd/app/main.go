package main

import (
	"context"
	"sync"

	"github.com/dzhordano/ecom-thing/services/payment/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/payment/internal/config"
	"github.com/dzhordano/ecom-thing/services/payment/internal/infrastructure/billing"
	"github.com/dzhordano/ecom-thing/services/payment/internal/infrastructure/repository/pg"
	grpc_server "github.com/dzhordano/ecom-thing/services/payment/internal/interfaces/grpc"
	"github.com/dzhordano/ecom-thing/services/payment/pkg/logger"
)

func main() {
	ctx := context.Background()
	waitWg := sync.WaitGroup{}

	cfg := config.MustNew()

	log := logger.NewZapLogger(
		cfg.Logger.Level,
		logger.WithEncoding(cfg.Logger.Encoding),
		logger.WithOutputPaths(cfg.Logger.OutputPaths),
		logger.WithErrorOutputPaths(cfg.Logger.ErrorOutputPaths),
	)

	db := pg.MustNewPGXPool(ctx, cfg.PG.DSN())
	defer db.Close()

	repo := pg.NewPaymentRepository(db)

	billingSvc := billing.NewStubBilling()

	svc := service.NewPaymerService(log, repo, billingSvc, &waitWg)

	srv := grpc_server.MustNew(
		log,
		grpc_server.NewPaymentHandler(svc),
		grpc_server.WithAddr(cfg.GRPC.Addr()),
		// FIXME ещо
	)

	if err := srv.Run(); err != nil {
		log.Error("run grpc server error", "error", err)
	}

	//run server & outbox cron
	//graceful

	waitWg.Wait()
}
