package product

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

//go:generate ${LOCAL_BIN}/mockgen -source=handler.go -destination=./mocks/service_mock.go -package=mocks
type productService interface {
	Create(ctx context.Context, createIn dto.ProductCreate) (*domain.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) (*domain.Product, error)
}

type ProductHandlers struct {
	validator      *validation.Validator
	productService productService
}

func New(validator *validation.Validator, productService productService) *ProductHandlers {
	return &ProductHandlers{
		validator,
		productService,
	}
}

// @Summary Delete the last product of a PVZ
// @Description Deletes the most recently added product for the given PVZ ID.
// @ID DeleteLastProduct
// @Tags PVZ
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param pvzID path string true "PVZ ID (UUID)"
// @Success 200 {object} response.Empty "Successfully deleted"
// @Failure 400 {object} response.Error "pvzID is not recorded"
// @Failure 400 {object} response.Error "Invalid pvzID format"
// @Failure 409 {object} response.Error "No products to delete"
// @Failure 500 {object} response.Error "Internal server error"
// @Router /pvz/{pvzID}/delete_last_product  [post]
func (h *ProductHandlers) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
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

	_, err = h.productService.DeleteLastProduct(ctx, pvzID)
	if err != nil {
		mess, code := mapErrorToHTTP(err)

		logger.ErrorCtx(ctx, mess, "error", err)
		response.WriteError(w, ctx, code, mess, err)
		return
	}

	response.WriteJSON(w, ctx, http.StatusOK, nil)
}

// @Summary Create a new product
// @Description Creates a new product in the PVZ system.
// @ID CreateProduct
// @Tags Product
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body CreateRequest true "Product creation payload"
// @Success 201 {object} CreateResponse "Product successfully created"
// @Failure 400 {object} response.Error "Invalid request or validation failed"
// @Failure 400 {object} response.Error "No reception is currently in progress"
// @Failure 409 {object} response.Error "Product already exists or conflict error"
// @Failure 500 {object} response.Error "Internal server error"
// @Router /products [post]
func (h *ProductHandlers) Create(w http.ResponseWriter, r *http.Request) {
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

	productRes, err := h.productService.Create(ctx, ToCreateIn(req))
	if err != nil {
		mess, code := mapErrorToHTTP(err)

		logger.ErrorCtx(ctx, mess, "error", err)
		response.WriteError(w, ctx, code, mess, err)
		return
	}

	metrics.CreatedProductsInc()

	res := ToCreateResponse(*productRes)
	response.WriteJSON(w, ctx, http.StatusCreated, res)
}

func mapErrorToHTTP(err error) (msg string, statusCode int) {
	switch {
	case errors.Is(err, domain.ErrNoReceptionIsCurrentlyInProgress):
		msg = err.Error()
		statusCode = http.StatusConflict

	case errors.Is(err, domain.ErrProductToDelete):
		msg = err.Error()
		statusCode = http.StatusConflict

	default:
		statusCode = http.StatusInternalServerError
		msg = "internal server error"
	}

	return msg, statusCode
}
