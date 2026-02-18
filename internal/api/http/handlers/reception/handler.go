package reception

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/response"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
	"github.com/valeragav/avito-pvz-service/internal/metrics"
	"github.com/valeragav/avito-pvz-service/internal/validation"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

//go:generate mockgen -source=handler.go -destination=./mocks/service_mock.go -package=mocks
type receptionService interface {
	CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*domain.Reception, error)
	Create(ctx context.Context, createIn dto.ReceptionCreate) (*domain.Reception, error)
}

type ReceptionHandlers struct {
	validator        *validation.Validator
	receptionService receptionService
}

func New(validator *validation.Validator, receptionService receptionService) *ReceptionHandlers {
	return &ReceptionHandlers{
		validator,
		receptionService,
	}
}

// @Summary Create Reception
// @Description Create a new reception entry. Requires JWT-Token with Employee role.
// @ID CreateReception
// @Tags Reception
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body CreateRequest true "Reception creation data"
// @Success 201 {object} CreateResponse "Reception successfully created"
// @Failure 400 {object} response.Error "Invalid request or validation failed"
// @Failure 404 {object} response.Error "PVZ not found"
// @Failure 500 {object} response.Error "Internal server error"
// @Router /receptions [post]
func (h *ReceptionHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			response.WriteError(w, ctx, http.StatusBadRequest, "request body is empty", nil)
			return
		}
		response.WriteError(w, ctx, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.WriteError(w, ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	receptionRes, err := h.receptionService.Create(ctx, ToCreateIn(req))
	if err != nil {
		mess, code := mapErrorToHTTP(err)

		logger.ErrorCtx(ctx, mess, "error", err)
		response.WriteError(w, ctx, code, mess, err)
		return
	}

	metrics.CreatedReceptionsInc()

	res := ToCreateResponse(*receptionRes)
	response.WriteJSON(w, ctx, http.StatusCreated, res)
}

// @Summary Close Last Reception
// @Description Close the last reception for a given PVZ ID. Requires JWT-Token with Employee role.
// @ID CloseLastReception
// @Tags PVZ
// @Security ApiKeyAuth
// @Produce json
// @Param pvzID path string true "PVZ ID"
// @Success 200 {object} CloseLastReceptionResponse "Successfully closed last reception"
// @Failure 400 {object} response.Error "Invalid or missing PVZ ID"
// @Failure 404 {object} response.Error "No open reception found for this PVZ"
// @Failure 500 {object} response.Error "Internal server error"
// @Router /pvz/{pvzID}/close_last_reception [post]
func (h *ReceptionHandlers) CloseLastReception(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pvzIDParam := chi.URLParam(r, "pvzID")

	if pvzIDParam == "" {
		response.WriteError(w, ctx, http.StatusBadRequest, "pvzID is not recorded", nil)
		return
	}

	pvzID, err := uuid.Parse(pvzIDParam)
	if err != nil {
		response.WriteError(w, ctx, http.StatusBadRequest, "invalid pvz format", nil)
		return
	}

	pvzRes, err := h.receptionService.CloseLastReception(ctx, pvzID)
	if err != nil {
		mess, code := mapErrorToHTTP(err)

		logger.ErrorCtx(ctx, mess, "error", err)
		response.WriteError(w, ctx, code, mess, err)
		return
	}

	res := ToCloseLastReceptionResponse(*pvzRes)

	response.WriteJSON(w, ctx, http.StatusOK, res)
}

func mapErrorToHTTP(err error) (msg string, statusCode int) {
	switch {
	case errors.Is(err, domain.ErrNoReceptionIsCurrentlyInProgress):
		msg = err.Error()
		statusCode = http.StatusNotFound

	case errors.Is(err, domain.ErrReceptionNotFound), errors.Is(err, domain.ErrPVZNotFound):
		msg = err.Error()
		statusCode = http.StatusBadRequest

	default:
		statusCode = http.StatusInternalServerError
		msg = "internal server error"
	}

	return msg, statusCode
}
