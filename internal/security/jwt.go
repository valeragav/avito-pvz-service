package security

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrUnknownPublisher = errors.New("unknown token publisher")
)

type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type JwtService struct {
	iss            string
	accessLifeTime time.Duration
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
}

func New(
	privatePemFile, publicPemFile, iss string,
	accessLifetime time.Duration,
) (*JwtService, error) {
	const op = "security.jwt.New"

	if iss == "" {
		return nil, fmt.Errorf("%s: issuer must not be empty", op)
	}
	if accessLifetime <= 0 {
		return nil, fmt.Errorf("%s: access lifetime must be positive", op)
	}

	privateKey, err := parsePrivateKey(privatePemFile)
	if err != nil {
		return nil, err
	}

	publicKey, err := parsePublicKey(publicPemFile)
	if err != nil {
		return nil, err
	}

	return &JwtService{
		iss:            iss,
		accessLifeTime: accessLifetime,
		privateKey:     privateKey,
		publicKey:      publicKey,
	}, nil
}

func (j JwtService) SignJwt(userClaims domain.UserClaims) (string, error) {
	userClaimsStruct := claims{
		Role: string(userClaims.Role),
	}

	userClaimsStruct.RegisteredClaims = j.registeredClaims()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, userClaimsStruct)

	signedToken, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (j JwtService) registeredClaims() jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessLifeTime)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    j.iss,
	}
}

func (j JwtService) ValidateJwt(incomingToken string) (*domain.UserClaims, error) {
	claims := &claims{}
	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	}

	token, err := jwt.ParseWithClaims(incomingToken, claims, keyFunc)

	if err != nil || token == nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	if claims.Issuer != j.iss {
		return nil, ErrUnknownPublisher
	}

	return &domain.UserClaims{
		Role: domain.Role(claims.Role),
	}, nil
}

func parsePrivateKey(path string) (*rsa.PrivateKey, error) {
	const op = "security.jwt.parsePrivateKey"

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return key, nil
}

func parsePublicKey(path string) (*rsa.PublicKey, error) {
	const op = "security.jwt.parsePublicKey"

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return key, nil
}
