package schema

import (
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

type PVZ struct {
	ID               uuid.UUID `db:"pvz.id"`
	CityID           uuid.UUID `db:"pvz.city_id"`
	RegistrationDate time.Time `db:"pvz.registration_date"`
}

type PVZWithCityName struct {
	PVZ
	City
}

func NewPVZ(d *domain.PVZ) *PVZ {
	return &PVZ{
		ID:               d.ID,
		RegistrationDate: d.RegistrationDate,
		CityID:           d.CityID,
	}
}

func NewDomainPVZ(d PVZ) *domain.PVZ {
	return &domain.PVZ{
		ID:               d.ID,
		RegistrationDate: d.RegistrationDate,
		CityID:           d.CityID,
	}
}

func NewDomainPVZList(d []PVZ) []*domain.PVZ {
	var res = make([]*domain.PVZ, 0, len(d))
	for _, record := range d {
		res = append(res, NewDomainPVZ(record))
	}
	return res
}

func NewDomainPVZWithCityName(d PVZWithCityName) *domain.PVZ {
	return &domain.PVZ{
		ID:               d.PVZ.ID,
		RegistrationDate: d.RegistrationDate,
		CityID:           d.CityID,
		City: &domain.City{
			ID:   d.City.ID,
			Name: d.Name,
		},
	}
}

func NewDomainPVZWithCityNameList(d []PVZWithCityName) []*domain.PVZ {
	var res = make([]*domain.PVZ, 0, len(d))
	for _, record := range d {
		res = append(res, NewDomainPVZWithCityName(record))
	}
	return res
}

func (p PVZWithCityName) Columns() []string {
	res := PVZ{}.Columns()
	res = append(res, City{}.Columns()...)
	return res
}

func (PVZ) TableName() string {
	return "pvz"
}

func (pvz PVZ) InsertColumns() []string {
	return []string{"id", "city_id", "registration_date"}
}

func (pvz PVZ) Columns() []string {
	return []string{"pvz.id as \"pvz.id\"", "pvz.city_id as \"pvz.city_id\"", "pvz.registration_date as \"pvz.registration_date\""}
}

func (pvz PVZ) Values() []any {
	return []any{pvz.ID, pvz.CityID, pvz.RegistrationDate}
}

var PVZCols = struct {
	ID               string
	RegistrationDate string
	CityID           string
}{
	"id",
	"registration_date",
	"city_id",
}
