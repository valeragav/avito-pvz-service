package pvz

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/VaLeraGav/avito-pvz-service/internal/http/handlers/response"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage"
	"github.com/VaLeraGav/avito-pvz-service/internal/validation"
	"github.com/VaLeraGav/avito-pvz-service/pkg/logger"
)

type PvzHandlers struct {
	validator  *validation.Validator
	pvzService productsService
}

func New(validator *validation.Validator, pvzService productsService) *PvzHandlers {
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

	res := ToCreateResponse(*pvzRes)

	response.WriteJSON(w, ctx, http.StatusOK, res)
}

func mapErrorToHTTP(err error) (msg string, statusCode int) {
	statusCode = http.StatusInternalServerError
	msg = "internal server error"

	switch {
	case errors.Is(err, storage.ErrAlreadyExists):
		msg = "email already exists"
		statusCode = http.StatusBadRequest

	case errors.Is(err, storage.ErrNotFound):
		msg = "not found such user"
		statusCode = http.StatusBadRequest
	}

	return msg, statusCode
}
