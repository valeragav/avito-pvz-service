package grpc

import (
	"context"
	"net"

	"github.com/VaLeraGav/avito-pvz-service/pkg/logger"
	"google.golang.org/grpc"
)

// RegisterFunc — функция, которая регистрирует сервис на gRPC сервере
type RegisterFunc func(*grpc.Server)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	addr       string
}

func NewServer(addr string, registerFuncs []RegisterFunc, opts ...grpc.ServerOption) (*Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(opts...)

	for _, register := range registerFuncs {
		register(s)
	}

	return &Server{
		grpcServer: s,
		listener:   lis,
		addr:       addr,
	}, nil
}

func (s *Server) StartServer(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		err := s.grpcServer.Serve(s.listener)
		if err != nil {
			errCh <- err
			return
		}
	}()

	logger.Info("listening grpc", "addr", s.listener.Addr().String())

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.grpcServer.GracefulStop()
	return nil
}
