package products

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/http/handlers/products/mocks"
	"github.com/valeragav/avito-pvz-service/internal/http/handlers/response"
	"github.com/valeragav/avito-pvz-service/internal/service/products"
	"github.com/valeragav/avito-pvz-service/internal/validation"
	"github.com/valeragav/avito-pvz-service/pkg/testutils"
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
		productsMock  func(*mocks.MockproductsService)
		expectedError *response.ErrorResponse
	}{
		{
			name:         "successful delete",
			pvzIDParam:   productID.String(),
			expectedCode: http.StatusOK,
			productsMock: func(service *mocks.MockproductsService) {
				service.
					EXPECT().
					DeleteLastProduct(gomock.Any(), productID).
					Return(nil)
			},
		},
		{
			name:         "missing pvzID",
			pvzIDParam:   "",
			expectedCode: http.StatusBadRequest,
			expectedError: &response.ErrorResponse{
				Message: "pvzID is not recorded",
			},
		},
		{
			name:         "invalid pvzID format",
			pvzIDParam:   "not-uuid",
			expectedCode: http.StatusBadRequest,
			expectedError: &response.ErrorResponse{
				Message: "invalid pvzID format",
			},
		},
		{
			name:         "service error",
			pvzIDParam:   productID.String(),
			expectedCode: http.StatusInternalServerError,
			productsMock: func(service *mocks.MockproductsService) {
				service.
					EXPECT().
					DeleteLastProduct(gomock.Any(), productID).
					Return(errors.New("storage error"))
			},
			expectedError: &response.ErrorResponse{
				Message: "internal server error",
				Details: "storage error",
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			productsService := mocks.NewMockproductsService(ctrl)
			handler := New(valid, productsService)

			if tt.productsMock != nil {
				tt.productsMock(productsService)
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
				var errorRes response.ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&errorRes)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedError, &errorRes)
			}
		})
	}
}

func TestProductsHandlers_CreateProduct(t *testing.T) {
	testutils.InitTestLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	valid := validation.New()
	validProductID := uuid.New()
	validPvzID := uuid.New()
	validReceptionID := uuid.New()
	validTypeName := "Product 1"
	validDateTime := time.Date(2026, time.February, 11, 10, 30, 0, 0, time.UTC)

	testcases := []struct {
		name          string
		requestBody   any
		expectedCode  int
		productsMock  func(*mocks.MockproductsService)
		expected      *CreateResponse
		expectedError *response.ErrorResponse
	}{
		{
			name: "successful create",
			requestBody: map[string]any{
				"type":  validTypeName,
				"pvzId": validPvzID,
			},
			expectedCode: http.StatusOK,
			productsMock: func(service *mocks.MockproductsService) {
				service.
					EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&products.CreateOut{
						ID:          validProductID,
						TypeName:    validTypeName,
						ReceptionID: validReceptionID,
						DateTime:    validDateTime,
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
			expectedError: &response.ErrorResponse{
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
			productsMock: func(service *mocks.MockproductsService) {
				service.
					EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			expectedError: &response.ErrorResponse{
				Message: "internal server error",
				Details: "storage error",
			},
		},
		// TODO: еще сделать
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			productsService := mocks.NewMockproductsService(ctrl)
			handler := New(valid, productsService)

			if tt.productsMock != nil {
				tt.productsMock(productsService)
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
				var errorRes response.ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&errorRes)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedError, &errorRes)
			}
		})
	}
}
