package interceptors

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/product/pkg/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

func InterceptorLogger(l logger.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, logger.Level(lvl), msg, fields...)
	})
}
