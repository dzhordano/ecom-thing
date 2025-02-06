package grpc

import (
	api "github.com/dzhordano/ecom-thing/services/product/pkg/grpc/product/v1"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"log/slog"
	"net"
)

type Server struct {
	s    *grpc.Server
	addr string
}

func MustNew(log *slog.Logger, addr string, handler api.ProductServiceV1Server) *Server {
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))

			panic(err)
		}),
	}

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
			MetricsInterceptor(),
			ErrorMapperInterceptor(),
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
