package keys

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	t.Parallel()

	t.Run("symmetric", func(t *testing.T) {
		t.Parallel()

		key, err := generateAESKey()
		require.NoError(t, err)

		curr := KeySet{EncryptionKey: key}
		holder := Holder{curr: curr}

		var trials = []string{"hello world", "🔥", "this encryption/decryption stuff such an overkill"}
		for _, expected := range trials {
			cifers, err := holder.encrypt(expected)
			require.NoError(t, err)

			actual, err := holder.decrypt(string(cifers))
			require.NoError(t, err)

			assert.Equal(t, expected, actual)
		}
	})

	t.Run("encrypt to many to decrypt to one", func(t *testing.T) {
		t.Parallel()

		key, err := generateAESKey()
		require.NoError(t, err)

		curr := KeySet{EncryptionKey: key}
		holder := Holder{curr: curr}

		expected := "hello world"
		repeats := 10000

		var cifers = make([]string, 0, repeats)

		for i := 0; i < repeats; i++ {
			cifer, err := holder.encrypt(expected)
			require.NoError(t, err)
			cifers = append(cifers, cifer)
		}

		require.Len(t, cifers, repeats)
		slices.Sort(cifers)
		cifers = slices.Compact(cifers)
		require.Len(t, cifers, repeats)

		var actuals = make([]string, 0, repeats)
		for _, cifer := range cifers {
			actual, err := holder.decrypt(cifer)
			require.NoError(t, err)
			actuals = append(actuals, actual)
		}

		actuals = slices.Compact(actuals)
		require.Len(t, actuals, 1)
	})
}

func BenchmarkEncrypt(b *testing.B) {
	key, err := generateAESKey()
	if err != nil {
		b.Fatalf("failed to generate AES key: %v", err)
	}

	curr := KeySet{EncryptionKey: key}
	holder := Holder{curr: curr}

	message := "hello world"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := holder.encrypt(message)
		if err != nil {
			b.Fatalf("failed to encrypt message: %v", err)
		}
	}
}
