package products

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/http/handlers/response"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage"
	"github.com/valeragav/avito-pvz-service/internal/service/products"
	"github.com/valeragav/avito-pvz-service/internal/validation"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

type ProductsHandlers struct {
	validator       *validation.Validator
	productsService productsService
}

func New(validator *validation.Validator, productsService productsService) *ProductsHandlers {
	return &ProductsHandlers{
		validator,
		productsService,
	}
}

func (h *ProductsHandlers) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pvzIDParam := chi.URLParam(r, "pvzID")

	if pvzIDParam == "" {
		response.WriteError(w, ctx, http.StatusBadRequest, "pvzID is not recorded", nil)
		return
	}

	pvzID, err := uuid.Parse(pvzIDParam)
	if err != nil {
		response.WriteError(w, ctx, http.StatusBadRequest, "invalid pvzID format", nil)
		return
	}

	err = h.productsService.DeleteLastProduct(ctx, pvzID)
	if err != nil {
		mess, code := mapErrorToHTTP(err)

		logger.ErrorCtx(ctx, mess, "error", err)
		response.WriteError(w, ctx, code, mess, err)
		return
	}

	response.WriteJSON(w, ctx, http.StatusOK, nil)
}

func (h *ProductsHandlers) Create(w http.ResponseWriter, r *http.Request) {
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

	productRes, err := h.productsService.Create(ctx, ToCreateIn(req))
	if err != nil {
		mess, code := mapErrorToHTTP(err)

		logger.ErrorCtx(ctx, mess, "error", err)
		response.WriteError(w, ctx, code, mess, err)
		return
	}

	res := ToCreateResponse(*productRes)

	response.WriteJSON(w, ctx, http.StatusOK, res)
}

func mapErrorToHTTP(err error) (msg string, statusCode int) {
	statusCode = http.StatusInternalServerError
	msg = "internal server error"

	switch {
	case errors.Is(err, storage.ErrNotFound):
		msg = "not found such product"
		statusCode = http.StatusBadRequest

	case errors.Is(err, products.ErrNotFoundReceptionsRepoInProgress):
		msg = err.Error()
		statusCode = http.StatusBadRequest
	}

	return msg, statusCode
}
