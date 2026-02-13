package pvz

import (
	"time"

	"github.com/VaLeraGav/avito-pvz-service/internal/service/pvz"
	"github.com/google/uuid"
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

func ToBuildPvzListParams(pvzListParams PvzListParams) pvz.PvzListParams {
	return pvz.PvzListParams{
		Filter: pvz.PvzFilter{
			StartDate: pvzListParams.Filter.StartDate,
			EndDate:   pvzListParams.Filter.EndDate,
		},
		Pagination: pvzListParams.Pagination,
	}
}

func ToCreateIn(req CreateRequest) pvz.CreateIn {
	return pvz.CreateIn{
		ID:               req.ID,
		CityName:         req.City,
		RegistrationDate: req.RegistrationDate,
	}
}

func ToCreateResponse(out pvz.CreateOut) CreateResponse {
	return CreateResponse{
		ID:               out.ID,
		City:             out.CityName,
		RegistrationDate: out.RegistrationDate,
	}
}

func ToListResponse(out *pvz.PvzListResponse) []OutResponse {
	result := make([]OutResponse, 0, len(out.Outs))
	for _, out := range out.Outs {
		pvzResp := PvzResponse{
			ID:               out.Pvz.ID,
			RegistrationDate: out.Pvz.RegistrationDate,
			City:             out.Pvz.CityName,
		}

		receptionsResp := make([]ReceptionsWithProduct, 0, len(out.Receptions))
		for _, rwp := range out.Receptions {
			productsResp := make([]ProductsResponse, 0, len(rwp.Products))
			for _, p := range rwp.Products {
				productsResp = append(productsResp, ProductsResponse{
					ID:          p.ID,
					DateTime:    p.DateTime,
					Type:        p.TypeName,
					ReceptionID: p.ReceptionID,
				})
			}

			receptionsResp = append(receptionsResp, ReceptionsWithProduct{
				Reception: ReceptionsResponse{
					ID:       rwp.Reception.ID,
					DateTime: rwp.Reception.DateTime,
					PvzID:    rwp.Reception.PvzID,
					Status:   rwp.Reception.StatusName,
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
