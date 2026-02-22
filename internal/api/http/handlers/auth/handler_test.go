package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/auth/mocks"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/response"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/validation"
	"github.com/valeragav/avito-pvz-service/pkg/testutils"
	"go.uber.org/mock/gomock"
)

func TestAuthHandlers_DummyLogin(t *testing.T) {
	testutils.InitTestLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	valid := validation.New()

	testcases := []struct {
		name            string
		requestBody     string
		expectedCode    int
		authServiceMock func(*mocks.MockauthService)
		expected        string
		expectedError   *response.Error
	}{
		{
			name:         "successful login",
			requestBody:  fmt.Sprintf(`{"role":%q}`, "moderator"),
			expectedCode: http.StatusOK,
			expected:     "token",
			authServiceMock: func(authService *mocks.MockauthService) {
				token := domain.Token("token")
				authService.
					EXPECT().
					GenerateToken(domain.ModeratorRole).
					Return(&token, nil)
			},
		},
		{
			name:         "empty body",
			requestBody:  "",
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "request body is empty",
				Details: "",
			},
		},
		{
			name:         "validation failed - empty role",
			requestBody:  `{"role":""}`,
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "field 'Role' failed on the 'required' validation",
				Details: "",
			},
		},
		{
			name:         "validation failed - invalid role",
			requestBody:  `{"role":"test"}`,
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "field 'Role' failed on the 'oneofci' validation",
				Details: "",
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			authService := mocks.NewMockauthService(ctrl)
			handler := New(valid, authService)

			if tt.authServiceMock != nil {
				tt.authServiceMock(authService)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/dummylogin", strings.NewReader(tt.requestBody))

			handler.DummyLogin(w, req)

			t.Logf("Response code: %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expected != "" {
				res := w.Body.String()

				assert.NotEmpty(t, res)
				assert.Contains(t, res, tt.expected)
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

func TestAuthHandlers_Register(t *testing.T) {
	testutils.InitTestLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	valid := validation.New()

	validUserID := uuid.New()
	validEmail := "test@example.com"
	validPassword := "valid_password"
	userRoleEmployee := "employee"
	userRoleModerator := "moderator"

	testcases := []struct {
		name            string
		requestBody     map[string]any
		expectedCode    int
		authServiceMock func(*mocks.MockauthService)
		expected        *RegisterResponse
		expectedError   *response.Error
	}{
		{
			name: "successful register - employee",
			requestBody: map[string]any{
				"email":    validEmail,
				"password": validPassword,
				"role":     userRoleEmployee,
			},
			expectedCode: http.StatusCreated,
			expected: &RegisterResponse{
				ID:    validUserID,
				Email: validEmail,
				Role:  userRoleEmployee,
			},
			authServiceMock: func(authService *mocks.MockauthService) {
				authService.
					EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(&domain.User{
						ID:    validUserID,
						Email: validEmail,
						Role:  domain.Role(userRoleEmployee),
					}, nil)
			},
		},
		{
			name: "successful register - moderator",
			requestBody: map[string]any{
				"email":    validEmail,
				"password": validPassword,
				"role":     userRoleModerator,
			},
			expectedCode: http.StatusCreated,
			expected: &RegisterResponse{
				ID:    validUserID,
				Email: validEmail,
				Role:  userRoleModerator,
			},
			authServiceMock: func(authService *mocks.MockauthService) {
				authService.
					EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(&domain.User{
						ID:    validUserID,
						Email: validEmail,
						Role:  domain.Role(userRoleModerator),
					}, nil)
			},
		},
		{
			name: "validation failed - missing email",
			requestBody: map[string]any{
				"password": validPassword,
				"role":     userRoleEmployee,
			},
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "field 'Email' failed on the 'required' validation",
			},
		},
		{
			name: "validation failed - valid email",
			requestBody: map[string]any{
				"email":    "notvalidemail",
				"password": validPassword,
				"role":     userRoleEmployee,
			},
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "field 'Email' failed on the 'email' validation",
			},
		},
		{
			name: "validation failed - invalid role",
			requestBody: map[string]any{
				"email":    validEmail,
				"password": validPassword,
				"role":     "test",
			},
			expectedCode: http.StatusBadRequest,
			authServiceMock: func(authService *mocks.MockauthService) {
				authService.
					EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrInvalidRole)
			},
			expectedError: &response.Error{
				Message: domain.ErrInvalidRole.Error(),
				Details: domain.ErrInvalidRole.Error(),
			},
		},
		{
			name: "service error - email exists",
			requestBody: map[string]any{
				"email":    validEmail,
				"password": validPassword,
				"role":     userRoleEmployee,
			},
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "email already exists",
				Details: domain.ErrAlreadyExists.Error(),
			},
			authServiceMock: func(authService *mocks.MockauthService) {
				authService.
					EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrAlreadyExists)
			},
		},
		{
			name: "service error - infra error",
			requestBody: map[string]any{
				"email":    validEmail,
				"password": validPassword,
				"role":     userRoleEmployee,
			},
			expectedCode: http.StatusInternalServerError,
			expectedError: &response.Error{
				Message: "internal server error",
				Details: "infra error",
			},
			authServiceMock: func(authService *mocks.MockauthService) {
				authService.
					EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("infra error"))
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			authServiceMock := mocks.NewMockauthService(ctrl)
			handler := New(valid, authServiceMock)

			if tt.authServiceMock != nil {
				tt.authServiceMock(authServiceMock)
			}

			bodyReader, err := testutils.MakeRequestBody(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/register", bodyReader)

			w := httptest.NewRecorder()
			handler.Register(w, req)

			t.Logf("Response code: %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expected != nil {
				var res RegisterResponse
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

func TestAuthHandlers_Login(t *testing.T) {
	testutils.InitTestLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	valid := validation.New()

	validEmail := "test@example.com"
	validPassword := "valid_password"

	testcases := []struct {
		name            string
		requestBody     map[string]any
		expectedCode    int
		authServiceMock func(*mocks.MockauthService)
		expected        string
		expectedError   *response.Error
	}{
		{
			name: "successful login - employee",
			requestBody: map[string]any{
				"email":    validEmail,
				"password": validPassword,
			},
			expectedCode: http.StatusOK,
			authServiceMock: func(authService *mocks.MockauthService) {
				tkn := domain.Token("token")
				authService.
					EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(&tkn, nil)
			},
			expected: "token",
		},
		{
			name: "validation failed - missing email",
			requestBody: map[string]any{
				"password": validPassword,
			},
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "field 'Email' failed on the 'required' validation",
			},
		},
		{
			name: "validation failed - valid email",
			requestBody: map[string]any{
				"email":    "notvalidemail",
				"password": validPassword,
			},
			expectedCode: http.StatusBadRequest,
			expectedError: &response.Error{
				Message: "field 'Email' failed on the 'email' validation",
			},
		},
		{
			name: "service error - error invalid email or password",
			requestBody: map[string]any{
				"email":    validEmail,
				"password": validPassword,
			},
			expectedCode: http.StatusBadRequest,
			authServiceMock: func(authService *mocks.MockauthService) {
				authService.
					EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrInvalidEmailOrPassword)
			},
			expectedError: &response.Error{
				Message: domain.ErrInvalidEmailOrPassword.Error(),
				Details: domain.ErrInvalidEmailOrPassword.Error(),
			},
		},
		{
			name: "service error - email already exists",
			requestBody: map[string]any{
				"email":    validEmail,
				"password": validPassword,
			},
			expectedCode: http.StatusBadRequest,
			authServiceMock: func(authService *mocks.MockauthService) {
				authService.
					EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrAlreadyExists)
			},
			expectedError: &response.Error{
				Message: "email already exists",
				Details: domain.ErrAlreadyExists.Error(),
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			authService := mocks.NewMockauthService(ctrl)
			handler := New(valid, authService)

			if tt.authServiceMock != nil {
				tt.authServiceMock(authService)
			}

			bodyReader, err := testutils.MakeRequestBody(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/login", bodyReader)

			w := httptest.NewRecorder()
			handler.Login(w, req)

			t.Logf("Response code: %d", w.Code)
			t.Logf("Response body: %s", w.Body.String())

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expected != "" {
				res := w.Body.String()

				assert.NotEmpty(t, res)
				assert.Contains(t, res, tt.expected)
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
