package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/user"
	"github.com/valeragav/avito-pvz-service/internal/security"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	jwtService *security.JwtService
	userRepo   RepositoryUser
}

func New(jwtService *security.JwtService, userRepo RepositoryUser) *AuthService {
	return &AuthService{
		jwtService,
		userRepo,
	}
}

func (s *AuthService) GenerateToken(role string) (string, error) {
	const op = "auth.GenerateToken"

	token, err := s.jwtService.SignJwt(security.Claims{
		Role: role,
	})
	if err != nil {
		return "", fmt.Errorf("%s: failed to generate token: %w", op, err)
	}

	return token, nil
}

func (s *AuthService) Register(ctx context.Context, registerReq RegisterIn) (*RegisterOut, error) {
	const op = "auth.Register"

	exists, err := s.userRepo.Get(ctx, user.User{Email: registerReq.Email})
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, fmt.Errorf("%s: failed to check if user exists: %w", op, err)
	}
	if exists != nil {
		return nil, storage.ErrAlreadyExists
	}

	ent := ToEnt(registerReq)

	ent.PasswordHash, err = generateHashPass(registerReq.Password)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	createdUser, err := s.userRepo.Create(ctx, ent)
	if err != nil {
		return nil, fmt.Errorf("%s: to create user  %w", op, err)
	}

	return &RegisterOut{User: *createdUser}, nil
}

func (s *AuthService) Login(ctx context.Context, loginReq LoginIn) (string, error) {
	const op = "auth.Login"

	userFound, err := s.userRepo.Get(ctx, user.User{Email: loginReq.Email})
	if err != nil {
		return "", fmt.Errorf("%s: failed to get user: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(userFound.PasswordHash), []byte(loginReq.Password))
	if err != nil {
		return "", ErrInvalidEmailOrPassword
	}

	token, err := s.jwtService.SignJwt(security.Claims{
		Role: userFound.Role,
	})
	if err != nil {
		return "", fmt.Errorf("%s: failed to generate token: %w", op, err)
	}

	return token, nil
}

func generateHashPass(reqPass string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(reqPass), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate password hash: %w", err)
	}
	return string(hashBytes), nil
}
