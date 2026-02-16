package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo/schema"
	"github.com/valeragav/avito-pvz-service/pkg/listparams"
)

type PVZRepository struct {
	db  *pgxpool.Pool
	sqb sq.StatementBuilderType
}

func NewPvzRepository(db *pgxpool.Pool) *PVZRepository {
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
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(record.Columns(), ", ")))

	pvzCreate, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.PVZ])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainPVZ(&pvzCreate), nil
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

	return schema.NewDomainPVZ(&result), nil
}

//  TODO: передавать dto
// func (r *PVZRepository) ListPvzByAcceptanceDate(ctx context.Context, pagination listparams.Pagination, startDate, endDate *time.Time) ([]*domain.PVZ, error) {
// 	pvzAlias := "p"
// 	receptionsAlias := "r"

// 	pvzTable := schema.PVZ{}.TableName()
// 	receptionsTable := schema.Reception{}.TableName()

// 	queryBuilder := r.sqb.
// 		Select(schema.PVZ{}.AliasedCols(pvzAlias)...).
// 		From(pvzTable + " " + pvzAlias)

// 	if startDate != nil || endDate != nil {
// 		queryBuilder = queryBuilder.Join(
// 			fmt.Sprintf("%s %s ON %s.id = %s.pvz_id", receptionsTable, receptionsAlias, pvzAlias, receptionsAlias),
// 		)

// 		dataTimeCol := receptionsAlias + "." + receptions.Cols.DateTime
// 		if startDate != nil {
// 			queryBuilder = queryBuilder.Where(sq.GtOrEq{dataTimeCol: startDate})
// 		}

// 		if endDate != nil {
// 			queryBuilder = queryBuilder.Where(sq.LtOrEq{dataTimeCol: endDate})
// 		}

// 		queryBuilder = queryBuilder.GroupBy(pvzAlias + ".id")
// 	}

// 	offset := pagination.Offset()
// 	queryBuilder = queryBuilder.
// 		OrderBy(pvzAlias + ".registration_date DESC").
// 		Offset(uint64(offset)).
// 		Limit(uint64(pagination.Limit))

// 	sql, args, err := queryBuilder.ToSql()
// 	if err != nil {
// 		return nil, fmt.Errorf("%w: %w", infra.ErrBuildQuery, err)
// 	}

// 	rows, err := r.db.Query(ctx, sql, args...)
// 	if err != nil {
// 		return nil, fmt.Errorf("%w: %w", infra.ErrExecuteQuery, err)
// 	}
// 	defer rows.Close()

// 	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[Pvz])
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return nil, fmt.Errorf("%w: %w", infra.ErrNotFound, err)
// 		}
// 		return nil, fmt.Errorf("%w: %w", infra.ErrScanResult, err)
// 	}

// 	return results, nil
// }

func (r *PVZRepository) ListPvzByAcceptanceDateAndCity(ctx context.Context, pagination listparams.Pagination, startDate, endDate *time.Time) ([]*domain.PVZ, error) {
	qb := r.sqb.
		Select(
			"p.id as id",
			"p.registration_date as registration_date",
			"p.city_id as city_id",
			"c.name AS city_name",
		).
		From("pvz p").
		Join("receptions r ON r.pvz_id = p.id").
		Join("cities c ON c.id = p.city_id").
		OrderBy("p.registration_date DESC").
		Limit(uint64(pagination.Limit)).
		Offset(uint64(pagination.Offset()))

	if startDate != nil && endDate != nil {
		qb = qb.Where(
			sq.And{
				sq.GtOrEq{"r.date_time": startDate},
				sq.LtOrEq{"r.date_time": endDate},
			},
		)
	}

	qb = qb.GroupBy(
		"p.id",
		"p.registration_date",
		"p.city_id",
		"c.name",
	)

	results, err := CollectRows(ctx, r.db, qb, pgx.RowToStructByName[*schema.PVZWithCityName])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainPVZWithCityNameList(results), nil
}

func (r *PVZRepository) GetList(ctx context.Context) ([]*domain.PVZ, error) {
	qb := r.sqb.
		Select(schema.PVZ{}.Columns()...).
		From(schema.PVZ{}.TableName())

	results, err := CollectRows(ctx, r.db, qb, pgx.RowToStructByName[*schema.PVZ])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainPVZList(results), nil
}
