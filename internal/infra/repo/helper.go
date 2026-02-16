package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

type builder interface {
	ToSql() (string, []interface{}, error)
}

// CollectRows executes a sql query built by sqb, collects rows into dst using RowMapper.
func CollectRows[T any](ctx context.Context, db *pgxpool.Pool, builder builder, rowMapper func(pgx.CollectableRow) (T, error)) ([]T, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		logger.DebugCtx(ctx, "err builder", "sql", sql, "args", args, "err", err)
		return nil, fmt.Errorf("%w: %w", infra.ErrBuildQuery, err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", infra.ErrExecuteQuery, err)
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, rowMapper)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infra.ErrNotFound
		}
		if IsDuplicateKeyError(err) {
			return nil, infra.ErrDuplicate
		}
		return nil, fmt.Errorf("%w: %w", infra.ErrScanResult, err)
	}

	return results, nil
}

// CollectOneRow executes a sql query built by sqb, collects a single row into dst using RowMapper.
func CollectOneRow[T any](ctx context.Context, db *pgxpool.Pool, builder builder, rowMapper func(pgx.CollectableRow) (T, error)) (T, error) {
	var zero T

	sql, args, err := builder.ToSql()
	if err != nil {
		logger.DebugCtx(ctx, "err builder query", "sql", sql, "args", args, "err", err)
		return zero, fmt.Errorf("%w: %w", infra.ErrBuildQuery, err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		logger.DebugCtx(ctx, "err execute query", "sql", sql, "args", args, "err", err)
		return zero, fmt.Errorf("%w: %w", infra.ErrExecuteQuery, err)
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, rowMapper)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return zero, infra.ErrNotFound
		}
		if IsDuplicateKeyError(err) {
			return zero, infra.ErrDuplicate
		}
		logger.DebugCtx(ctx, "err scan", "sql", sql, "args", args, "err", err)
		return zero, fmt.Errorf("%w: %w", infra.ErrScanResult, err)
	}

	return result, nil
}

func IsDuplicateKeyError(err error) bool {
	return IsPgErrorWithCode(err, pgerrcode.UniqueViolation)
}

func IsPgErrorWithCode(err error, code string) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.SQLState() == code
}
