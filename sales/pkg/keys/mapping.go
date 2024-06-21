package keys

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt"
)

// Encode takes in a internal serial id of an object and returns the current external facing ID.
// Implements the EncoderDecoder interface.
func (k *Holder) Encode(ctx context.Context, internalID int) (string, error) {
	go k.sync()

	key, err := k.holding(ctx)
	if err != nil {
		return "", fmt.Errorf("could not retrieve latest keyset: %w", err)
	}

	encrypted, err := k.encrypt(strconv.Itoa(internalID))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt object with internal id %d: %w", internalID, err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"kid":         key.ID,
		"internal_id": encrypted,
		"exp":         key.Expiry.Unix(),
	})

	block, _ := pem.Decode([]byte(key.SigningKey))
	if block == nil {
		return "", errors.New("could not decode signing key")
	}

	signingKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("could not parse signing key: %w", err)
	}

	id, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("could not sign token: %w", err)
	}

	return id, nil
}

// Decode takes an externalID and returns the internal facing ID.
func (k *Holder) Decode(ctx context.Context, externalID string) (int, error) {
	if k.revoke {
		return -1, ErrHolderRevoked
	}
	go k.sync()

	token, err := jwt.Parse(externalID, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		key, err := k.holding(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve latest keyset: %w", err)
		}

		block, _ := pem.Decode([]byte(key.PublicKey))
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

		return rsaPublicKey, nil
	})
	if err != nil {
		return -1, fmt.Errorf("could not parse token: %w", err)
	}

	if !token.Valid {
		return -1, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return -1, errors.New("could not parse claims")
	}

	encryptedID, ok := claims["internal_id"].(string)
	if !ok {
		return -1, errors.New("internal_id claim not found in token")
	}

	internalIDStr, err := k.decrypt(encryptedID)
	if err != nil {
		return -1, fmt.Errorf("could not decrypt internal_id: %w", err)
	}

	internalID, err := strconv.Atoi(internalIDStr)
	if err != nil {
		return -1, fmt.Errorf("could not convert internal_id to int: %w", err)
	}

	return internalID, nil
}