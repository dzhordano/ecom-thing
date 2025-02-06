package grpc

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"log/slog"
	"time"
)

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func ErrorMapperInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, MapError(err)
		}

		return resp, nil
	}
}

func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		infrastructure.IncRequestCounter.WithLabelValues(info.FullMethod).Inc()

		resp, err := handler(ctx, req)
		latency := time.Since(start).Seconds()

		if err != nil {
			infrastructure.IncResponseCounter.WithLabelValues(info.FullMethod, "ERROR").Inc()
			infrastructure.HistRequestDuration.WithLabelValues(info.FullMethod, "ERROR").Observe(latency)
		} else {
			infrastructure.IncResponseCounter.WithLabelValues(info.FullMethod, "OK").Inc()
			infrastructure.HistRequestDuration.WithLabelValues(info.FullMethod, "OK").Observe(latency)
		}

		return resp, err
	}
}
