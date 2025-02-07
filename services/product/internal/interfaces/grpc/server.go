package grpc

import (
	"github.com/dzhordano/ecom-thing/services/product/internal/interfaces/grpc/interceptors"
	"log"
	"log/slog"
	"net"

	api "github.com/dzhordano/ecom-thing/services/product/pkg/grpc/product/v1"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	s    *grpc.Server
	addr string
}

func MustNew(log *slog.Logger, addr string, rps uint16, handler api.ProductServiceV1Server) *Server {
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

	ratelimiter := interceptors.NewRateLimiter(rps)

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			ratelimiter.RateLimiterInterceptor(),
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(interceptors.InterceptorLogger(log), loggingOpts...),
			interceptors.ErrorMapperInterceptor(),
			interceptors.MetricsInterceptor(),
		),
	)

	api.RegisterProductServiceV1Server(s, handler)

	reflection.Register(s)

	return &Server{
		s:    s,
		addr: addr,
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
