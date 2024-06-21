package keys

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type KeySet struct {
	ID uuid.UUID

	EncryptionKey string
	PrivateKey    string
	PublicKey     string

	Expiry  time.Time
	Revoked bool
}

func (k KeySet) active() bool {
	return !k.Revoked && time.Now().Before(k.Expiry)
}

func (k KeySet) privateSigningKey() (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(k.PrivateKey))
	if block == nil {
		return nil, errors.New("could not decode signing key")
	}

	signingKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse signing key: %w", err)
	}

	return signingKey, err
}
