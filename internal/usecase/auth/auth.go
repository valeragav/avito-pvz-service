package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"golang.org/x/crypto/bcrypt"
)

//go:generate ${LOCAL_BIN}/mockgen -source=auth.go -destination=./mocks/auth_mock.go -package=mocks
type jwtService interface {
	SignJwt(userClaims domain.UserClaims) (string, error)
	ValidateJwt(incomingToken string) (*domain.UserClaims, error)
}

type userRepository interface {
	Create(ctx context.Context, user domain.User) (*domain.User, error)
	Get(ctx context.Context, filter domain.User) (*domain.User, error)
}

type AuthUseCase struct {
	jwtService jwtService
	userRepo   userRepository
}

func New(jwtService jwtService, userRepo userRepository) *AuthUseCase {
	return &AuthUseCase{
		jwtService,
		userRepo,
	}
}

func (s *AuthUseCase) GenerateToken(role domain.Role) (*domain.Token, error) {
	const op = "auth.GenerateToken"

	token, err := s.jwtService.SignJwt(domain.UserClaims{
		Role: role,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate token: %w", op, err)
	}

	domainToken := domain.Token(token)

	return &domainToken, nil
}

func (s *AuthUseCase) Register(ctx context.Context, registerReq dto.RegisterIn) (*domain.User, error) {
	const op = "auth.Register"

	role := domain.Role(strings.ToLower(registerReq.Role))

	// По хорошему нужно сделать проверку роли через запрос в талицу, что роль существует
	if !role.IsValid() {
		return nil, domain.ErrInvalidRole
	}

	exists, err := s.userRepo.Get(ctx, domain.User{Email: registerReq.Email})
	if err != nil && !errors.Is(err, infra.ErrNotFound) {
		return nil, fmt.Errorf("%s: failed to check if user exists: %w", op, err)
	}
	if exists != nil {
		return nil, domain.ErrAlreadyExists
	}

	domainUser, err := domain.NewUser(registerReq.Email, registerReq.Password, role)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	createdUser, err := s.userRepo.Create(ctx, *domainUser)
	if err != nil {
		return nil, fmt.Errorf("%s: to create user: %w", op, err)
	}

	return createdUser, nil
}

func (s *AuthUseCase) Login(ctx context.Context, loginReq dto.LoginIn) (*domain.Token, error) {
	const op = "auth.Login"

	userFound, err := s.userRepo.Get(ctx, domain.User{Email: loginReq.Email})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get user: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(userFound.PasswordHash), []byte(loginReq.Password))
	if err != nil {
		return nil, domain.ErrInvalidEmailOrPassword
	}

	token, err := s.jwtService.SignJwt(domain.UserClaims{
		Role: userFound.Role,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate token: %w", op, err)
	}

	domainToken := domain.Token(token)

	return &domainToken, nil
}
