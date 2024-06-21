package keys

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzEncodeDecode(f *testing.F) {
	f.Fuzz(func(t *testing.T, internalID int) {
		ctx := context.Background()

		encryptionKey, err := generateAESKey()
		require.NoError(t, err)

		private, public, err := generateRSAKeyPair()
		require.NoError(t, err)

		holder := Holder{
			store: nil,
			curr: &KeySet{
				ID:            uuid.New(),
				EncryptionKey: encryptionKey,
				PrivateKey:    private,
				PublicKey:     public,
				Expiry:        time.Now().Add(1 * time.Hour),
				Revoked:       false,
			},
			chain:  map[uuid.UUID]*KeySet{},
			poll:   time.Now().Add(1 * time.Hour),
			revoke: false,
		}

		externalID, err := holder.Encode(ctx, internalID)
		require.NoError(t, err)

		decodedID, err := holder.Decode(ctx, externalID)
		require.NoError(t, err)

		assert.Equal(t, internalID, decodedID)
	})
}

func BenchmarkEncode(b *testing.B) {
	ctx := context.Background()

	encryptionKey, err := generateAESKey()
	if err != nil {
		b.Fatalf("failed to generate AES key: %v", err)
	}

	private, public, err := generateRSAKeyPair()
	if err != nil {
		b.Fatalf("failed to generate public/private RSA key pairs: %v", err)
	}

	holder := Holder{
		store: nil,
		curr: &KeySet{
			ID:            uuid.New(),
			EncryptionKey: encryptionKey,
			PrivateKey:    private,
			PublicKey:     public,
			Expiry:        time.Now().Add(1 * time.Hour),
			Revoked:       false,
		},
		chain:  map[uuid.UUID]*KeySet{},
		poll:   time.Now().Add(1 * time.Hour),
		revoke: false,
	}

	err = holder.curr.heat()
	if err != nil {
		b.Fatalf("failed to heat current keyset: %v", err)
	}

	trials := make([]int, b.N)
	for i := range trials {
		trials[i] = rand.Int()
	}

	b.ResetTimer()
	for _, trial := range trials {
		_, err = holder.Encode(ctx, trial)
		if err != nil {
			b.Fatalf("failed to encrypt message: %v", err)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	ctx := context.Background()

	encryptionKey, err := generateAESKey()
	if err != nil {
		b.Fatalf("failed to generate AES key: %v", err)
	}

	private, public, err := generateRSAKeyPair()
	if err != nil {
		b.Fatalf("failed to generate public/private RSA key pairs: %v", err)
	}

	holder := Holder{
		store: nil,
		curr: &KeySet{
			ID:            uuid.New(),
			EncryptionKey: encryptionKey,
			PrivateKey:    private,
			PublicKey:     public,
			Expiry:        time.Now().Add(1 * time.Hour),
			Revoked:       false,
		},
		chain:  map[uuid.UUID]*KeySet{},
		poll:   time.Now().Add(1 * time.Hour),
		revoke: false,
	}

	err = holder.curr.heat()
	if err != nil {
		b.Fatalf("failed to heat current keyset: %v", err)
	}

	trials := make([]string, b.N)
	for i := range trials {
		trials[i], err = holder.Encode(ctx, rand.Int())
		if err != nil {
			b.Fatalf("failed to create encoded trial value: %v", err)
		}
	}

	b.ResetTimer()
	for _, trial := range trials {
		_, err = holder.Decode(ctx, trial)
		if err != nil {
			b.Fatalf("failed to encrypt message: %v", err)
		}
	}
}
