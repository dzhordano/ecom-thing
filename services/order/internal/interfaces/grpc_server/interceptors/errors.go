package interceptors

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// HashMap for domain errors for efficient mapping.
	// Не знаю, можно ли улучшить. Над подумац.
	errorMap = map[error]codes.Code{
		domain.ErrOrderNotFound:         codes.NotFound,
		domain.ErrInvalidOrderStatus:    codes.InvalidArgument,
		domain.ErrInvalidCurrency:       codes.InvalidArgument,
		domain.ErrInvalidPaymentMethod:  codes.InvalidArgument,
		domain.ErrInvalidDeliveryMethod: codes.InvalidArgument,
		domain.ErrInvalidArgument:       codes.InvalidArgument,
		domain.ErrInvalidDescription:    codes.InvalidArgument,

		domain.ErrInternal:               codes.Internal,
		domain.ErrInvalidUUID:            codes.InvalidArgument,
		domain.ErrInvalidPrice:           codes.InvalidArgument,
		domain.ErrInvalidDiscount:        codes.InvalidArgument,
		domain.ErrInvalidDeliveryAddress: codes.InvalidArgument,
		domain.ErrInvalidDeliveryDate:    codes.InvalidArgument,
		domain.ErrInvalidOrderItems:      codes.InvalidArgument,

		domain.ErrOrderAlreadyCompleted: codes.InvalidArgument,
		domain.ErrOrderAlreadyCancelled: codes.InvalidArgument,

		domain.ErrCouponExpired:   codes.InvalidArgument,
		domain.ErrCouponNotFound:  codes.InvalidArgument,
		domain.ErrCouponNotActive: codes.InvalidArgument,
	}
)

func mapError(err error) error {
	if s, ok := status.FromError(err); ok {
		return s.Err() // Return the status error if it's a gRPC error
	}

	if code, ok := errorMap[err]; ok {
		return status.Error(code, err.Error())
	}

	return status.Error(codes.Internal, err.Error())
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
