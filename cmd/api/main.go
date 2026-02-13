package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/VaLeraGav/avito-pvz-service/internal/container"
	"github.com/VaLeraGav/avito-pvz-service/internal/http/routes"
	"github.com/VaLeraGav/avito-pvz-service/internal/metrics"
	serviceGrpc "github.com/VaLeraGav/avito-pvz-service/internal/servers/grpc"
	serviceHttp "github.com/VaLeraGav/avito-pvz-service/internal/servers/http"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/VaLeraGav/avito-pvz-service/internal/config"
	"github.com/VaLeraGav/avito-pvz-service/pkg/closer"
	"github.com/VaLeraGav/avito-pvz-service/pkg/dbconnect"
	"github.com/VaLeraGav/avito-pvz-service/pkg/logger"
)

var shutdownTimeout = 5 * time.Second

func main() {
	envFile := flag.String("env", "", "path to config file")
	flag.Parse()

	cfg := config.LoadConfig(*envFile)

	lg := logger.New("avito-pvz-service", cfg.Env, cfg.LogLevel)
	logger.MustSetGlobal(lg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c := closer.New()
	defer shutdown(c, lg)

	metrics.Init()

	connPostgres, err := connectPostgres(cfg)
	if err != nil {
		logger.Error("database connection error", "err", err)
		return
	}
	c.Add(func(ctx context.Context) error {
		lg.Info("shutting down connectPostgres")
		connPostgres.Close()
		return nil
	})

	ct := container.New(cfg, lg, connPostgres)
	err = ct.Init()
	if err != nil {
		logger.Error("container error", "err", err)
		return
	}

	router := routes.NewRouter(ct)

	// Create and start server HTTP
	httpService := newHTTPServer(cfg, router)
	c.Add(func(ctx context.Context) error {
		lg.Info("shutting down server http")
		return httpService.Shutdown(ctx)
	})

	// Create and start server gRPC
	grpcServer, err := newGrpcServer(cfg, []serviceGrpc.RegisterFunc{})
	if err != nil {
		logger.Error("serviceGrpc startup error", "err", err)
		return
	}
	c.Add(func(ctx context.Context) error {
		lg.Info("shutting down service grpc")
		return grpcServer.Shutdown(ctx)
	})

	// Запуск HTTP сервера
	errCh := make(chan error, 2)
	go func() {
		if err := httpService.StartServer(ctx); err != nil {
			errCh <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Запуск gRPC сервера
	go func() {
		if err := grpcServer.StartServer(ctx); err != nil {
			errCh <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Ждем сигнала или ошибки сервера
	select {
	case <-ctx.Done():
		lg.Info("shutdown signal has been received")
	case err := <-errCh:
		lg.Error("server error", "err", err)
	}
}

func newHTTPServer(cfg *config.Config, router http.Handler) *serviceHttp.Server {
	return serviceHttp.NewServer(&http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	})
}

func newGrpcServer(cfg *config.Config, registerFuncs []serviceGrpc.RegisterFunc) (*serviceGrpc.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return serviceGrpc.NewServer(ctx, cfg.GRPC.Address, registerFuncs)
}

func connectPostgres(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := dbconnect.Connect(ctx, dbconnect.PostgresConnectCfg{
		User:     cfg.Db.User,
		Password: cfg.Db.Password,
		Host:     cfg.Db.Host,
		Port:     cfg.Db.Port,
		Dbname:   cfg.Db.NameDb,
		Options:  cfg.Db.Option,
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func shutdown(c *closer.Closer, lg *logger.Logger) {
	closeCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := c.Close(closeCtx); err != nil {
		lg.Error("error when closing resources", "err", err)
	}
}
