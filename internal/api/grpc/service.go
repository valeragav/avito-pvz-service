package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/valeragav/avito-pvz-service/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// функция, которая регистрирует сервис на gRPC сервере
type RegisterFunc func(*grpc.Server)

type Server struct {
	name       string
	grpcServer *grpc.Server
	listener   net.Listener
	addr       string
}

func NewServer(ctx context.Context, name, addr string, registerFuncs []RegisterFunc, opts ...grpc.ServerOption) (*Server, error) {
	const op = "grpc.NewServer"

	lc := net.ListenConfig{}
	lis, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("%s: listen %s: %w", op, addr, err)
	}

	s := grpc.NewServer(opts...)

	for _, register := range registerFuncs {
		register(s)
	}

	reflection.Register(s)

	return &Server{
		name:       name,
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

	logger.Info("listening", "nameServer", s.name, "addr", s.listener.Addr().String())

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

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}
