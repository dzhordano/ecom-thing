package interceptors

import (
	"context"
	"time"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/infrastructure"
	"google.golang.org/grpc"
)

// FIXME в идеале везде нужна обертка, чтобы полностью изолироваться от infrastructure
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
