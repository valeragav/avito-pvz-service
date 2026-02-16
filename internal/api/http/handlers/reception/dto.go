package reception

import (
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
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

func ToCreateIn(req CreateRequest) dto.ReceptionCreate {
	return dto.ReceptionCreate{
		PvzID: req.PvzID,
	}
}

func ToCreateResponse(out domain.Reception) CreateResponse {
	var status string
	if out.ReceptionStatus != nil {
		status = string(out.ReceptionStatus.Name)
	}

	return CreateResponse{
		ID:       out.ID,
		DateTime: out.DateTime,
		PvzID:    out.PvzID,
		Status:   status,
	}
}

type CloseLastReceptionResponse struct {
	ID       uuid.UUID `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PvzID    uuid.UUID `json:"pvzID"`
	Status   string    `json:"status"`
}

func ToCloseLastReceptionResponse(out domain.Reception) CloseLastReceptionResponse {
	var status string
	if out.ReceptionStatus != nil {
		status = string(out.ReceptionStatus.Name)
	}
	return CloseLastReceptionResponse{
		ID:       out.ID,
		DateTime: out.DateTime,
		PvzID:    out.PvzID,
		Status:   status,
	}
}
