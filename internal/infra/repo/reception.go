package repo

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo/schema"
)

type ReceptionRepository struct {
	db  *pgxpool.Pool
	sqb sq.StatementBuilderType
}

func NewReceptionRepository(db *pgxpool.Pool) *ReceptionRepository {
	return &ReceptionRepository{
		db:  db,
		sqb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r ReceptionRepository) Create(ctx context.Context, reception domain.Reception) (*domain.Reception, error) {
	if reception.ID == uuid.Nil {
		reception.ID = uuid.New()
	}

	record := schema.NewReception(&reception)

	qb := r.sqb.
		Insert(record.TableName()).
		Columns(record.InsertColumns()...).
		Values(record.Values()...).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(record.Columns(), ", ")))

	productCreate, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.Reception])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainReception(&productCreate), nil
}

func (r *ReceptionRepository) GetList(ctx context.Context, filter domain.Reception) ([]*domain.Reception, error) {
	where := ToWhereMap(filter)

	record := schema.NewReception(&filter)

	qb := r.sqb.
		Select(record.Columns()...).
		From(record.TableName()).
		Where(where)

	results, err := CollectRows(ctx, r.db, qb, pgx.RowToStructByName[*schema.Reception])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainReceptionList(results), nil
}

func (r *ReceptionRepository) GetByIDs(ctx context.Context, receptionIDs []uuid.UUID) ([]*domain.Reception, error) {
	qb := r.sqb.
		Select(schema.Reception{}.Columns()...).
		From(schema.Reception{}.TableName()).
		Where(sq.Eq{schema.ReceptionCols.PvzID: receptionIDs})

	results, err := CollectRows(ctx, r.db, qb, pgx.RowToStructByName[*schema.Reception])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainReceptionList(results), nil
}

func (r *ReceptionRepository) Get(ctx context.Context, filter domain.Reception) (*domain.Reception, error) {
	where := ToWhereMap(filter)

	record := schema.NewReception(&filter)

	qb := r.sqb.
		Select(record.Columns()...).
		From(record.TableName()).
		Where(where)

	result, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.Reception])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainReception(&result), nil
}

func (r *ReceptionRepository) ListByIDsWithStatus(ctx context.Context, receptionIDs []uuid.UUID) ([]*domain.Reception, error) {
	// TODO: нужно ли делать запросом или лучше строит строки из entity
	qb := r.sqb.
		Select(schema.ReceptionWithStatus{}.Columns()...).
		From(schema.Reception{}.TableName() + " r").
		Join("reception_statuses ON s.id = r.status_id").
		Where(sq.Eq{"r.pvz_id": receptionIDs})

	results, err := CollectRows(ctx, r.db, qb, pgx.RowToStructByName[*schema.ReceptionWithStatus])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainReceptionWithStatusList(results), nil
}

func (r *ReceptionRepository) FindByStatus(ctx context.Context, statusName domain.ReceptionStatusCode, filter domain.Reception) (*domain.Reception, error) {
	where := ToWhereMap(filter)

	where["reception_statuses.name"] = statusName

	qb := r.sqb.
		Select(schema.ReceptionWithStatus{}.Columns()...).
		From(schema.Reception{}.TableName()).
		Join("reception_statuses ON reception_statuses.id = receptions.status_id").
		Where(where).
		OrderBy("receptions.date_time DESC").
		Limit(1)

	result, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.ReceptionWithStatus])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainReceptionWithStatus(&result), nil
}

func (r *ReceptionRepository) Update(ctx context.Context, receptionID uuid.UUID, update domain.Reception) (*domain.Reception, error) {
	qb := r.sqb.
		Update(schema.Reception{}.TableName()).
		Where(sq.Eq{schema.ReceptionCols.ID: receptionID}).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(schema.Reception{}.Columns(), ", ")))

	var clauses = make(map[string]any)

	if update.ID != uuid.Nil {
		clauses[schema.ReceptionCols.ID] = update.ID
	}
	if !update.DateTime.IsZero() {
		clauses[schema.ReceptionCols.DateTime] = update.DateTime
	}
	if update.PvzID != uuid.Nil {
		clauses[schema.ReceptionCols.PvzID] = update.PvzID
	}
	if update.StatusID != uuid.Nil {
		clauses[schema.ReceptionCols.StatusID] = update.StatusID
	}

	qb = qb.SetMap(clauses)

	result, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.Reception])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainReception(&result), nil
}

func ToWhereMap(elem domain.Reception) sq.Eq {
	where := sq.Eq{}

	if elem.ID != uuid.Nil {
		where[schema.ReceptionCols.ID] = elem.ID
	}
	if !elem.DateTime.IsZero() {
		where[schema.ReceptionCols.DateTime] = elem.DateTime
	}
	if elem.PvzID != uuid.Nil {
		where[schema.ReceptionCols.PvzID] = elem.PvzID
	}
	if elem.StatusID != uuid.Nil {
		where[schema.ReceptionCols.StatusID] = elem.StatusID
	}
	return where
}
