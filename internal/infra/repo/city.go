package repo

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo/schema"
)

type CityRepository struct {
	db  *pgxpool.Pool
	sqb sq.StatementBuilderType
}

func NewCityRepository(db *pgxpool.Pool) *CityRepository {
	return &CityRepository{
		db:  db,
		sqb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *CityRepository) Get(ctx context.Context, filter domain.City) (*domain.City, error) {
	where := sq.Eq{}
	if filter.ID != uuid.Nil {
		where[schema.CityCols.ID] = filter.ID
	}
	if filter.Name != "" {
		where[schema.CityCols.Name] = filter.Name
	}

	record := schema.NewCity(&filter)

	qb := r.sqb.
		Select(record.Columns()...).
		From(record.TableName()).
		Where(where)

	result, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.City])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainCities(&result), nil
}

// use it if there are many records
func (r CityRepository) CreateBatchPgx(ctx context.Context, cities []domain.City) error {
	batch := &pgx.Batch{}

	for _, city := range cities {
		if city.ID == uuid.Nil {
			city.ID = uuid.New()
		}

		sql := fmt.Sprintf(
			"INSERT INTO %s (%s, %s) VALUES ($1, $2)",
			schema.City{}.TableName(),
			schema.CityCols.ID,
			schema.CityCols.Name,
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

func (r CityRepository) CreateBatch(ctx context.Context, cities []domain.City) error {
	qb := r.sqb.
		Insert(schema.City{}.TableName()).
		Columns(schema.City{}.InsertColumns()...)

	for _, city := range cities {
		if city.ID == uuid.Nil {
			city.ID = uuid.New()
		}

		qb = qb.Values(city.ID, city.Name)
	}

	_, err := CollectRows(ctx, r.db, qb, pgx.RowToStructByName[schema.City])
	if err != nil {
		return err
	}

	return nil
}
