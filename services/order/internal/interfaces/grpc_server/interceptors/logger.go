package interceptors

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/order/pkg/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

func InterceptorLogger(l logger.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, loggingLevelToStr(lvl), msg, fields...)
	})
}

// Just converts logging.Level to string.
//
// Defaults to level info.
func loggingLevelToStr(lvl logging.Level) string {
	switch lvl {
	case logging.LevelDebug:
		return "debug"
	case logging.LevelInfo:
		return "info"
	case logging.LevelWarn:
		return "warn"
	case logging.LevelError:
		return "error"
	default:
		return "info"
	}
}
