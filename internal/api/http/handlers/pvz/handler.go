package pvz

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/response"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
	"github.com/valeragav/avito-pvz-service/internal/metrics"
	"github.com/valeragav/avito-pvz-service/internal/validation"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

//go:generate mockgen -source=handler.go -destination=./mocks/service_mock.go -package=mocks
type pvzService interface {
	Create(ctx context.Context, createIn dto.PVZCreate) (*domain.PVZ, error)
	List(ctx context.Context, pvzListParams *dto.PVZListParams) ([]*domain.PVZ, error)
}

type PvzHandlers struct {
	validator  *validation.Validator
	pvzService pvzService
}

func New(validator *validation.Validator, pvzService pvzService) *PvzHandlers {
	return &PvzHandlers{
		validator,
		pvzService,
	}
}

func (h *PvzHandlers) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pvzListParams, err := getParsePvzParam(r)
	if err != nil {
		response.WriteError(w, ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	pvzListParamsDto := ToBuildPvzListParams(pvzListParams)

	pvzRes, err := h.pvzService.List(ctx, &pvzListParamsDto)
	if err != nil {
		mess, code := mapErrorToHTTP(err)

		logger.ErrorCtx(ctx, mess, "error", err)
		response.WriteError(w, ctx, code, mess, err)
		return
	}

	res := ToListResponse(pvzRes)

	response.WriteJSON(w, ctx, http.StatusOK, res)
}

func (h *PvzHandlers) Create(w http.ResponseWriter, r *http.Request) {
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

	pvzRes, err := h.pvzService.Create(ctx, ToCreateIn(req))
	if err != nil {
		mess, code := mapErrorToHTTP(err)

		logger.ErrorCtx(ctx, mess, "error", err)
		response.WriteError(w, ctx, code, mess, err)
		return
	}

	metrics.CreatedPVZInc()

	res := ToCreateResponse(*pvzRes)
	response.WriteJSON(w, ctx, http.StatusCreated, res)
}

func mapErrorToHTTP(err error) (msg string, statusCode int) {
	switch {
	case errors.Is(err, domain.ErrCityNotFound):
		msg = err.Error()
		statusCode = http.StatusNotFound

	case errors.Is(err, domain.ErrDuplicatePvzID):
		msg = "pvz with this id already exists"
		statusCode = http.StatusConflict

	default:
		statusCode = http.StatusInternalServerError
		msg = "internal server error"
	}

	return msg, statusCode
}
