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

	cachedPrivateSigningKey *rsa.PrivateKey
	cachedPublicKey         *rsa.PublicKey
}

func (k *KeySet) heat() error {
	var err error
	k.cachedPrivateSigningKey, err = k.privateSigningKey()
	if err != nil {
		return fmt.Errorf("failed to cache the parsed signing key: %w", err)
	}

	k.cachedPublicKey, err = k.publicKey()
	if err != nil {
		return fmt.Errorf("failed to cache the parsed public key: %w", err)
	}

	return nil
}

func (k KeySet) active() bool {
	return !k.Revoked && time.Now().Before(k.Expiry)
}

func (k *KeySet) publicKey() (*rsa.PublicKey, error) {
	if k.cachedPublicKey != nil {
		return k.cachedPublicKey, nil
	}

	block, _ := pem.Decode([]byte(k.PublicKey))
	if block == nil {
		return nil, errors.New("could not decode public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPublicKey, err
}

func (k KeySet) privateSigningKey() (*rsa.PrivateKey, error) {
	if k.cachedPrivateSigningKey != nil {
		return k.cachedPrivateSigningKey, nil
	}

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
