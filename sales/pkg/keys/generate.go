package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// aesKeySize here must be either 16, 24 or 32 bytes by AES.
// For us that then corresponds to respectively AES-128-GCM, AES-192-GCM, AES-256-GCM encryption scheme.
const aesKeySize = 16

// generateAESKey generates a new AES key of the given size.
func generateAESKey() (string, error) {
	bytes := make([]byte, aesKeySize)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("could not generate random bytes: %w", err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// generateRSAKeyPair generates a new RSA key pair.
func generateRSAKeyPair() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("could not generate RSA private key: %w", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("could not marshal RSA public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(privateKeyPEM), string(publicKeyPEM), nil
}
