package keys

import (
	"math/rand"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	t.Parallel()

	t.Run("encrypt to many to decrypt to one", func(t *testing.T) {
		t.Parallel()

		key, err := generateAESKey()
		require.NoError(t, err)

		curr := &KeySet{EncryptionKey: key}
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

func FuzzEncryptDecrypt(f *testing.F) {
	f.Fuzz(func(t *testing.T, expected string) {
		key, err := generateAESKey()
		require.NoError(t, err)

		curr := &KeySet{EncryptionKey: key}
		holder := Holder{curr: curr}

		cifers, err := holder.encrypt(expected)
		require.NoError(t, err)

		actual, err := holder.decrypt(string(cifers))
		require.NoError(t, err)

		assert.Equal(t, expected, actual)
	})
}

func BenchmarkEncrypt(b *testing.B) {
	key, err := generateAESKey()
	if err != nil {
		b.Fatalf("failed to generate AES key: %v", err)
	}

	curr := &KeySet{EncryptionKey: key}
	holder := Holder{curr: curr}

	trials := make([]string, b.N)
	for i := range trials {
		trials[i] = strconv.Itoa(rand.Int())
	}

	b.ResetTimer()
	for _, trial := range trials {
		_, err := holder.encrypt(trial)
		if err != nil {
			b.Fatalf("failed to encrypt message: %v", err)
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key, err := generateAESKey()
	if err != nil {
		b.Fatalf("failed to generate AES key: %v", err)
	}

	curr := &KeySet{EncryptionKey: key}
	holder := Holder{curr: curr}

	trials := make([]string, b.N)
	for i := range trials {
		trials[i], err = holder.encrypt(strconv.Itoa(rand.Int()))
		if err != nil {
			b.Fatalf("failed to create encrypted trial value: %v", err)
		}
	}

	b.ResetTimer()
	for _, trial := range trials {
		_, err := holder.decrypt(trial)
		if err != nil {
			b.Fatalf("failed to encrypt message: %v", err)
		}
	}
}
