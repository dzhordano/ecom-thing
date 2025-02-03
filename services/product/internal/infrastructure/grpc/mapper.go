package grpc

import (
	"context"
	"errors"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		{
			return status.Error(codes.InvalidArgument, err.Error())
		}
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
