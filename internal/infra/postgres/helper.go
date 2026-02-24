package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

type builder interface {
	ToSql() (string, []any, error)
}

func Exec(ctx context.Context, db DBTX, builder builder) error {
	sql, args, err := builder.ToSql()
	if err != nil {
		logger.DebugCtx(ctx, "err builder", "sql", sql, "args", args, "err", err)
		return fmt.Errorf("%w: %w", ErrBuildQuery, err)
	}
	_, err = db.Exec(ctx, sql, args...)
	return err
}

// CollectRows executes a sql query built by sqb, collects rows into dst using RowMapper.
func CollectRows[T any](ctx context.Context, db DBTX, builder builder, rowMapper func(pgx.CollectableRow) (T, error)) ([]T, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		logger.DebugCtx(ctx, "err builder", "sql", sql, "args", args, "err", err)
		return nil, fmt.Errorf("%w: %w", ErrBuildQuery, err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		logger.DebugCtx(ctx, "err execute query", "sql", sql, "args", args, "err", err)
		return nil, fmt.Errorf("%w: %w", ErrExecuteQuery, err)
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, rowMapper)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return nil, infra.ErrDuplicate
		}
		logger.DebugCtx(ctx, "err scan", "sql", sql, "args", args, "err", err)
		return nil, fmt.Errorf("%w: %w", ErrScanResult, err)
	}

	return results, nil
}

// CollectOneRow executes a sql query built by sqb, collects a single row into dst using RowMapper.
func CollectOneRow[T any](ctx context.Context, db DBTX, builder builder, rowMapper func(pgx.CollectableRow) (T, error)) (T, error) {
	var zero T

	sql, args, err := builder.ToSql()
	if err != nil {
		logger.DebugCtx(ctx, "err builder query", "sql", sql, "args", args, "err", err)
		return zero, fmt.Errorf("%w: %w", ErrBuildQuery, err)
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		logger.DebugCtx(ctx, "err execute query", "sql", sql, "args", args, "err", err)
		return zero, fmt.Errorf("%w: %w", ErrExecuteQuery, err)
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
		return zero, fmt.Errorf("%w: %w", ErrScanResult, err)
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
	if errors.As(err, &pgErr) {
		return pgErr.Code == code
	}

	return false
}
