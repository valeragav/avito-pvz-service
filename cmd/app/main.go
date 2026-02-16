package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/api"
	"github.com/valeragav/avito-pvz-service/internal/config"
	"github.com/valeragav/avito-pvz-service/internal/container"
	"github.com/valeragav/avito-pvz-service/internal/metrics"
	"github.com/valeragav/avito-pvz-service/pkg/closer"
	"github.com/valeragav/avito-pvz-service/pkg/dbconnect"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

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

	connPostgres, err := connectPostgres(cfg, c)
	if err != nil {
		logger.Error("database connection error", "err", err)
		return
	}

	ctn := container.New(cfg, lg, connPostgres)
	err = ctn.Init()
	if err != nil {
		logger.Error("container error", "err", err)
		return
	}

	api.NewApi(ctx, c, cfg, ctn)
}

func connectPostgres(cfg *config.Config, c *closer.Closer) (*pgxpool.Pool, error) {
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
	c.Add(func(ctx context.Context) error {
		logger.Info("shutting down connectPostgres")
		conn.Close()
		return nil
	})

	return conn, nil
}

func shutdown(c *closer.Closer, lg *logger.Logger) {
	var shutdownTimeout = 5 * time.Second

	closeCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := c.Close(closeCtx); err != nil {
		lg.Error("error when closing resources", "err", err)
	}
}
