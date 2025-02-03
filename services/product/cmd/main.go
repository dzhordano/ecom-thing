package main

import (
	"github.com/dzhordano/ecom-thing/services/product/internal/config"
	grpcServer "github.com/dzhordano/ecom-thing/services/product/internal/infrastructure/grpc"
	"log/slog"
	"os"
)

func main() {
	cfg := config.MustNew()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	server := grpcServer.MustNew(log, cfg.GRPC.Addr(), grpcServer.NewProductHandler())

	if err := server.Run(); err != nil {
		log.Error("failed to run grpc server", err)
	}

}
