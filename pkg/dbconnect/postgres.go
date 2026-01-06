package dbconnect

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConnectCfg struct {
	User, Password, Host, Port, Dbname, Options string
}

func Connect(ctx context.Context, cnt PostgresConnectCfg) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cnt.User, cnt.Password, cnt.Host, cnt.Port, cnt.Dbname,
	)

	if cnt.Options != "" {
		dsn += "?" + cnt.Options
	}

	fmt.Println(dsn)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
