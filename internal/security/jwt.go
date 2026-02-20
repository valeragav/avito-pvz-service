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

type claims struct {
	Role string
	jwt.RegisteredClaims
}

type JwtService struct {
	Iss            string
	AccessLifeTime time.Duration

	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func New(
	privatePemFile, publicPemFile, iss string,
	accessLifetime time.Duration,
) (*JwtService, error) {
	const op = "security.jwt.New"

	privPEM, err := os.ReadFile(privatePemFile)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read private key file: %w", op, err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privPEM)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse private key: %w", op, err)
	}

	pubPEM, err := os.ReadFile(publicPemFile)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read public key file: %w", op, err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pubPEM)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse public key: %w", op, err)
	}

	return &JwtService{
		Iss:            iss,
		AccessLifeTime: accessLifetime,
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
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.AccessLifeTime)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    j.Iss,
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
		return nil, errors.New("invalid token")
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Issuer != j.Iss {
		return nil, errors.New("unknown token publisher")
	}

	return &domain.UserClaims{
		Role: domain.Role(claims.Role),
	}, nil
}
