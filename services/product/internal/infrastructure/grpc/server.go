package grpc

import (
	"log"
	"log/slog"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Server struct {
	s    *grpc.Server
	port string
}

func NewServer(log *slog.Logger, port string) *Server {
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			// logging.StartCall, logging.FinishCall,
			logging.PayloadReceived, logging.PayloadSent,
		),
		// Add any other option (check functions starting with logging.With).
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
			ErrorMapperInterceptor(),
		),
	)

	reflection.Register(s)

	return &Server{
		s:    s,
		port: ":" + port,
	}
}

func (s *Server) Run() error {
	list, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}

	log.Printf("starting gRPC server on port %s", s.port)

	return s.s.Serve(list)
}

func (s *Server) GracefulStop() {
	s.s.GracefulStop()
}
