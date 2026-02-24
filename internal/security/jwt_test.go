package security_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/security"
)

func generateRSAKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return privateKey, &privateKey.PublicKey
}

func encodePrivateKeyToPEM(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

// MarshalPKIXPublicKey + "PUBLIC KEY" — именно этот формат ожидает jwt библиотека
func encodePublicKeyToPEM(t *testing.T, key *rsa.PublicKey) []byte {
	t.Helper()

	der, err := x509.MarshalPKIXPublicKey(key)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})
}

// writeKeyFiles записывает PEM-данные во временные файлы.
// Если передать nil, файл не создаётся — путь будет указывать в никуда.
func writeKeyFiles(t *testing.T, privPEM, pubPEM []byte) (privPath, pubPath string) {
	t.Helper()

	dir := t.TempDir()
	privPath = filepath.Join(dir, "private.pem")
	pubPath = filepath.Join(dir, "public.pem")

	if privPEM != nil {
		require.NoError(t, os.WriteFile(privPath, privPEM, 0o600))
	}
	if pubPEM != nil {
		require.NoError(t, os.WriteFile(pubPath, pubPEM, 0o600))
	}

	return privPath, pubPath
}

// newService — быстрая сборка сервиса с валидными ключами.
func newServiceFromKeys(t *testing.T, priv *rsa.PrivateKey, pub *rsa.PublicKey, iss string) *security.JwtService {
	t.Helper()

	privPath, pubPath := writeKeyFiles(
		t,
		encodePrivateKeyToPEM(priv),
		encodePublicKeyToPEM(t, pub),
	)

	svc, err := security.New(privPath, pubPath, iss, time.Hour)
	require.NoError(t, err)

	return svc
}

func newService(t *testing.T, iss string) *security.JwtService {
	t.Helper()

	priv, pub := generateRSAKeyPair(t)
	privPath, pubPath := writeKeyFiles(
		t,
		encodePrivateKeyToPEM(priv),
		encodePublicKeyToPEM(t, pub),
	)

	svc, err := security.New(privPath, pubPath, iss, time.Hour)
	require.NoError(t, err)

	return svc
}
func TestNew(t *testing.T) {
	t.Parallel()

	priv, pub := generateRSAKeyPair(t)
	privPEM := encodePrivateKeyToPEM(priv)
	pubPEM := encodePublicKeyToPEM(t, pub)

	tests := []struct {
		name        string
		setupPaths  func(t *testing.T) (string, string)
		iss         string
		lifetime    time.Duration
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupPaths: func(t *testing.T) (string, string) {
				return writeKeyFiles(t, privPEM, pubPEM)
			},
			iss:      "test-issuer",
			lifetime: time.Hour,
			wantErr:  false,
		},
		{
			name: "empty issuer",
			setupPaths: func(t *testing.T) (string, string) {
				return writeKeyFiles(t, privPEM, pubPEM)
			},
			iss:         "",
			lifetime:    time.Hour,
			wantErr:     true,
			errContains: "issuer must not be empty",
		},
		{
			name: "zero lifetime",
			setupPaths: func(t *testing.T) (string, string) {
				return writeKeyFiles(t, privPEM, pubPEM)
			},
			iss:         "issuer",
			lifetime:    0,
			wantErr:     true,
			errContains: "access lifetime must be positive",
		},
		{
			name: "negative lifetime",
			setupPaths: func(t *testing.T) (string, string) {
				return writeKeyFiles(t, privPEM, pubPEM)
			},
			iss:         "issuer",
			lifetime:    -time.Second,
			wantErr:     true,
			errContains: "access lifetime must be positive",
		},
		{
			name: "missing private key file",
			setupPaths: func(t *testing.T) (string, string) {
				// pubPEM есть, privPEM нет — путь к приватному ключу несуществующий
				_, pubPath := writeKeyFiles(t, nil, pubPEM)
				return "/nonexistent/private.pem", pubPath
			},
			iss:         "issuer",
			lifetime:    time.Hour,
			wantErr:     true,
			errContains: "parsePrivateKey",
		},
		{
			name: "missing public key file",
			setupPaths: func(t *testing.T) (string, string) {
				privPath, _ := writeKeyFiles(t, privPEM, nil)
				return privPath, "/nonexistent/public.pem"
			},
			iss:         "issuer",
			lifetime:    time.Hour,
			wantErr:     true,
			errContains: "parsePublicKey",
		},
		{
			name: "corrupt private key content",
			setupPaths: func(t *testing.T) (string, string) {
				return writeKeyFiles(t, []byte("not-a-pem"), pubPEM)
			},
			iss:         "issuer",
			lifetime:    time.Hour,
			wantErr:     true,
			errContains: "parsePrivateKey",
		},
		{
			name: "corrupt public key content",
			setupPaths: func(t *testing.T) (string, string) {
				return writeKeyFiles(t, privPEM, []byte("not-a-pem"))
			},
			iss:         "issuer",
			lifetime:    time.Hour,
			wantErr:     true,
			errContains: "parsePublicKey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			privPath, pubPath := tt.setupPaths(t)
			svc, err := security.New(privPath, pubPath, tt.iss, tt.lifetime)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, svc)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, svc)
		})
	}
}

func TestSignJwt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		claims  domain.UserClaims
		wantErr bool
	}{
		{
			name:    "ModeratorRole",
			claims:  domain.UserClaims{Role: domain.ModeratorRole},
			wantErr: false,
		},
		{
			name:    "EmployeeRole",
			claims:  domain.UserClaims{Role: domain.EmployeeRole},
			wantErr: false,
		},
		{
			name:    "empty role",
			claims:  domain.UserClaims{Role: ""},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newService(t, "issuer")
			token, err := svc.SignJwt(tt.claims)

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, token)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, token)
			// JWT всегда состоит из трёх частей
			assert.Len(t, strings.Split(token, "."), 3)
		})
	}
}

func TestSignJwt_DifferentRolesProduceDifferentTokens(t *testing.T) {
	t.Parallel()

	svc := newService(t, "issuer")

	tokenAdmin, err := svc.SignJwt(domain.UserClaims{Role: domain.ModeratorRole})
	require.NoError(t, err)

	tokenUser, err := svc.SignJwt(domain.UserClaims{Role: domain.EmployeeRole})
	require.NoError(t, err)

	assert.NotEqual(t, tokenAdmin, tokenUser)
}

func TestSignJwt_ContainsExpectedClaims(t *testing.T) {
	t.Parallel()

	const iss = "my-service"
	svc := newService(t, iss)

	tokenStr, err := svc.SignJwt(domain.UserClaims{Role: domain.ModeratorRole})
	require.NoError(t, err)

	// Парсим без валидации подписи, чтобы проверить payload
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenStr, &jwt.MapClaims{})
	require.NoError(t, err)

	mc, ok := token.Claims.(*jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, iss, (*mc)["iss"])
	assert.Equal(t, string(domain.ModeratorRole), (*mc)["role"])
	assert.NotNil(t, (*mc)["exp"])
	assert.NotNil(t, (*mc)["iat"])
}

func TestValidateJwt(t *testing.T) {
	t.Parallel()

	const iss = "test-issuer"

	tests := []struct {
		name       string
		buildToken func(svc *security.JwtService) string
		wantErr    bool
		wantErrIs  error
		wantRole   domain.Role
	}{
		{
			name: "valid token",
			buildToken: func(svc *security.JwtService) string {
				tok, err := svc.SignJwt(domain.UserClaims{Role: domain.ModeratorRole})
				require.NoError(t, err)
				return tok
			},
			wantErr:  false,
			wantRole: domain.ModeratorRole,
		},
		{
			name:       "empty string",
			buildToken: func(svc *security.JwtService) string { return "" },
			wantErr:    true,
			wantErrIs:  security.ErrInvalidToken,
		},
		{
			name:       "random garbage",
			buildToken: func(svc *security.JwtService) string { return "aaa.bbb.ccc" },
			wantErr:    true,
			wantErrIs:  security.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newService(t, iss)
			tokenStr := tt.buildToken(svc)

			got, err := svc.ValidateJwt(tokenStr)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				}
				require.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, tt.wantRole, got.Role)
		})
	}
}

func TestValidateJwt_WrongIssuer(t *testing.T) {
	t.Parallel()

	priv, pub := generateRSAKeyPair(t)

	// Оба сервиса используют одну пару ключей, но разные issuer
	signer := newServiceFromKeys(t, priv, pub, "issuer-A")
	validator := newServiceFromKeys(t, priv, pub, "issuer-B")

	tokenStr, err := signer.SignJwt(domain.UserClaims{Role: domain.ModeratorRole})
	require.NoError(t, err)

	got, err := validator.ValidateJwt(tokenStr)

	require.Error(t, err)
	require.ErrorIs(t, err, security.ErrUnknownPublisher)
	require.Nil(t, got)
}

func TestValidateJwt_SignedWithDifferentPrivateKey(t *testing.T) {
	t.Parallel()

	priv1, pub1 := generateRSAKeyPair(t)
	priv2, pub2 := generateRSAKeyPair(t)

	svc1 := newServiceFromKeys(t, priv1, pub1, "issuer")
	svc2 := newServiceFromKeys(t, priv2, pub2, "issuer")

	// svc1 подписывает, svc2 (с другим публичным ключом) валидирует
	tokenStr, err := svc1.SignJwt(domain.UserClaims{Role: domain.ModeratorRole})
	require.NoError(t, err)

	got, err := svc2.ValidateJwt(tokenStr)

	require.Error(t, err)
	require.ErrorIs(t, err, security.ErrInvalidToken)
	require.Nil(t, got)
}

func TestValidateJwt_WrongAlgorithm_HS256(t *testing.T) {
	t.Parallel()

	svc := newService(t, "issuer")

	// Создаём токен с HMAC вместо RSA
	hmacToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "admin",
		"iss":  "issuer",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := hmacToken.SignedString([]byte("secret"))
	require.NoError(t, err)

	got, err := svc.ValidateJwt(tokenStr)

	require.Error(t, err)
	require.ErrorIs(t, err, security.ErrInvalidToken)
	require.Nil(t, got)
}

// TestValidateJwt_RoundTrip проверяет симметрию для всех ролей.
func TestValidateJwt_RoundTrip(t *testing.T) {
	t.Parallel()

	roles := []domain.Role{
		domain.ModeratorRole,
		domain.EmployeeRole,
	}

	svc := newService(t, "issuer")

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			t.Parallel()

			original := domain.UserClaims{Role: role}

			tokenStr, err := svc.SignJwt(original)
			require.NoError(t, err)

			got, err := svc.ValidateJwt(tokenStr)
			require.NoError(t, err)
			require.NotNil(t, got)

			assert.Equal(t, original.Role, got.Role)
		})
	}
}
