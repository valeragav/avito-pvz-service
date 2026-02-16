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
