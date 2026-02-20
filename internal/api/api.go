package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	serviceGrpc "github.com/valeragav/avito-pvz-service/internal/api/grpc"
	serviceHttp "github.com/valeragav/avito-pvz-service/internal/api/http"
	"github.com/valeragav/avito-pvz-service/internal/app"
	"github.com/valeragav/avito-pvz-service/internal/config"
	"github.com/valeragav/avito-pvz-service/pkg/closer"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

func NewApi(ctx context.Context, c *closer.Closer, cfg *config.Config, app *app.App) {
	errCh := make(chan error, 4)

	runServer := func(name string, start func(context.Context) error, enabled bool) {
		if !enabled {
			return
		}
		go func() {
			if err := start(ctx); err != nil {
				errCh <- fmt.Errorf("%s server error: %w", name, err)
			}
		}()
	}

	gRPCService := "gRPC"
	runServer(gRPCService, func(ctx context.Context) error {
		registers := serviceGrpc.CollectRegisters(app)
		grpcServer, err := newGrpcServer(cfg, gRPCService, c, registers)
		if err != nil {
			return fmt.Errorf("failed to create gRPC server: %w", err)
		}
		return grpcServer.StartServer(ctx)
	}, true)

	httpNameService := "HTTP"
	runServer(httpNameService, func(ctx context.Context) error {
		router := serviceHttp.NewRouter(app)
		httpService := newHTTPServer(cfg, httpNameService, c, router)
		return httpService.StartServer(ctx)
	}, true)

	metricsNameService := "Metrics"
	runServer(metricsNameService, func(ctx context.Context) error {
		metricsRoute := serviceHttp.NewMetricsRoute()
		metricsService := newMetricsServer(cfg, metricsNameService, c, metricsRoute)
		return metricsService.StartServer(ctx)
	}, true)

	swaggerNameService := "Swagger"
	runServer(swaggerNameService, func(ctx context.Context) error {
		swaggerRoute := serviceHttp.NewSwaggerRoute()
		swaggerService := newSwaggerServer(cfg, swaggerNameService, c, swaggerRoute)
		return swaggerService.StartServer(ctx)
	}, cfg.SwaggerServer.Enabled)

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal has been received")
	case err := <-errCh:
		logger.Error("server error", "err", err)
	}
}

func newHTTPServer(cfg *config.Config, name string, c *closer.Closer, router http.Handler) *serviceHttp.Server {
	service := serviceHttp.NewServer(name, &http.Server{
		Handler:           router,
		Addr:              cfg.HTTPServer.Address,
		ReadTimeout:       cfg.HTTPServer.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTPServer.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPServer.WriteTimeout,
		IdleTimeout:       cfg.HTTPServer.IdleTimeout,
	})

	c.Add(func(ctx context.Context) error {
		logger.Info("shutting down server http")
		return service.Shutdown(ctx)
	})

	return service
}

func newMetricsServer(cfg *config.Config, name string, c *closer.Closer, router http.Handler) *serviceHttp.Server {
	service := serviceHttp.NewServer(name, &http.Server{
		Handler:      router,
		Addr:         cfg.MetricsServer.Address,
		ReadTimeout:  cfg.MetricsServer.ReadTimeout,
		WriteTimeout: cfg.MetricsServer.WriteTimeout,
		IdleTimeout:  cfg.MetricsServer.IdleTimeout,
	})

	c.Add(func(ctx context.Context) error {
		logger.Info("shutting down server metrics")
		return service.Shutdown(ctx)
	})

	return service
}

func newSwaggerServer(cfg *config.Config, name string, c *closer.Closer, router http.Handler) *serviceHttp.Server {
	service := serviceHttp.NewServer(name, &http.Server{
		Handler:      router,
		Addr:         cfg.SwaggerServer.Address,
		ReadTimeout:  cfg.SwaggerServer.ReadTimeout,
		WriteTimeout: cfg.SwaggerServer.WriteTimeout,
		IdleTimeout:  cfg.SwaggerServer.IdleTimeout,
	})

	c.Add(func(ctx context.Context) error {
		logger.Info("shutting down server swagger")
		return service.Shutdown(ctx)
	})

	return service
}

func newGrpcServer(cfg *config.Config, name string, c *closer.Closer, registerFuncs []serviceGrpc.RegisterFunc) (*serviceGrpc.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	service, err := serviceGrpc.NewServer(ctx, name, cfg.GRPC.Address, registerFuncs)
	if err != nil {
		return nil, err
	}
	c.Add(func(ctx context.Context) error {
		logger.Info("shutting down service grpc")
		return service.Shutdown(ctx)
	})

	return service, nil
}
