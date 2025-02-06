package grpc

import (
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// HashMap for domain errors for efficient mapping.
	errorMap = map[error]codes.Code{
		domain.ErrInvalidArgument:      codes.InvalidArgument,
		domain.ErrProductNotFound:      codes.NotFound,
		domain.ErrProductAlreadyExists: codes.AlreadyExists,
	}
)

func MapError(err error) error {
	if s, ok := status.FromError(err); ok {
		return s.Err() // Return the status error if it's a gRPC error
	}

	if code, ok := errorMap[err]; ok {
		return status.Error(code, err.Error())
	}

	return status.Error(codes.Internal, err.Error())
}
