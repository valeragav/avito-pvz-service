package statuses

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage"
)

type Repository struct {
	db  *pgxpool.Pool
	sqb sq.StatementBuilderType
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		db:  db,
		sqb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *Repository) Get(ctx context.Context, filter Statuses) (*Statuses, error) {
	where := sq.Eq{}
	if filter.ID != uuid.Nil {
		where[StatusCols.ID] = filter.ID
	}
	if filter.Name != "" {
		where[StatusCols.Name] = filter.Name
	}

	selectBuilder := r.sqb.
		Select(filter.AllCols()...).
		From(filter.TableName()).
		Where(where)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Statuses])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &result, nil
}

func (r Repository) CreateBatch(ctx context.Context, statuses []Statuses) ([]Statuses, error) {
	if len(statuses) == 0 {
		return nil, nil
	}

	qb := r.sqb.
		Insert(Statuses{}.TableName()).
		Columns(
			StatusCols.ID,
			StatusCols.Name,
		)

	for _, status := range statuses {
		if status.ID == uuid.Nil {
			status.ID = uuid.New()
		}

		qb = qb.Values(status.ID, status.Name)
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	createdStatuses, err := pgx.CollectRows(rows, pgx.RowToStructByName[Statuses])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return createdStatuses, nil
}
