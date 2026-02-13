package products

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (r Repository) Create(ctx context.Context, product Products) (*Products, error) {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	queryBuilder := r.sqb.
		Insert(product.TableName()).
		Columns(Cols.ID, Cols.DateTime, Cols.TypeIs, Cols.ReceptionID).
		Values(product.ID, product.DateTime, product.TypeIs, product.ReceptionID).
		Suffix("RETURNING *")

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	productCreate, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Products])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &productCreate, nil
}

func (r *Repository) GetLastProductInReception(ctx context.Context, receptionID uuid.UUID) (*Products, error) {
	selectBuilder := r.sqb.
		Select((Products{}).AllCols()...).
		From((Products{}).TableName()).
		Where(sq.Eq{Cols.ReceptionID: receptionID}).
		OrderBy(fmt.Sprintf("%s %s", Cols.DateTime, "DESC")).
		Limit(1)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Products])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &result, nil
}

func (r *Repository) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	deleteBuilder := r.sqb.
		Delete((Products{}).TableName()).
		Where(sq.Eq{Cols.ID: productID})

	sql, args, err := deleteBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}

	if tag.RowsAffected() == 0 {
		return storage.ErrNotFound
	}

	return nil
}

func (r *Repository) ListByReceptionIDsWithTypeName(ctx context.Context, receptionIDs []uuid.UUID) ([]ProductsWithTypeName, error) {
	selectBuilder := r.sqb.Select(
		"p.id as id",
		"p.date_time as date_time",
		"p.type_id as type_id",
		"p.reception_id as reception_id",
		"pt.name AS type_name",
	).
		From("products p").
		Join("product_types pt ON pt.id = p.type_id").
		Where(sq.Eq{"p.reception_id": receptionIDs})

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[ProductsWithTypeName])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return results, nil
}
