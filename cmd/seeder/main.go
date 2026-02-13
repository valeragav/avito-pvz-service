package main

import (
	"context"
	"flag"
	"time"

	"github.com/VaLeraGav/avito-pvz-service/internal/config"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/cities"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/product_types"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/statuses"
	"github.com/VaLeraGav/avito-pvz-service/internal/seed"
	"github.com/VaLeraGav/avito-pvz-service/pkg/dbconnect"
	"github.com/VaLeraGav/avito-pvz-service/pkg/logger"
	"github.com/VaLeraGav/avito-pvz-service/pkg/seeder"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	envFile := flag.String("env", "", "path to config file")
	flag.Parse()

	cfg := config.LoadConfig(*envFile)

	lg := logger.New("seeder", cfg.Env, cfg.LogLevel)
	logger.MustSetGlobal(lg)

	connPostgres, err := connectPostgres(cfg)
	if err != nil {
		lg.Error("database connection error", "err", err)
		return
	}

	citiesRepo := cities.New(connPostgres)
	statusesRepo := statuses.New(connPostgres)
	productTypesRepo := product_types.New(connPostgres)

	sd := seeder.New()
	sd.Add(seed.NewCitiesSeed(citiesRepo))
	sd.Add(seed.NewStatusesSeed(statusesRepo))
	sd.Add(seed.NewProductTypesSeed(productTypesRepo))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sd.Run(ctx)
	if err != nil {
		lg.Error("felid to run seeds", "err", err)
		return
	}

	lg.Info("seeding finished successfully")
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
