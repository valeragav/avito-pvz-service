package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/pkg/listparams"
)

type PVZCreate struct {
	ID               uuid.UUID
	CityName         string
	RegistrationDate time.Time
}

type PVZListParams struct {
	Filter     *PVZFilter
	Pagination *listparams.Pagination
}

type PVZFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
}
