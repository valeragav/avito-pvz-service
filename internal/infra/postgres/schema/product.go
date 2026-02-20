package schema

import (
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

type Product struct {
	ID          uuid.UUID `db:"products.id"`
	DateTime    time.Time `db:"products.date_time"`
	TypeID      uuid.UUID `db:"products.type_id"`
	ReceptionID uuid.UUID `db:"products.reception_id"`
}

type ProductWithTypeName struct {
	Product
	ProductType
}

func NewProduct(d *domain.Product) *Product {
	return &Product{
		ID:          d.ID,
		DateTime:    d.DateTime,
		TypeID:      d.TypeID,
		ReceptionID: d.ReceptionID,
	}
}

func NewDomainProduct(d *Product) *domain.Product {
	return &domain.Product{
		ID:          d.ID,
		DateTime:    d.DateTime,
		TypeID:      d.TypeID,
		ReceptionID: d.ReceptionID,
	}
}

func NewDomainProductWithTypeName(d ProductWithTypeName) *domain.Product {
	return &domain.Product{
		ID:          d.Product.ID,
		DateTime:    d.DateTime,
		TypeID:      d.TypeID,
		ReceptionID: d.ReceptionID,
		ProductType: &domain.ProductType{
			ID:   d.ProductType.ID,
			Name: d.ProductType.Name,
		},
	}
}

func NewDomainProductWithTypeNameList(d []ProductWithTypeName) []*domain.Product {
	var res = make([]*domain.Product, 0, len(d))
	for _, record := range d {
		res = append(res, NewDomainProductWithTypeName(record))
	}
	return res
}

func (p ProductWithTypeName) Columns() []string {
	res := Product{}.Columns()
	res = append(res, ProductType{}.Columns()...)
	return res
}

func (Product) TableName() string {
	return "products"
}

func (p Product) InsertColumns() []string {
	return []string{"id", "date_time", "type_id", "reception_id"}
}

func (p Product) Columns() []string {
	return []string{"products.id as \"products.id\"", "products.date_time as \"products.date_time\"", "products.type_id as \"products.type_id\"",
		"products.reception_id as \"products.reception_id\""}
}

func (p Product) Values() []any {
	return []any{p.ID, p.DateTime, p.TypeID, p.ReceptionID}
}

var ProductCols = struct {
	ID          string
	DateTime    string
	TypeID      string
	ReceptionID string
}{
	"id",
	"date_time",
	"type_id",
	"reception_id",
}
