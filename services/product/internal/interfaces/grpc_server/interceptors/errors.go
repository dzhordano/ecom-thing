package interceptors

import (
	"context"
	"errors"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	if s, ok := status.FromError(err); ok {
		return s.Err() // Return the status error if it's a gRPC error
	}

	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		code := appErr.GRPCCode()
		if code != codes.Internal {
			return status.Error(code, appErr.Error())
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
