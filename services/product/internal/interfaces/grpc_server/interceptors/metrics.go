package interceptors

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure"
	"google.golang.org/grpc"
	"time"
)

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
			infrastructure.IncResponseCounter.WithLabelValues(info.FullMethod, err.Error()).Inc()
			infrastructure.HistRequestDuration.WithLabelValues(info.FullMethod, err.Error()).Observe(latency)
		} else {
			infrastructure.IncResponseCounter.WithLabelValues(info.FullMethod, "OK").Inc()
			infrastructure.HistRequestDuration.WithLabelValues(info.FullMethod, "OK").Observe(latency)
		}

		return resp, err
	}
}
