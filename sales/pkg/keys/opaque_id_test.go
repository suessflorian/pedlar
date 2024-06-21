package keys

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgxIntegration(t *testing.T) {
	t.Parallel()

	err := godotenv.Load()
	require.NoError(t, err)

	url := os.Getenv("TEST_DATABASE_URL")
	require.NotEmpty(t, url)

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	testTable := "test_" + t.Name()

	ctx := context.Background()
	_, err = conn.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %q", testTable))
	require.NoError(t, err)

	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE TABLE %q (id integer)", testTable))
	require.NoError(t, err)

	oid := OpaqueID{ID: 42}
	_, err = conn.Exec(ctx, fmt.Sprintf("INSERT INTO %q (id) VALUES ($1)", testTable), oid.ID)
	require.NoError(t, err)

	var id int
	err = conn.QueryRow(ctx, fmt.Sprintf("SELECT id FROM %q WHERE id = $1", testTable), oid.ID).Scan(&id)
	require.NoError(t, err)
	require.Equal(t, oid.ID, id)

	var newOid OpaqueID
	err = conn.QueryRow(ctx, fmt.Sprintf("SELECT id FROM %q WHERE id = $1", testTable), oid.ID).Scan(&newOid.ID)
	require.NoError(t, err)
	require.Equal(t, oid.ID, newOid.ID)
}

// mockCodec implements the EncoderDecoder
type mockCodec struct{}

func (m *mockCodec) Encode(ctx context.Context, id int) (string, error) {
	return "encoded-" + fmt.Sprintf("%d", id), nil
}

func (m *mockCodec) Decode(ctx context.Context, id string) (int, error) {
	if strings.HasPrefix(id, "encoded-") {
		stringID := strings.TrimPrefix(id, "encoded-")
		return strconv.Atoi(stringID)
	}
	return -1, fmt.Errorf("invalid encoded string")
}

func TestGQLGenIntegration(t *testing.T) {
	t.Parallel()

	t.Run("unmarshalling", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		var id OpaqueID
		err := id.UnmarshalGQLContext(ctx, "test-12")
		require.NoError(t, err)
		assert.Equal(t, "test-12", id.external)

		err = id.UnmarshalGQLContext(ctx, 123)
		require.Error(t, err)
	})

	t.Run("marshalling", func(t *testing.T) {
		ctx := context.Background()
		id := OpaqueID{
			ID:    42,
			codec: new(mockCodec),
		}

		var buf bytes.Buffer
		err := id.MarshalGQLContext(ctx, &buf)
		require.NoError(t, err)

		assert.Equal(t, `"encoded-42"`, buf.String())

		id.codec = nil
		err = id.MarshalGQLContext(ctx, &buf)
		require.Error(t, err)
	})
}
