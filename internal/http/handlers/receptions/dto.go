package receptions

import (
	"time"

	"github.com/VaLeraGav/avito-pvz-service/internal/service/receptions"
	"github.com/google/uuid"
)

type CreateRequest struct {
	PvzID uuid.UUID `json:"pvzID" validate:"required,uuid"`
}

type CreateResponse struct {
	ID       uuid.UUID `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PvzID    uuid.UUID `json:"pvzID"`
	Status   string    `json:"status"`
}

func ToCreateIn(req CreateRequest) receptions.CreateIn {
	return receptions.CreateIn{
		PvzID: req.PvzID,
	}
}

func ToCreateResponse(out receptions.CreateOut) CreateResponse {
	return CreateResponse{
		ID:       out.ID,
		DateTime: out.DateTime,
		PvzID:    out.PvzID,
		Status:   out.Status,
	}
}

type CloseLastReceptionResponse struct {
	ID       uuid.UUID `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PvzID    uuid.UUID `json:"pvzID"`
	Status   string    `json:"status"`
}

func ToCloseLastReceptionResponse(out receptions.CloseLastReceptionOut) CloseLastReceptionResponse {
	return CloseLastReceptionResponse{
		ID:       out.ID,
		DateTime: out.DateTime,
		PvzID:    out.PvzID,
		Status:   out.Status,
	}
}
