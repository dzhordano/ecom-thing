package grpc_server

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/infrastructure"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/interfaces/grpc_server/interceptors"
	api "github.com/dzhordano/ecom-thing/services/inventory/pkg/api/inventory/v1"
	"github.com/dzhordano/ecom-thing/services/inventory/pkg/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/soheilhy/cmux"
	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Option func(*Server)

type Server struct {
	s    *grpc.Server
	addr string

	profilingOn bool

	ratelimiterLimit int
	ratelimiterBurst int

	cb *gobreaker.Settings
}

func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

func WithProfiling() Option {
	return func(s *Server) {
		s.profilingOn = true
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
			Name:        "inventory-app-cb",
			MaxRequests: maxRequests,
			Interval:    interval,
			Timeout:     timeout,
		}
	}
}

func MustNew(log logger.Logger, handler api.InventoryServiceServer, opts ...Option) *Server {
	s := &Server{
		profilingOn:      false,
		ratelimiterLimit: 100, // TODO магические числа
		ratelimiterBurst: 100,
		cb: &gobreaker.Settings{
			Name:        "inventory-app-cb",
			MaxRequests: 5,
			Interval:    60 * time.Second,
			Timeout:     5 * time.Second,
		},
	}

	for _, o := range opts {
		o(s)
	}

	if s.addr == "" {
		panic("addr is required")
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", "panic", p)
			return
		}),
	}

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	cb := gobreaker.NewCircuitBreaker[any](gobreaker.Settings{
		Name:        "inventory-app-cb",
		MaxRequests: s.cb.MaxRequests,
		Interval:    s.cb.Interval,
		Timeout:     s.cb.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failRatio := float64(counts.TotalFailures) / float64(counts.Requests)

			return failRatio >= 0.50 // TODO маг число
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Info("circuit breaker state changed",
				"name", name,
				"from", from.String(),
				"to", to.String(),
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

	api.RegisterInventoryServiceServer(srv, handler)

	reflection.Register(srv)

	s.s = srv

	return s
}

// Run starts grpc server using cmux.
//
// Handles all HTTP2 requests with 'content-type: application/grpc' headers with grpc server
// Other paths are hardcoded (for now at least).
//
// Hardcoded ones are: <addr>/metrics. And if profiling is enabled: <addr>/debug/pprof{/,/cmdline,/profile,/symbol,/trace}.
func (s *Server) Run(_ context.Context) error {
	list, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	m := cmux.New(list)

	grpcL := m.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
	)

	httpL := m.Match(cmux.Any())

	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())
	infrastructure.InitMetrics()

	// Turn pprof server ON if flag 'profilingOn' is set.
	if s.profilingOn {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	// Создаём HTTP сервер для метрик.
	httpServer := &http.Server{
		Handler: mux,
	}

	// gRPC сервер в отдельной горутине.
	go func() {
		if err := s.s.Serve(grpcL); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// HTTP сервер в отдельной горутине.
	go func() {
		if err := httpServer.Serve(httpL); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to serve HTTP: %v", err)
		}
	}()

	log.Printf("starting grpc and http server on addr: %s", s.addr)

	return m.Serve()
}

func (s *Server) GracefulStop() {
	s.s.GracefulStop()
}
