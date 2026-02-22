package main

import (
	"context"
	"flag"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/config"
	"github.com/valeragav/avito-pvz-service/internal/infra/postgres"

	"github.com/valeragav/avito-pvz-service/internal/seed"
	"github.com/valeragav/avito-pvz-service/pkg/dbconnect"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
	"github.com/valeragav/avito-pvz-service/pkg/seeder"
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

	citiesRepo := postgres.NewCityRepository(connPostgres)
	statusesRepo := postgres.NewReceptionStatusRepository(connPostgres)
	productTypesRepo := postgres.NewProductTypeRepository(connPostgres)

	sd := seeder.New()
	sd.Add(seed.NewCitySeed(citiesRepo))
	sd.Add(seed.NewReceptionStatusSeed(statusesRepo))
	sd.Add(seed.NewProductTypeSeed(productTypesRepo))

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
