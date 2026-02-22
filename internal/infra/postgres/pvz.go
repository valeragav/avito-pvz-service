package postgres

import (
	"context"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra/postgres/schema"
	"github.com/valeragav/avito-pvz-service/pkg/listparams"
)

type PVZRepository struct {
	db  DBTX
	sqb sq.StatementBuilderType
}

func NewPVZRepository(db DBTX) *PVZRepository {
	return &PVZRepository{
		db:  db,
		sqb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r PVZRepository) Create(ctx context.Context, pvz domain.PVZ) (*domain.PVZ, error) {
	if pvz.ID == uuid.Nil {
		pvz.ID = uuid.New()
	}

	record := schema.NewPVZ(&pvz)

	qb := r.sqb.
		Insert(record.TableName()).
		Columns(record.InsertColumns()...).
		Values(record.Values()...).
		Suffix("RETURNING " + strings.Join(record.Columns(), ", "))

	pvzCreate, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.PVZ])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainPVZ(pvzCreate), nil
}

func (r *PVZRepository) Get(ctx context.Context, filter domain.PVZ) (*domain.PVZ, error) {
	where := sq.Eq{}
	if filter.ID != uuid.Nil {
		where[schema.PVZCols.ID] = filter.ID
	}
	if !filter.RegistrationDate.IsZero() {
		where[schema.PVZCols.RegistrationDate] = filter.RegistrationDate
	}
	if filter.CityID != uuid.Nil {
		where[schema.PVZCols.CityID] = filter.CityID
	}

	record := schema.NewPVZ(&filter)

	qb := r.sqb.
		Select(record.Columns()...).
		From(record.TableName()).
		Where(where)

	result, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.PVZ])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainPVZ(result), nil
}

func (r *PVZRepository) ListPvzByAcceptanceDateAndCity(ctx context.Context, pagination *listparams.Pagination, startDate, endDate *time.Time) ([]*domain.PVZ, error) {
	qb := r.sqb.
		Select(schema.PVZWithCityName{}.Columns()...).
		From("pvz").
		Join("receptions ON receptions.pvz_id = pvz.id").
		Join("cities ON cities.id = pvz.city_id").
		OrderBy("pvz.registration_date DESC")

	if pagination != nil {
		qb = qb.Limit(uint64(pagination.Limit)).
			Offset(uint64(pagination.Offset()))
	}

	if startDate != nil && endDate != nil {
		qb = qb.Where(
			sq.And{
				sq.GtOrEq{"receptions.date_time": startDate},
				sq.LtOrEq{"receptions.date_time": endDate},
			},
		)
	}

	qb = qb.GroupBy(
		"pvz.id",
		"pvz.city_id",
		"pvz.registration_date",
		"cities.name",
		"cities.id",
	)

	results, err := CollectRows(ctx, r.db, qb, pgx.RowToStructByName[schema.PVZWithCityName])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainPVZWithCityNameList(results), nil
}

func (r *PVZRepository) GetList(ctx context.Context, pagination *listparams.Pagination) ([]*domain.PVZ, error) {
	qb := r.sqb.
		Select(schema.PVZWithCityName{}.Columns()...).
		From(schema.PVZ{}.TableName()).
		Join("cities ON cities.id = pvz.city_id")

	if pagination != nil {
		qb = qb.Limit(uint64(pagination.Limit)).
			Offset(uint64(pagination.Offset()))
	}

	results, err := CollectRows(ctx, r.db, qb, pgx.RowToStructByName[schema.PVZWithCityName])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainPVZWithCityNameList(results), nil
}
