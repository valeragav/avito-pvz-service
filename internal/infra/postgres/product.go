package postgres

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/infra/postgres/schema"
)

type ProductRepository struct {
	db  DBTX
	sqb sq.StatementBuilderType
}

func NewProductRepository(db DBTX) *ProductRepository {
	return &ProductRepository{
		db:  db,
		sqb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r ProductRepository) Create(ctx context.Context, product domain.Product) (*domain.Product, error) {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	record := schema.NewProduct(&product)

	qb := r.sqb.
		Insert(record.TableName()).
		Columns(record.InsertColumns()...).
		Values(record.Values()...).
		Suffix("RETURNING " + strings.Join(record.Columns(), ", "))

	productCreate, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.Product])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainProduct(&productCreate), nil
}

func (r ProductRepository) Get(ctx context.Context, filter domain.Product) (*domain.Product, error) {
	where := sq.Eq{}
	if filter.ID != uuid.Nil {
		where["products.id"] = filter.ID
	}
	if !filter.DateTime.IsZero() {
		where["products.date_time"] = filter.DateTime
	}
	if filter.TypeID != uuid.Nil {
		where["products.type_id"] = filter.TypeID
	}
	if filter.ReceptionID != uuid.Nil {
		where["products.reception_id"] = filter.ReceptionID
	}

	record := schema.NewProduct(&filter)

	qb := r.sqb.
		Select(schema.ProductWithTypeName{}.Columns()...).
		From(record.TableName()).
		Join("product_types ON product_types.id = products.type_id").
		Where(where)

	results, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.ProductWithTypeName])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainProductWithTypeName(results), nil
}

func (r *ProductRepository) GetLastProductInReception(ctx context.Context, receptionID uuid.UUID) (*domain.Product, error) {
	qb := r.sqb.
		Select(schema.ProductWithTypeName{}.Columns()...).
		From(schema.Product{}.TableName()).
		Join("product_types ON product_types.id = products.type_id").
		Where(sq.Eq{schema.ProductCols.ReceptionID: receptionID}).
		OrderBy(fmt.Sprintf("%s %s", schema.ProductCols.DateTime, "DESC")).
		Limit(1)

	result, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.ProductWithTypeName])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainProductWithTypeName(result), nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	qb := r.sqb.
		Delete(schema.Product{}.TableName()).
		Where(sq.Eq{schema.ProductCols.ID: productID})

	sql, args, err := qb.ToSql()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBuildQuery, err)
	}

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExecuteQuery, err)
	}

	if tag.RowsAffected() == 0 {
		return infra.ErrNotFound
	}

	return nil
}

func (r *ProductRepository) ListByReceptionIDsWithTypeName(ctx context.Context, receptionIDs []uuid.UUID) ([]*domain.Product, error) {
	qb := r.sqb.
		Select(schema.ProductWithTypeName{}.Columns()...).
		From(schema.Product{}.TableName()).
		Join("product_types ON product_types.id = products.type_id").
		Where(sq.Eq{"products.reception_id": receptionIDs})

	results, err := CollectRows(ctx, r.db, qb, pgx.RowToStructByName[schema.ProductWithTypeName])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainProductWithTypeNameList(results), nil
}
