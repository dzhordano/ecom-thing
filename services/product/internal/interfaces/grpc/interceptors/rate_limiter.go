package interceptors

import (
	"context"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RateLimiter struct {
	rl *rate.Limiter
}

func NewRateLimiter(limit, burst int) *RateLimiter {
	return &RateLimiter{
		rl: rate.NewLimiter(rate.Limit(limit), burst),
	}
}

func (r *RateLimiter) RateLimiterInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
		if !r.rl.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}
