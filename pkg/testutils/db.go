package testutils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/valeragav/avito-pvz-service/internal/config"
	"github.com/valeragav/avito-pvz-service/migrations"
	"github.com/valeragav/avito-pvz-service/pkg/dbconnect"
)

func ConnectTestDB() (*pgxpool.Pool, error) {
	cfg := loadTestConfig()

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

// TODO: вытащить в conf
func loadTestConfig() *config.Config {
	return &config.Config{
		Db: config.Db{
			User:     "root",
			Password: "root",
			Host:     "localhost",
			Port:     "5439",
			NameDb:   "pvz-service_db",
			Option:   "sslmode=disable",
		},
	}
}

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func Cleanup(ctx context.Context, db querier) error {
	rows, err := db.Query(ctx, `
SELECT table_name
FROM information_schema.tables
WHERE table_schema='public' AND table_type='BASE TABLE';
`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return err
		}
		if t == "schema_migrations" {
			continue
		}
		tables = append(tables, t)
	}

	for _, table := range tables {
		if _, err := db.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table)); err != nil {
			return err
		}
	}

	return nil
}

func Migrate(dbPool *pgxpool.Pool) error {
	sqlDB := stdlib.OpenDBFromPool(dbPool)

	driver, err := postgresMigrate.WithInstance(sqlDB, &postgresMigrate.Config{})
	if err != nil {
		return err
	}

	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	srcErr, dbErr := m.Close()

	var combinedErr error
	if srcErr != nil {
		combinedErr = fmt.Errorf("%w; migrate source close error: %w", combinedErr, srcErr)
	}
	if dbErr != nil {
		combinedErr = fmt.Errorf("%w; migrate database close error: %w", combinedErr, dbErr)
	}

	return combinedErr
}
