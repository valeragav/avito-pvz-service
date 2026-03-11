//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	serviceHttp "github.com/valeragav/avito-pvz-service/internal/api/http"
	"github.com/valeragav/avito-pvz-service/internal/app"
	"github.com/valeragav/avito-pvz-service/internal/config"
	"github.com/valeragav/avito-pvz-service/internal/infra/postgres"
	"github.com/valeragav/avito-pvz-service/pkg/testutils"
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
	if err := testutils.Cleanup(ctx, testApp.DB); err != nil {
		log.Printf("failed to cleanup db: %v", err)
	}

	app.DB.Close()

	return code
}

type TestApp struct {
	DB     *pgxpool.Pool
	TM     *postgres.TransactionManager
	Server *httptest.Server
}

func NewTestApp() (*TestApp, error) {
	connPostgres, err := testutils.ConnectTestDB()
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	if err := testutils.Migrate(connPostgres); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	testutils.InitTestLogger()
	cfg, err := buildTestConfig()
	if err != nil {
		return nil, fmt.Errorf("build config: %w", err)
	}

	appService, err := app.New(cfg, connPostgres)
	if err != nil {
		return nil, fmt.Errorf("create app: %w", err)
	}

	tm := postgres.NewTransactionManager(&postgres.PoolAdapter{Pool: connPostgres})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := testutils.Seed(ctx, tm); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}

	return &TestApp{
		DB:     connPostgres,
		TM:     tm,
		Server: httptest.NewServer(serviceHttp.NewRouter(appService, cfg)),
	}, nil
}

func buildTestConfig() (*config.Config, error) {
	cfg := config.LoadConfig("")

	privatePath, publicPath, err := testutils.GenerateTestJWTKeys()
	if err != nil {
		return nil, fmt.Errorf("generate jwt keys: %w", err)
	}

	cfg.Jwt.RSAPrivateFile = privatePath
	cfg.Jwt.RSAPublicFile = publicPath

	return cfg, nil
}
