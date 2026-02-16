package pvz

import (
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
)

type CreateRequest struct {
	ID               uuid.UUID `json:"id" validate:"required,uuid"`
	City             string    `json:"city" validate:"required,max=255"`
	RegistrationDate time.Time `json:"registrationDate" validate:"required"`
}

type CreateResponse struct {
	ID               uuid.UUID `json:"id"`
	City             string    `json:"city"`
	RegistrationDate time.Time `json:"registrationDate"`
}

type OutResponse struct {
	Pvz        PvzResponse             `json:"pvz"`
	Receptions []ReceptionsWithProduct `json:"receptions"`
}

type PvzResponse struct {
	ID               uuid.UUID `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

type ReceptionsWithProduct struct {
	Reception ReceptionsResponse `json:"reception"`
	Products  []ProductsResponse `json:"products"`
}

type ReceptionsResponse struct {
	ID       uuid.UUID `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PvzID    uuid.UUID `json:"pvzId"`
	Status   string    `json:"status"`
}

type ProductsResponse struct {
	ID          uuid.UUID `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionID uuid.UUID `json:"receptionId"`
}

func ToBuildPvzListParams(pvzListParams PvzListParams) dto.PVZListParams {
	return dto.PVZListParams{
		Filter: dto.PVZFilter{
			StartDate: pvzListParams.Filter.StartDate,
			EndDate:   pvzListParams.Filter.EndDate,
		},
		Pagination: pvzListParams.Pagination,
	}
}

func ToCreateIn(req CreateRequest) dto.PVZCreate {
	return dto.PVZCreate{
		ID:               req.ID,
		CityName:         req.City,
		RegistrationDate: req.RegistrationDate,
	}
}

func ToCreateResponse(out domain.PVZ) CreateResponse {
	var city string
	if out.City != nil {
		city = out.City.Name
	}
	return CreateResponse{
		ID:               out.ID,
		City:             city,
		RegistrationDate: out.RegistrationDate,
	}
}

func ToListResponse(pvzs []*domain.PVZ) []OutResponse {
	result := make([]OutResponse, 0, len(pvzs))
	for _, pvz := range pvzs {
		var city string
		if pvz.City != nil {
			city = pvz.City.Name
		}

		pvzResp := PvzResponse{
			ID:               pvz.ID,
			RegistrationDate: pvz.RegistrationDate,
			City:             city,
		}

		receptionsResp := make([]ReceptionsWithProduct, 0, len(pvz.Receptions))
		for _, rwp := range pvz.Receptions {
			productsResp := make([]ProductsResponse, 0, len(rwp.Products))
			for _, p := range rwp.Products {
				var productTypeName string
				if p.ProductType != nil {
					productTypeName = p.ProductType.Name
				}

				productsResp = append(productsResp, ProductsResponse{
					ID:          p.ID,
					DateTime:    p.DateTime,
					Type:        productTypeName,
					ReceptionID: p.ReceptionID,
				})
			}

			var status string
			if rwp.ReceptionStatus != nil {
				status = string(rwp.ReceptionStatus.Name)
			}

			receptionsResp = append(receptionsResp, ReceptionsWithProduct{
				Reception: ReceptionsResponse{
					ID:       rwp.ID,
					DateTime: rwp.DateTime,
					PvzID:    rwp.PvzID,
					Status:   status,
				},
				Products: productsResp,
			})
		}

		result = append(result, OutResponse{
			Pvz:        pvzResp,
			Receptions: receptionsResp,
		})
	}

	return result
}
