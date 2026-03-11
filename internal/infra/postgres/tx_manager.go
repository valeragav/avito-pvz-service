package postgres

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

type txKey struct{}

type (
	Transactor interface {
		RunTx(ctx context.Context, opts pgx.TxOptions, fn func(ctx context.Context) error) error

		RunRepeatableRead(ctx context.Context, fn func(ctx context.Context) error) error
		RunReadCommitted(ctx context.Context, fn func(ctx context.Context) error) error

		GetQueryEngine(ctx context.Context) QueryEngine
	}

	Pool interface {
		QueryEngine
		BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	}

	TransactionManager struct {
		pool Pool
	}
)

func NewTransactionManager(pool Pool) *TransactionManager {
	return &TransactionManager{
		pool: pool,
	}
}

func (tm *TransactionManager) RunRepeatableRead(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.RunTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	}, fn)
}

func (tm *TransactionManager) RunReadCommitted(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.RunTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	}, fn)
}

func (tm *TransactionManager) RunTx(ctx context.Context, opts pgx.TxOptions, fn func(ctx context.Context) error) (err error) {
	// If we are already inside the transaction, we will reuse it.
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := tm.pool.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, txKey{}, tx)

	defer func() {
		if p := recover(); p != nil {
			if rErr := tx.Rollback(ctx); rErr != nil {
				logger.ErrorCtx(ctx, "failed to rollback transaction after panic",
					"panic", p,
					"rollback_err", rErr,
				)
			}
			err = fmt.Errorf("task panicked: %v\n%s", p, debug.Stack())
			return
		}

		if err != nil {
			if rErr := tx.Rollback(ctx); rErr != nil {
				logger.ErrorCtx(ctx, "failed to rollback transaction after error",
					"orig_err", err,
					"rollback_err", rErr,
				)
			}
			return
		}

		if cErr := tx.Commit(ctx); cErr != nil {
			err = fmt.Errorf("transaction_manager: commit failed: %w", cErr)
			if rErr := tx.Rollback(ctx); rErr != nil {
				logger.ErrorCtx(ctx, "failed to rollback transaction after commit failure",
					"commit_err", cErr,
					"rollback_err", rErr,
				)
			}
		}
	}()

	return fn(ctx)
}

func (tm *TransactionManager) GetQueryEngine(ctx context.Context) QueryEngine {
	if v := ctx.Value(txKey{}); v != nil {
		if tx, ok := v.(QueryEngine); ok {
			return tx
		}
	}
	return tm.pool
}

type PoolAdapter struct {
	Pool *pgxpool.Pool
}

func (p *PoolAdapter) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return p.Pool.BeginTx(ctx, opts)
}

func (p *PoolAdapter) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.Pool.Exec(ctx, sql, args...)
}

func (p *PoolAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return p.Pool.Query(ctx, sql, args...)
}

func (p *PoolAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.Pool.QueryRow(ctx, sql, args...)
}

func (p *PoolAdapter) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return p.Pool.SendBatch(ctx, b)
}
