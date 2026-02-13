package receptions

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VaLeraGav/avito-pvz-service/internal/http/handlers/receptions/mocks"
	"github.com/VaLeraGav/avito-pvz-service/internal/http/handlers/response"
	"github.com/VaLeraGav/avito-pvz-service/internal/service/receptions"
	"github.com/VaLeraGav/avito-pvz-service/internal/validation"
	"github.com/VaLeraGav/avito-pvz-service/pkg/testutils"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestReceptionsHandlers_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validator := validation.New()

	validPvzID := uuid.New()
	validReceptionID := uuid.New()
	validTime := time.Now().UTC()

	tests := []struct {
		name           string
		requestBody    any
		receptionsMock func(*mocks.MockreceptionsService)
		expectedCode   int
		expected       *CreateResponse
		expectedError  *response.ErrorResponse
	}{
		{
			name: "success",
			requestBody: CreateRequest{
				PvzID: validPvzID,
			},
			receptionsMock: func(s *mocks.MockreceptionsService) {
				s.EXPECT().
					Create(gomock.Any(), receptions.CreateIn{PvzID: validPvzID}).
					Return(&receptions.CreateOut{
						ID:       validReceptionID,
						DateTime: validTime,
						PvzID:    validPvzID,
						Status:   "open",
					}, nil)
			},
			expectedCode: http.StatusOK,
			expected: &CreateResponse{
				ID:       validReceptionID,
				DateTime: validTime,
				PvzID:    validPvzID,
				Status:   "open",
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
			name:         "validation failed",
			requestBody:  CreateRequest{},
			expectedCode: http.StatusBadRequest,
			expectedError: &response.ErrorResponse{
				Message: "field 'PvzID' failed on the 'required' validation",
			},
		},
		{
			name: "service error",
			requestBody: CreateRequest{
				PvzID: validPvzID,
			},
			receptionsMock: func(s *mocks.MockreceptionsService) {
				s.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := mocks.NewMockreceptionsService(ctrl)
			handler := New(validator, service)

			if tt.receptionsMock != nil {
				tt.receptionsMock(service)
			}

			bodyReader, err := testutils.MakeRequestBody(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/receptions", bodyReader)

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

func TestReceptionsHandlers_CloseLastReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validator := validation.New()

	validPvzID := uuid.New()
	validReceptionID := uuid.New()
	validTime := time.Now().UTC()

	tests := []struct {
		name           string
		pvzID          string
		receptionsMock func(*mocks.MockreceptionsService)
		expectedCode   int
		expected       *CreateResponse
		expectedError  *response.ErrorResponse
	}{
		{
			name:  "success",
			pvzID: validPvzID.String(),
			receptionsMock: func(s *mocks.MockreceptionsService) {
				s.EXPECT().
					CloseLastReception(gomock.Any(), validPvzID).
					Return(&receptions.CloseLastReceptionOut{
						ID:       validReceptionID,
						DateTime: validTime,
						PvzID:    validPvzID,
						Status:   "closed",
					}, nil)
			},
			expectedCode: http.StatusOK,
			expected: &CreateResponse{
				ID:       validReceptionID,
				DateTime: validTime,
				PvzID:    validPvzID,
				Status:   "closed",
			},
		},
		{
			name:         "missing pvzID",
			pvzID:        "",
			expectedCode: http.StatusBadRequest,
			expectedError: &response.ErrorResponse{
				Message: "pvzID is not recorded",
			},
		},
		{
			name:         "invalid uuid",
			pvzID:        "invalid",
			expectedCode: http.StatusBadRequest,
			expectedError: &response.ErrorResponse{
				Message: "invalid pvz format",
			},
		},
		{
			name:  "service error",
			pvzID: validPvzID.String(),
			receptionsMock: func(s *mocks.MockreceptionsService) {
				s.EXPECT().
					CloseLastReception(gomock.Any(), validPvzID).
					Return(nil, errors.New("storage error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedError: &response.ErrorResponse{
				Message: "internal server error",
				Details: "storage error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := mocks.NewMockreceptionsService(ctrl)
			handler := New(validator, service)

			if tt.receptionsMock != nil {
				tt.receptionsMock(service)
			}

			req := httptest.NewRequest(http.MethodPost, "/receptions/"+tt.pvzID, http.NoBody)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("pvzID", tt.pvzID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler.CloseLastReception(w, req)

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
