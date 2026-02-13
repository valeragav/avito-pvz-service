package pvz

import (
	"time"

	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/products"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/pvz"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/receptions"
	"github.com/VaLeraGav/avito-pvz-service/pkg/listparams"
	"github.com/google/uuid"
)

type CreateIn struct {
	ID               uuid.UUID
	CityName         string
	RegistrationDate time.Time
}

type CreateOut struct {
	ID               uuid.UUID
	CityName         string
	RegistrationDate time.Time
}

type PvzListParams struct {
	Filter     PvzFilter
	Pagination listparams.Pagination
}

type PvzFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
}

type PvzListResponse struct {
	Outs []Out
}

type Out struct {
	Pvz        pvz.PvzWithCityName
	Receptions []ReceptionsWithProduct
}

type ReceptionsWithProduct struct {
	Reception receptions.ReceptionsWithStatus
	Products  []products.ProductsWithTypeName
}
