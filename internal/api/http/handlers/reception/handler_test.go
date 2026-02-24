package reception

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
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/reception/mocks"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/response"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/pkg/testutils"
	"github.com/valeragav/avito-pvz-service/pkg/validation"
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
		name                  string
		requestBody           any
		receptionsServiceMock func(*mocks.MockreceptionService)
		expectedCode          int
		expected              *CreateResponse
		expectedError         *response.Error
	}{
		{
			name: "success",
			requestBody: CreateRequest{
				PvzID: validPvzID,
			},
			receptionsServiceMock: func(s *mocks.MockreceptionService) {
				s.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&domain.Reception{
						ID:       validReceptionID,
						DateTime: validTime,
						PvzID:    validPvzID,
						ReceptionStatus: &domain.ReceptionStatus{
							ID:   validReceptionID,
							Name: "closed",
						},
					}, nil)
			},
			expectedCode: http.StatusCreated,
			expected: &CreateResponse{
				ID:       validReceptionID,
				DateTime: validTime,
				PvzID:    validPvzID,
				Status:   "closed",
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
			name:         "validation failed",
			requestBody:  CreateRequest{},
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "field 'PvzID' failed on the 'required' validation",
			},
		},
		{
			name: "service error",
			requestBody: CreateRequest{
				PvzID: validPvzID,
			},
			receptionsServiceMock: func(s *mocks.MockreceptionService) {
				s.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receptionServiceMock := mocks.NewMockreceptionService(ctrl)
			handler := New(validator, receptionServiceMock)

			if tt.receptionsServiceMock != nil {
				tt.receptionsServiceMock(receptionServiceMock)
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
				var errorRes response.Error
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
		name                  string
		pvzID                 string
		receptionsServiceMock func(*mocks.MockreceptionService)
		expectedCode          int
		expected              *CreateResponse
		expectedError         *response.Error
	}{
		{
			name:  "success",
			pvzID: validPvzID.String(),
			receptionsServiceMock: func(s *mocks.MockreceptionService) {
				s.EXPECT().
					CloseLastReception(gomock.Any(), gomock.Any()).
					Return(&domain.Reception{
						ID:       validReceptionID,
						DateTime: validTime,
						PvzID:    validPvzID,
						ReceptionStatus: &domain.ReceptionStatus{
							ID:   validReceptionID,
							Name: "closed",
						},
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
			expectedError: &response.Error{
				Message: "pvzID is not recorded",
			},
		},
		{
			name:         "invalid uuid",
			pvzID:        "invalid",
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "invalid pvz format",
			},
		},
		{
			name:  "service error",
			pvzID: validPvzID.String(),
			receptionsServiceMock: func(s *mocks.MockreceptionService) {
				s.EXPECT().
					CloseLastReception(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedError: &response.Error{
				Message: "internal server error",
				Details: "storage error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receptionServiceMock := mocks.NewMockreceptionService(ctrl)
			handler := New(validator, receptionServiceMock)

			if tt.receptionsServiceMock != nil {
				tt.receptionsServiceMock(receptionServiceMock)
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
				var errorRes response.Error
				err := json.NewDecoder(w.Body).Decode(&errorRes)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedError, &errorRes)
			}
		})
	}
}
