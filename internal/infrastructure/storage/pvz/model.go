package pvz

import (
	"time"

	"github.com/google/uuid"
)

type Pvz struct {
	ID               uuid.UUID `db:"id"`
	RegistrationDate time.Time `db:"registration_date"`
	CityID           uuid.UUID `db:"city_id"`
}

type PvzWithCityName struct {
	ID               uuid.UUID `db:"id"`
	RegistrationDate time.Time `db:"registration_date"`
	CityID           uuid.UUID `db:"city_id"`
	CityName         string    `db:"city_name"`
}

func (Pvz) TableName() string {
	return "pvz"
}

func (r Pvz) AliasedCols(alias string) []string {
	cols := r.AllCols()

	out := make([]string, 0, len(cols))
	for _, c := range cols {
		out = append(out, alias+"."+c)
	}

	return out
}

func (Pvz) AllCols() []string {
	return []string{
		Cols.ID,
		Cols.RegistrationDate,
		Cols.CityID,
	}
}

var Cols = struct {
	ID               string
	RegistrationDate string
	CityID           string
}{
	"id",
	"registration_date",
	"city_id",
}
