package product_types

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

func (r *Repository) Get(ctx context.Context, filter ProductTypes) (*ProductTypes, error) {
	where := sq.Eq{}
	if filter.ID != uuid.Nil {
		where[productTypeCols.ID] = filter.ID
	}
	if filter.Name != "" {
		where[productTypeCols.Name] = filter.Name
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

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[ProductTypes])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &result, nil
}

func (r Repository) CreateBatch(ctx context.Context, productTypes []ProductTypes) ([]ProductTypes, error) {
	qb := r.sqb.
		Insert(ProductTypes{}.TableName()).
		Columns(
			productTypeCols.ID,
			productTypeCols.Name,
		)

	for _, productType := range productTypes {
		if productType.ID == uuid.Nil {
			productType.ID = uuid.New()
		}

		qb = qb.Values(productType.ID, productType.Name)
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

	createdProductTypes, err := pgx.CollectRows(rows, pgx.RowToStructByName[ProductTypes])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return createdProductTypes, nil
}
