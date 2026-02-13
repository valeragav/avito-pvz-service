package cities

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

func (r *Repository) Get(ctx context.Context, filter Cities) (*Cities, error) {
	where := sq.Eq{}
	if filter.ID != uuid.Nil {
		where[cityCols.ID] = filter.ID
	}
	if filter.Name != "" {
		where[cityCols.Name] = filter.Name
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

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Cities])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &result, nil
}

// use it if there are many records
func (r Repository) CreateBatchPgx(ctx context.Context, cities []Cities) error {
	batch := &pgx.Batch{}

	for _, city := range cities {
		if city.ID == uuid.Nil {
			city.ID = uuid.New()
		}

		sql := fmt.Sprintf(
			"INSERT INTO %s (%s, %s) VALUES ($1, $2)",
			Cities{}.TableName(),
			cityCols.ID,
			cityCols.Name,
		)

		batch.Queue(sql, city.ID, city.Name)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for range cities {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r Repository) CreateBatch(ctx context.Context, cities []Cities) ([]Cities, error) {
	qb := r.sqb.
		Insert(Cities{}.TableName()).
		Columns(
			cityCols.ID,
			cityCols.Name,
		)

	for _, city := range cities {
		if city.ID == uuid.Nil {
			city.ID = uuid.New()
		}

		qb = qb.Values(city.ID, city.Name)
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

	createdCities, err := pgx.CollectRows(rows, pgx.RowToStructByName[Cities])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return createdCities, nil
}
