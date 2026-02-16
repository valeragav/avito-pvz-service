package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/valeragav/avito-pvz-service/internal/config"
	"github.com/valeragav/avito-pvz-service/internal/container"
	"github.com/valeragav/avito-pvz-service/pkg/closer"
	"github.com/valeragav/avito-pvz-service/pkg/logger"

	serviceGrpc "github.com/valeragav/avito-pvz-service/internal/api/grpc"
	serviceHttp "github.com/valeragav/avito-pvz-service/internal/api/http"
)

func NewApi(ctx context.Context, c *closer.Closer, cfg *config.Config, ctn *container.DIContainer) {
	router := serviceHttp.NewRouter(ctn)

	httpService := newHTTPServer(cfg, c, router)

	grpcServer, err := newGrpcServer(cfg, c, []serviceGrpc.RegisterFunc{})
	if err != nil {
		logger.Error("serviceGrpc startup error", "err", err)
		return
	}

	errCh := make(chan error, 2)
	go func() {
		if err := httpService.StartServer(ctx); err != nil {
			errCh <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	go func() {
		if err := grpcServer.StartServer(ctx); err != nil {
			errCh <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal has been received")
	case err := <-errCh:
		logger.Error("server error", "err", err)
	}
}

func newHTTPServer(cfg *config.Config, c *closer.Closer, router http.Handler) *serviceHttp.Server {
	service := serviceHttp.NewServer(&http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	})

	c.Add(func(ctx context.Context) error {
		logger.Info("shutting down server http")
		return service.Shutdown(ctx)
	})

	return service
}

func newGrpcServer(cfg *config.Config, c *closer.Closer, registerFuncs []serviceGrpc.RegisterFunc) (*serviceGrpc.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	service, err := serviceGrpc.NewServer(ctx, cfg.GRPC.Address, registerFuncs)
	if err != nil {
		return nil, err
	}
	c.Add(func(ctx context.Context) error {
		logger.Info("shutting down service grpc")
		return service.Shutdown(ctx)
	})

	return service, nil
}
