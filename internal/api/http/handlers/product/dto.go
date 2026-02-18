package product

import (
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
)

type CreateRequest struct {
	Type  string    `json:"type" validate:"required,max=255"`
	PvzID uuid.UUID `json:"pvzId" validate:"required,uuid"`
}

type CreateResponse struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	ReceptionID uuid.UUID `json:"receptionId"`
	DateTime    time.Time `json:"dateTime"`
}

func ToCreateIn(req CreateRequest) dto.ProductCreate {
	return dto.ProductCreate{
		TypeName: req.Type,
		PvzID:    req.PvzID,
	}
}

func ToCreateResponse(out domain.Product) CreateResponse {

	var typeName string
	if out.ProductType != nil {
		typeName = out.ProductType.Name
	}

	return CreateResponse{
		ID:          out.ID,
		Type:        typeName,
		ReceptionID: out.ReceptionID,
		DateTime:    out.DateTime,
	}
}
