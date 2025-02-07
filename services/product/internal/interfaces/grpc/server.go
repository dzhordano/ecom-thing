package grpc

import (
	"github.com/dzhordano/ecom-thing/services/product/internal/interfaces/grpc/interceptors"
	"github.com/sony/gobreaker/v2"
	"log"
	"log/slog"
	"net"
	"time"

	api "github.com/dzhordano/ecom-thing/services/product/pkg/grpc/product/v1"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Option func(*Server)

type Server struct {
	s    *grpc.Server
	addr string

	ratelimiterLimit int
	ratelimiterBurst int

	cb *gobreaker.Settings
}

func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

func WithRateLimiter(limit, burst int) Option {
	return func(s *Server) {
		s.ratelimiterLimit = limit
		s.ratelimiterBurst = burst
	}
}

func WithCircuitBreakerSettings(maxRequests uint32, interval, timeout time.Duration) Option {
	return func(s *Server) {
		s.cb = &gobreaker.Settings{
			Name:        "product-app-cb",
			MaxRequests: maxRequests,
			Interval:    interval,
			Timeout:     timeout,
		}
	}
}

func MustNew(log *slog.Logger, handler api.ProductServiceV1Server, opts ...Option) *Server {
	s := &Server{}

	for _, o := range opts {
		o(s)
	}

	if s.addr == "" {
		panic("addr is required")
	}

	if s.ratelimiterLimit == 0 {
		log.Info("setting ratelimiter limit to default 100")
		s.ratelimiterLimit = 100
	}

	if s.ratelimiterBurst == 0 {
		log.Info("setting ratelimiter burst to default 100")
		s.ratelimiterBurst = 100
	}

	if s.cb == nil {
		log.Info("setting circuit breaker to default values")
		s.cb = &gobreaker.Settings{
			Name:        "product-app-cb",
			MaxRequests: 5,
			Interval:    60 * time.Second,
			Timeout:     5 * time.Second,
		}
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))
			return
		}),
	}

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	cb := gobreaker.NewCircuitBreaker[any](gobreaker.Settings{
		Name:        "product-app-cb",
		MaxRequests: s.cb.MaxRequests,
		Interval:    s.cb.Interval,
		Timeout:     s.cb.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failRatio := float64(counts.TotalFailures) / float64(counts.Requests)

			return failRatio >= 0.50
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Info("circuit breaker state changed",
				slog.String("name", name),
				slog.String("from", from.String()),
				slog.String("to", to.String()),
			)
		},
	})

	ratelimiter := interceptors.NewRateLimiter(s.ratelimiterLimit, s.ratelimiterBurst)

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.NewCircuitBreaker(cb).UnaryServerInterceptor(),
			ratelimiter.RateLimiterInterceptor(),
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(interceptors.InterceptorLogger(log), loggingOpts...),
			interceptors.ErrorMapperInterceptor(),
			interceptors.MetricsInterceptor(),
		),
	)

	api.RegisterProductServiceV1Server(srv, handler)

	reflection.Register(srv)

	return &Server{
		s:    srv,
		addr: s.addr,
	}
}

func (s *Server) Run() error {
	list, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	log.Printf("starting gRPC server on addr %s", s.addr)

	return s.s.Serve(list)
}

func (s *Server) GracefulStop() {
	s.s.GracefulStop()
}
