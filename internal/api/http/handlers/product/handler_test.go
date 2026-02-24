package product

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/product/mocks"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/response"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/pkg/testutils"
	"github.com/valeragav/avito-pvz-service/pkg/validation"
	"go.uber.org/mock/gomock"
)

func TestProductsHandlers_DeleteLastProduct(t *testing.T) {
	testutils.InitTestLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	valid := validation.New()
	productID := uuid.New()

	testcases := []struct {
		name          string
		pvzIDParam    string
		expectedCode  int
		productMock   func(*mocks.MockproductService)
		expectedError *response.Error
	}{
		{
			name:         "successful delete",
			pvzIDParam:   productID.String(),
			expectedCode: http.StatusOK,
			productMock: func(service *mocks.MockproductService) {
				service.
					EXPECT().
					DeleteLastProduct(gomock.Any(), productID).
					Return(nil, nil)
			},
		},
		{
			name:         "missing pvzID",
			pvzIDParam:   "",
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "pvzID is not recorded",
			},
		},
		{
			name:         "invalid pvzID format",
			pvzIDParam:   "not-uuid",
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "invalid pvzID format",
			},
		},
		{
			name:         "service error",
			pvzIDParam:   productID.String(),
			expectedCode: http.StatusInternalServerError,
			productMock: func(service *mocks.MockproductService) {
				service.
					EXPECT().
					DeleteLastProduct(gomock.Any(), productID).
					Return(nil, errors.New("storage error"))
			},
			expectedError: &response.Error{
				Message: "internal server error",
				Details: "storage error",
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			productServiceMock := mocks.NewMockproductService(ctrl)
			handler := New(valid, productServiceMock)

			if tt.productMock != nil {
				tt.productMock(productServiceMock)
			}

			req := httptest.NewRequest("POST", "/pvz/"+tt.pvzIDParam+"/delete_last_product", http.NoBody)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("pvzID", tt.pvzIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.DeleteLastProduct(w, req)

			t.Logf("Response code: %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedError != nil {
				var errorRes response.Error
				err := json.NewDecoder(w.Body).Decode(&errorRes)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedError, &errorRes)
			}
		})
	}
}

// TODO: ErrNoReceptionIsCurrentlyInProgress

func TestProductsHandlers_CreateProduct(t *testing.T) {
	testutils.InitTestLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	valid := validation.New()
	validProductID := uuid.New()
	validPvzID := uuid.New()
	validReceptionID := uuid.New()
	validTypeID := uuid.New()
	validTypeName := "Product 1"
	validDateTime := time.Date(2026, time.February, 11, 10, 30, 0, 0, time.UTC)

	testcases := []struct {
		name          string
		requestBody   any
		expectedCode  int
		productMock   func(*mocks.MockproductService)
		expected      *CreateResponse
		expectedError *response.Error
	}{
		{
			name: "successful create",
			requestBody: map[string]any{
				"type":  validTypeName,
				"pvzId": validPvzID,
			},
			expectedCode: http.StatusCreated,
			productMock: func(service *mocks.MockproductService) {
				service.
					EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&domain.Product{
						ID:          validProductID,
						TypeID:      validTypeID,
						ReceptionID: validReceptionID,
						DateTime:    validDateTime,
						ProductType: &domain.ProductType{
							ID:   validTypeID,
							Name: validTypeName,
						},
					}, nil)
			},
			expected: &CreateResponse{
				ID:          validProductID,
				Type:        validTypeName,
				ReceptionID: validReceptionID,
				DateTime:    validDateTime,
			},
		},
		{
			name:         "empty body",
			requestBody:  "",
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "request body is empty",
			},
		},
		{
			name: "service error",
			requestBody: CreateRequest{
				Type:  validTypeName,
				PvzID: validPvzID,
			},
			expectedCode: http.StatusInternalServerError,
			productMock: func(service *mocks.MockproductService) {
				service.
					EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			expectedError: &response.Error{
				Message: "internal server error",
				Details: "storage error",
			},
		},
		// TODO: еще сделать
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			productsService := mocks.NewMockproductService(ctrl)
			handler := New(valid, productsService)

			if tt.productMock != nil {
				tt.productMock(productsService)
			}

			bodyReader, err := testutils.MakeRequestBody(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/products", bodyReader)

			w := httptest.NewRecorder()
			handler.Create(w, req)

			t.Logf("Response code: %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expected != nil {
				var res CreateResponse
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)

				assert.Equal(t, tt.expected, &res)
			}

			if tt.expectedError != nil {
				var errorRes response.Error
				err := json.NewDecoder(w.Body).Decode(&errorRes)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedError, &errorRes)
			}
		})
	}
}
