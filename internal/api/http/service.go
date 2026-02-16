package http

import (
	"context"
	"fmt"
	nethttp "net/http"

	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

type Server struct {
	server *nethttp.Server
}

func NewServer(server *nethttp.Server) *Server {
	return &Server{
		server: server,
	}
}

func (s *Server) StartServer(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		err := s.server.ListenAndServe()
		if err != nil && err != nethttp.ErrServerClosed {
			errCh <- err
			return
		}
	}()

	logger.Info("listening http", "addr", s.server.Addr)

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}
	return nil
}
