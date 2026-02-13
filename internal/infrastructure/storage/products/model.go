package products

import (
	"time"

	"github.com/google/uuid"
)

type Products struct {
	ID          uuid.UUID `db:"id"`
	DateTime    time.Time `db:"date_time"`
	TypeIs      uuid.UUID `db:"type_id"`
	ReceptionID uuid.UUID `db:"reception_id"`
}

type ProductsWithTypeName struct {
	ID          uuid.UUID `db:"id"`
	DateTime    time.Time `db:"date_time"`
	TypeId      uuid.UUID `db:"type_id"`
	TypeName    string    `db:"type_name"`
	ReceptionID uuid.UUID `db:"reception_id"`
}

func (Products) TableName() string {
	return "products"
}

func (Products) AllCols() []string {
	return []string{
		Cols.ID,
		Cols.DateTime,
		Cols.TypeIs,
		Cols.ReceptionID,
	}
}

var Cols = struct {
	ID          string
	DateTime    string
	TypeIs      string
	ReceptionID string
}{
	"id",
	"date_time",
	"type_id",
	"reception_id",
}
