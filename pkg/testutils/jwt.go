package testutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
)

func GenerateTestJWTKeys() (privatePath, publicPath string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	dir, err := os.MkdirTemp("", "jwt-test-keys")
	if err != nil {
		return "", "", err
	}

	privatePath = filepath.Join(dir, "private.pem")
	publicPath = filepath.Join(dir, "public.pem")

	privBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	err = os.WriteFile(privatePath, privBytes, 0o600)
	if err != nil {
		return "", "", err
	}

	derPub, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPub,
	})
	err = os.WriteFile(publicPath, pubBytes, 0o600)
	if err != nil {
		return "", "", err
	}

	return privatePath, publicPath, nil
}
