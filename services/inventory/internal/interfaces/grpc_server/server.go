package grpc_server

import (
	"context"
	"errors"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/infrastructure"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/interfaces/grpc_server/interceptors"
	api "github.com/dzhordano/ecom-thing/services/inventory/pkg/api/inventory/v1"
	"github.com/dzhordano/ecom-thing/services/inventory/pkg/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sony/gobreaker/v2"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"time"
)

const (
	// FailedReq/TotalRequests cap.
	FailRatioCap = 0.65
)

type Option func(*Server)

type Server struct {
	s    *grpc.Server
	addr string

	profilingOn bool

	ratelimiterLimit int
	ratelimiterBurst int

	cb *gobreaker.Settings
	tp *tracesdk.TracerProvider
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

// WithTracing accepts url to send traces to.
// Decides whether to collect traces or not.
func WithTracerProvider(tp *tracesdk.TracerProvider) Option {
	return func(s *Server) {
		s.tp = tp
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

// DefaultServerOptions returns a set of default options for the server.
func DefaultServerOptions() []Option {
	return []Option{
		WithAddr(":50051"),
		WithRateLimiter(100, 100),
		WithCircuitBreakerSettings(5, 60*time.Second, 5*time.Second),
	}
}

func MustNew(log logger.Logger, handler api.InventoryServiceServer, opts ...Option) *Server {
	s := &Server{}

	// Default options
	for _, o := range DefaultServerOptions() {
		o(s)
	}

	// Custom options
	for _, o := range opts {
		o(s)
	}

	if s.addr == "" {
		panic("addr is required")
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("recovered from panic", "panic", p)
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

			return failRatio >= FailRatioCap
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

	sOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptors.NewCircuitBreaker(cb).UnaryServerInterceptor(),
			ratelimiter.RateLimiterInterceptor(),
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(interceptors.InterceptorLogger(log), loggingOpts...),
			interceptors.ErrorMapperInterceptor(),
			interceptors.MetricsInterceptor(),
		),
	}

	if s.tp != nil {
		sOpts = append(sOpts,
			grpc.StatsHandler(
				otelgrpc.NewServerHandler(
					otelgrpc.WithPropagators(propagation.TraceContext{}),
					otelgrpc.WithTracerProvider(s.tp),
				),
			),
		)
	}

	srv := grpc.NewServer(sOpts...)

	api.RegisterInventoryServiceServer(srv, handler)

	grpc_prometheus.Register(srv)

	reflection.Register(srv)

	s.s = srv

	return s
}

// Run starts grpc_server server using cmux.
//
// Handles all HTTP2 requests with 'content-type: application/grpc_server' headers with grpc_server server
// Other paths are hardcoded (for now at least).
//
// Hardcoded ones are: <addr>/metrics. And if profiling is enabled: <addr>/debug/pprof{/,/cmdline,/profile,/symbol,/trace}.
func (s *Server) Run(ctx context.Context) error {
	grpcLis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	gwMux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	}

	if err := api.RegisterInventoryServiceHandlerFromEndpoint(ctx, gwMux, s.addr, dialOpts); err != nil {
		return err
	}

	r := echo.New()

	// Endpoint for getting swagger docs.
	r.GET("/swagger.json", func(c echo.Context) error {
		return c.File("docs/apidocs.swagger.json")
	})

	// Add Swagger UI with /swagger.json a path specified to pull docs.
	// I don't like the way it has doc.json & doc.yaml though.
	r.GET("/swagger/*", echoSwagger.EchoWrapHandler(echoSwagger.URL("/swagger.json")))

	r.Use(
		middleware.Recover(),
	)

	apiGroup := r.Group("/api/v1")
	// Wrap gateway mux.
	apiGroup.Any("/*", echo.WrapHandler(http.StripPrefix("/api/v1", gwMux)))

	r.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	infrastructure.InitMetrics()

	if s.profilingOn {
		r.GET("/debug/pprof/*", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
		r.GET("/debug/pprof/cmdline", echo.WrapHandler(http.HandlerFunc(pprof.Cmdline)))
		r.GET("/debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
		r.GET("/debug/pprof/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
		r.GET("/debug/pprof/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	}

	go func() {
		log.Printf("grpc listening on %s", s.addr)
		if err := s.s.Serve(grpcLis); err != nil {
			log.Printf("grpc serve failed: %v", err)
		}
	}()

	// FIXME аддресс надо не хардкод
	go func() {
		if err := r.Start(":8001"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("http serve failed: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("shutting down servers...")
	s.s.GracefulStop()

	ctxShut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.Shutdown(ctxShut)
}

func (s *Server) GracefulStop() {
	s.s.GracefulStop()
}
