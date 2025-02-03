package grpc

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"log/slog"
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
			return nil, MapError(ctx, err)
		}

		return resp, nil
	}
}
