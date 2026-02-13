package pvz

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VaLeraGav/avito-pvz-service/internal/http/handlers/pvz/mocks"
	"github.com/VaLeraGav/avito-pvz-service/internal/http/handlers/response"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/products"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/pvz"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/receptions"
	pvzService "github.com/VaLeraGav/avito-pvz-service/internal/service/pvz"
	"github.com/VaLeraGav/avito-pvz-service/internal/validation"
	"github.com/VaLeraGav/avito-pvz-service/pkg/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPvzHandlers_List(t *testing.T) {
	testutils.InitTestLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	valid := validation.New()

	validDateTime := time.Date(2026, time.February, 11, 10, 30, 0, 0, time.UTC)

	samplePvzList := &pvzService.PvzListResponse{
		Outs: []pvzService.Out{
			{
				Pvz: pvz.PvzWithCityName{
					ID:               uuid.New(),
					RegistrationDate: validDateTime,
					CityID:           uuid.New(),
					CityName:         "Moscow",
				},
				Receptions: []pvzService.ReceptionsWithProduct{
					{
						Reception: receptions.ReceptionsWithStatus{
							ID:         uuid.New(),
							PvzID:      uuid.New(),
							StatusID:   uuid.New(),
							StatusName: "in_progress",
						},
						Products: []products.ProductsWithTypeName{
							{
								ID:          uuid.New(),
								TypeName:    "Electronics",
								DateTime:    validDateTime,
								TypeId:      uuid.New(),
								ReceptionID: uuid.New(),
							},
							{
								ID:          uuid.New(),
								TypeName:    "Books",
								DateTime:    validDateTime,
								TypeId:      uuid.New(),
								ReceptionID: uuid.New(),
							},
						},
					},
				},
			},
		},
	}

	expectedDto := ToListResponse(samplePvzList)

	testcases := []struct {
		name          string
		requestQuery  string
		productsMock  func(*mocks.MockproductsService)
		expectedCode  int
		expected      []OutResponse
		expectedError *response.ErrorResponse
	}{
		{
			name:         "successful list",
			requestQuery: "",
			expectedCode: http.StatusOK,
			productsMock: func(service *mocks.MockproductsService) {
				service.
					EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(samplePvzList, nil)
			},
			expected: expectedDto,
		},
		{
			name:         "invalid time data",
			requestQuery: "?startDate=2026-01-01&endDate=2026-02-01",
			expectedCode: http.StatusBadRequest,
			expectedError: &response.ErrorResponse{
				Message: "invalid startDate",
			},
		},
		{
			name:         "invalid pagination",
			requestQuery: "?startDate=2026-01-01&endDate=2026-02-01",
			expectedCode: http.StatusBadRequest,
			expectedError: &response.ErrorResponse{
				Message: "invalid startDate",
			},
		},
		{
			name:         "service error",
			requestQuery: "",
			expectedCode: http.StatusInternalServerError,
			productsMock: func(service *mocks.MockproductsService) {
				service.
					EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage error"))
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

			req := httptest.NewRequest("GET", "/pvz"+tt.requestQuery, http.NoBody)

			w := httptest.NewRecorder()
			handler.List(w, req)

			t.Logf("Response code: %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expected != nil {
				var res []OutResponse
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)

				assert.Equal(t, tt.expected, res)
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

func TestPvzHandlers_Create(t *testing.T) {
	testutils.InitTestLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	valid := validation.New()
	validPvzID := uuid.New()
	cityName := "Test"
	validPvzName := "PVZ 1"
	validDateTime := time.Date(2026, time.February, 11, 10, 30, 0, 0, time.UTC)

	_ = validPvzName

	testcases := []struct {
		name          string
		requestBody   any
		productsMock  func(*mocks.MockproductsService)
		expectedCode  int
		expected      *CreateResponse
		expectedError *response.ErrorResponse
	}{
		{
			name: "successful create",
			requestBody: CreateRequest{
				ID:               validPvzID,
				City:             cityName,
				RegistrationDate: validDateTime,
			},
			expectedCode: http.StatusOK,
			productsMock: func(service *mocks.MockproductsService) {
				service.
					EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&pvzService.CreateOut{
						ID:               validPvzID,
						RegistrationDate: validDateTime,
						CityName:         cityName,
					}, nil)
			},
			expected: &CreateResponse{
				ID:               validPvzID,
				RegistrationDate: validDateTime,
				City:             cityName,
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
				ID:               validPvzID,
				City:             cityName,
				RegistrationDate: validDateTime,
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
		{
			name: "validation failed",
			requestBody: map[string]any{
				"ID": uuid.NewString(),
			},
			expectedCode: http.StatusBadRequest,
			expectedError: &response.ErrorResponse{
				Message: "field 'City' failed on the 'required' validation",
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

			bodyReader, err := testutils.MakeRequestBody(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/pvz", bodyReader)

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
