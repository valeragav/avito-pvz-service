package products

import (
	"time"

	"github.com/VaLeraGav/avito-pvz-service/internal/service/products"
	"github.com/google/uuid"
)

type CreateRequest struct {
	Type  string    `json:"type" validate:"required,max=255"`
	PvzID uuid.UUID `json:"pvzID" validate:"required,uuid"`
}

type CreateResponse struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	ReceptionID uuid.UUID `json:"receptionID"`
	DateTime    time.Time `json:"dateTime"`
}

func ToCreateIn(req CreateRequest) products.CreateIn {
	return products.CreateIn{
		TypeName: req.Type,
		PvzID:    req.PvzID,
	}
}

func ToCreateResponse(out products.CreateOut) CreateResponse {
	return CreateResponse{
		ID:          out.ID,
		Type:        out.TypeName,
		ReceptionID: out.ReceptionID,
		DateTime:    out.DateTime,
	}
}
