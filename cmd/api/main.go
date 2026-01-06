package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/VaLeraGav/avito-pvz-service/internal/container"
	"github.com/VaLeraGav/avito-pvz-service/internal/http/handlers"
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

	lg := logger.InitLogger("avito-pvz-service", cfg.Env, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c := closer.New()
	defer shutdown(c, lg)

	connPostgres, err := connectPostgres(cfg, lg)
	if err != nil {
		logger.Error("ошибка подключения к базе данных", "err", err)
		return
	}
	c.Add(func(ctx context.Context) error {
		lg.Info("shutting down connectPostgres")
		connPostgres.Close()
		return nil
	})

	container := container.Init(cfg, lg, connPostgres)
	router := handlers.NewRouter(container)

	// Create and start server HTTP
	serviceHttp := newHTTPServer(cfg, router, lg)
	c.Add(func(ctx context.Context) error {
		lg.Info("shutting down server http")
		return serviceHttp.Shutdown(ctx)
	})

	// Create and start server gRPC
	serviceGrpc, err := newGrpcServer(cfg, []serviceGrpc.RegisterFunc{})
	if err != nil {
		logger.Error("ошибка запуска serviceGrpc", "err", err)
		return
	}
	c.Add(func(ctx context.Context) error {
		lg.Info("shutting down service grpc")
		return serviceGrpc.Shutdown(ctx)
	})

	// Запуск HTTP сервера
	errCh := make(chan error, 2)
	go func() {
		if err := serviceHttp.StartServer(ctx); err != nil {
			errCh <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Запуск gRPC сервера
	go func() {
		if err := serviceGrpc.StartServer(ctx); err != nil {
			errCh <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Ждем сигнала или ошибки сервера
	select {
	case <-ctx.Done():
		lg.Info("получен сигнал завершения работы")
	case err := <-errCh:
		lg.Error("ошибка сервера", "err", err)
	}
}

func newHTTPServer(cfg *config.Config, router http.Handler, lg *slog.Logger) *serviceHttp.Server {
	return serviceHttp.NewServer(&http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.HTTPServer.Timeout) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTPServer.Timeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	})
}

// TODO: конфики прокинуть
func newGrpcServer(cfg *config.Config, registerFuncs []serviceGrpc.RegisterFunc) (*serviceGrpc.Server, error) {
	return serviceGrpc.NewServer(":50051", registerFuncs)
}

func connectPostgres(cfg *config.Config, lg *slog.Logger) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := dbconnect.Connect(ctx, dbconnect.PostgresConnectCfg{
		User:     cfg.Db.User,
		Password: cfg.Db.Password,
		Host:     cfg.Db.Host,
		Port:     cfg.Db.ExternalPort,
		Dbname:   cfg.Db.NameDb,
		Options:  cfg.Db.Option,
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func shutdown(c *closer.Closer, lg *slog.Logger) {
	closeCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := c.Close(closeCtx); err != nil {
		lg.Error("ошибка при закрытии ресурсов", "err", err)
	}
}
