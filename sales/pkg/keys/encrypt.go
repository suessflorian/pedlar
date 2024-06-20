package keys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// encrypt symmetrically encrypts target strings using the current keyset's encryption key.
// We adopt AES-GCM encryption scheme.
func (k *Holder) encrypt(targets ...string) ([]string, error) {
	key, err := base64.StdEncoding.DecodeString(k.curr.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode encryption key: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not create cipher block: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("could not create GCM mode: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("could not generate nonce: %w", err)
	}

	var encrypted = make([]string, 0, len(targets))
	for _, target := range targets {
		ciphertext := aesGCM.Seal(nonce, nonce, []byte(target), nil)
		encrypted = append(encrypted, base64.StdEncoding.EncodeToString(ciphertext))
	}

	return encrypted, nil
}

// decrypt decrypts target strings using the current keyset's encryption key.
// We adopt AES-GCM encryption scheme. Assumes nonce is transmitted.
func (k *Holder) decrypt(targets ...string) ([]string, error) {
	key, err := base64.StdEncoding.DecodeString(k.curr.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not create cipher block: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("could not create GCM mode: %w", err)
	}

	var decrypted = make([]string, 0, len(targets))
	for _, target := range targets {
		ciphertext, err := base64.StdEncoding.DecodeString(target)
		if err != nil {
			return nil, fmt.Errorf("could not decode base64 string: %w", err)
		}

		nonceSize := aesGCM.NonceSize()
		if len(ciphertext) < nonceSize {
			return nil, fmt.Errorf("ciphertext too short")
		}

		nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
		plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return nil, fmt.Errorf("could not decrypt data: %w", err)
		}

		decrypted = append(decrypted, string(plaintext))
	}

	return decrypted, nil
}
