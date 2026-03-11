package postgres_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := testutils.Cleanup(ctx, testApp.DB); err != nil {
		log.Printf("failed to cleanup db: %v", err)
	}

	app.DB.Close()

	return code
}

type TestApp struct {
	DB *pgxpool.Pool
	TM *postgres.TransactionManager
}

func NewTestApp() (*TestApp, error) {
	connPostgres, err := testutils.ConnectTestDB()
	if err != nil {
		return nil, err
	}

	appTest := &TestApp{
		DB: connPostgres,
		TM: postgres.NewTransactionManager(&postgres.PoolAdapter{Pool: connPostgres}),
	}

	if err := testutils.Migrate(connPostgres); err != nil {
		return nil, err
	}

	return appTest, nil
}

func WithTx(t *testing.T, fn func(ctx context.Context, qeProvider postgres.QueryEngineProvider)) {
	t.Helper()
	ctx := context.Background()

	err := testApp.TM.RunTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	}, func(ctx context.Context) error {
		fn(ctx, testApp.TM)
		return ErrRollbackTest // always return an error to roll back the transaction.
	})

	// expect our error to be rolled back
	if !errors.Is(err, ErrRollbackTest) {
		require.NoError(t, err)
	}
}

var ErrRollbackTest = errors.New("test rollback")
