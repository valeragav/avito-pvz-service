package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/response"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
	"github.com/valeragav/avito-pvz-service/internal/validation"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

//go:generate mockgen -source=handler.go -destination=./mocks/service_mock.go -package=mocks
type authService interface {
	GenerateToken(role domain.Role) (*domain.Token, error)
	Login(ctx context.Context, loginReq dto.LoginIn) (*domain.Token, error)
	Register(ctx context.Context, registerReq dto.RegisterIn) (*domain.User, error)
}

type AuthHandlers struct {
	validator   *validation.Validator
	authService authService
}

func New(validator *validation.Validator, authService authService) *AuthHandlers {
	return &AuthHandlers{
		validator,
		authService,
	}
}

// @Summary Dummy login
// @Description Authenticates a user and returns a JWT token for role.
// @ID DummyLogin
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body DummyLoginRequest true "User credentials (email and password)"
// @Success 200 {string} string "JWT token issued successfully"
// @Failure 400 {object} response.Error "Invalid request or validation failed"
// @Failure 401 {object} response.Error "Generate token failed"
// @Failure 500 {object} response.Error "Internal server error"
// @Router /dummyLogin [post]
func (h *AuthHandlers) DummyLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req DummyLoginRequest
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

	token, err := h.authService.GenerateToken(domain.Role(req.Role))
	if err != nil {
		logger.ErrorCtx(ctx, "failed to generate token", "error", err)
		response.WriteError(w, ctx, http.StatusInternalServerError, "generate token failed", err)
		return
	}

	response.WriteString(w, ctx, http.StatusOK, string(*token))
}

// @Summary Register new user
// @Description Creates a new user with role and returns created user data.
// @ID Register
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body RegisterRequest true "User registration payload"
// @Success 201 {object} RegisterResponse "User successfully created"
// @Failure 400 {object} response.Error "Invalid request or validation failed"
// @Failure 409 {object} response.Error "Email already exists"
// @Failure 500 {object} response.Error "Internal server error"
// @Router /register [post]
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RegisterRequest
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

	user, err := h.authService.Register(ctx, ToRegisterIn(req))
	if err != nil {
		mess, code := mapErrorToHTTP(err)

		logger.ErrorCtx(ctx, mess, "error", err)
		response.WriteError(w, ctx, code, mess, err)
		return
	}

	res := ToRegisterResponse(*user)

	response.WriteJSON(w, ctx, http.StatusCreated, res)
}

// @Summary User login
// @Description Authenticate user and returns a JWT Bearer token.
// @ID Login
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body LoginRequest true "User credentials (email and password)"
// @Success 200 {string} string "JWT access token"
// @Failure 400 {object} response.Error "Invalid request or validation failed"
// @Failure 401 {object} response.Error "Invalid email or password"
// @Failure 500 {object} response.Error "Internal server error"
// @Router /login [post]
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req LoginRequest
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

	token, err := h.authService.Login(ctx, ToLoginIn(req))
	if err != nil {
		mess, code := mapErrorToHTTP(err)

		logger.ErrorCtx(ctx, mess, "error", err)
		response.WriteError(w, ctx, code, mess, err)
		return
	}

	response.WriteString(w, ctx, http.StatusOK, string(*token))
}

func mapErrorToHTTP(err error) (msg string, statusCode int) {
	switch {
	case errors.Is(err, domain.ErrAlreadyExists):
		msg = "email already exists"
		statusCode = http.StatusBadRequest

	case errors.Is(err, domain.ErrInvalidEmailOrPassword):
		msg = err.Error()
		statusCode = http.StatusBadRequest

	default:
		statusCode = http.StatusInternalServerError
		msg = "internal server error"
	}

	return msg, statusCode
}
