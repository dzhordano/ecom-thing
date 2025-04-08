package interceptors

import (
	"context"
	"errors"

	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker[any]
}

func NewCircuitBreaker(cb *gobreaker.CircuitBreaker[any]) *CircuitBreaker {
	return &CircuitBreaker{
		cb: cb,
	}
}

func (c *CircuitBreaker) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
		resp, err := c.cb.Execute(func() (any, error) {
			return handler(ctx, req)
		})

		if err != nil {
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, status.Error(codes.Unavailable, "service unavailable")
			}

			return nil, err
		}

		return resp, nil

	}
}
