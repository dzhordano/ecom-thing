package interceptors

import (
	"context"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"
	"log"
)

type RateLimiter struct {
	ratelimit.Limiter
}

func NewRateLimiter(rate int) *RateLimiter {
	return &RateLimiter{
		ratelimit.New(rate),
	}
}

func (rl *RateLimiter) RateLimiterInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
		rl.Take()

		log.Printf("TOOK")

		return handler(ctx, req)
	}
}
