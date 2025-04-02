package interceptors

import (
	"context"
	"errors"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// HashMap w/ domain errors for efficient mapping.
	errorMap = map[error]codes.Code{
		domain.ErrNotEnoughQuantity: codes.InvalidArgument,
		domain.ErrProductNotFound:   codes.NotFound,
		domain.ErrOperationUnknown:  codes.Unknown,
	}
)

func mapError(err error) error {
	if s, ok := status.FromError(err); ok {
		return s.Err() // Return status error if it's a gRPC error
	}

	for unwrappedErr := err; unwrappedErr != nil; unwrappedErr = errors.Unwrap(unwrappedErr) {
		if code, ok := errorMap[unwrappedErr]; ok {
			return status.Error(code, unwrappedErr.Error())
		}
	}

	return status.Error(codes.Internal, "internal error")
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
			return nil, mapError(err)
		}

		return resp, nil
	}
}
