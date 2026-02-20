package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	postgresDB "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/config"
	"github.com/valeragav/avito-pvz-service/internal/infra/postgres"
	"github.com/valeragav/avito-pvz-service/internal/seed"
	"github.com/valeragav/avito-pvz-service/migrations"
	"github.com/valeragav/avito-pvz-service/pkg/dbconnect"
	"github.com/valeragav/avito-pvz-service/pkg/seeder"
)

var testApp *TestApp

func TestMain(m *testing.M) {
	code := run(m)
	os.Exit(code)
}

func run(m *testing.M) int {
	app, err := NewTestApp()
	if err != nil {
		log.Printf("failed to init test app: %v", err)
		return 1
	}

	testApp = app

	code := m.Run()

	// очистка после тестов
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := testApp.Cleanup(ctx); err != nil {
		log.Printf("failed to cleanup db: %v", err)
	}

	app.DB.Close()

	return code
}

type TestApp struct {
	DB *pgxpool.Pool
}

func NewTestApp() (*TestApp, error) {
	db, err := connectTestDB()
	if err != nil {
		return nil, err
	}

	app := &TestApp{DB: db}

	if err := app.Migrate(); err != nil {
		return nil, err
	}

	// Предполагается что seed будем запускать только там где нужно
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// if err := app.Seed(ctx); err != nil {
	// 	return nil, err
	// }

	return app, nil
}

func connectTestDB() (*pgxpool.Pool, error) {
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

type SeedTarget int

const (
	SeedCities SeedTarget = iota
	SeedReceptionStatuses
	SeedProductTypes
)

func (a TestApp) Seed(ctx context.Context, db postgres.DBTX, targets ...SeedTarget) error {
	if len(targets) == 0 {
		// по умолчанию сидим всё
		targets = []SeedTarget{SeedCities, SeedReceptionStatuses, SeedProductTypes}
	}

	sd := seeder.New()

	for _, t := range targets {
		switch t {
		case SeedCities:
			citiesRepo := postgres.NewCityRepository(db)
			sd.Add(seed.NewCitySeed(citiesRepo))
		case SeedReceptionStatuses:
			statusesRepo := postgres.NewReceptionStatusRepository(db)
			sd.Add(seed.NewReceptionStatusSeed(statusesRepo))
		case SeedProductTypes:
			productTypesRepo := postgres.NewProductTypeRepository(db)
			sd.Add(seed.NewProductTypeSeed(productTypesRepo))
		}
	}

	return sd.Run(ctx)
}

func (a TestApp) Migrate() error {
	sqlDB := stdlib.OpenDBFromPool(a.DB)

	driver, err := postgresDB.WithInstance(sqlDB, &postgresDB.Config{})
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

func (a *TestApp) Cleanup(ctx context.Context) error {
	rows, err := a.DB.Query(ctx, `
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
		// TODO: так делать не надо
		if t == "schema_migrations" {
			continue
		}
		tables = append(tables, t)
	}

	for _, table := range tables {
		if _, err := a.DB.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table)); err != nil {
			return err
		}
	}

	return nil
}

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

func WithTx(t *testing.T, fn func(ctx context.Context, tx postgres.DBTX)) {
	t.Helper()

	ctx := context.Background()

	tx, err := testApp.DB.Begin(ctx)
	require.NoError(t, err)

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	fn(ctx, tx)
}
