package keys

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetCodec(t *testing.T) {
	var checkCodec = func(t *testing.T, oid *OpaqueID, expected EncoderDecoder) {
		t.Helper()
		if oid == nil {
			t.Fatalf("OpaqueID is nil")
		}
		if !reflect.DeepEqual(oid.codec, expected) {
			t.Fatalf("Expected codec %v, but got %v", expected, oid.codec)
		}
	}

	var codec mockCodec

	t.Run("simple", func(t *testing.T) {
		type example struct {
			ID *OpaqueID
		}
		ex := &example{ID: &OpaqueID{ID: 1, external: "1"}}
		SetCodec(ex, &codec)

		checkCodec(t, ex.ID, &codec)
		assert.Equal(t, ex.ID.ID, 1)
		assert.Equal(t, ex.ID.external, "1")
	})

	t.Run("skips unexported simple", func(t *testing.T) {
		type example struct {
			id *OpaqueID
		}
		ex := &example{id: &OpaqueID{ID: 1, external: "1"}}
		SetCodec(ex, &codec)

		assert.Nil(t, ex.id.codec)
		assert.Equal(t, ex.id.ID, 1)
		assert.Equal(t, ex.id.external, "1")
	})

	t.Run("nested", func(t *testing.T) {
		type example struct {
			ID   *OpaqueID
			Nest struct {
				ID *OpaqueID
			}
		}
		ex := &example{
			ID: &OpaqueID{ID: 1, external: "1"},
			Nest: struct {
				ID *OpaqueID
			}{ID: &OpaqueID{ID: 2, external: "2"}},
		}

		SetCodec(ex, &codec)

		checkCodec(t, ex.ID, &codec)
		checkCodec(t, ex.Nest.ID, &codec)

		assert.Equal(t, ex.ID.ID, 1)
		assert.Equal(t, ex.Nest.ID.ID, 2)
		assert.Equal(t, ex.ID.external, "1")
		assert.Equal(t, ex.Nest.ID.external, "2")
	})

	t.Run("nil", func(t *testing.T) {
		type example struct {
			ID *OpaqueID
		}
		ex := &example{ID: nil}

		SetCodec(ex, &codec)

		assert.Nilf(t, ex.ID, "expected id to remain nil")
	})

	t.Run("no opaque id's present", func(t *testing.T) {
		type example struct {
			Val int
		}
		ex := &example{Val: 42}

		SetCodec(ex, &codec)

		assert.Equalf(t, 42, ex.Val, "expected val to remain 42")
	})

	t.Run("multiple", func(t *testing.T) {
		type example struct {
			ID        *OpaqueID
			ProductID *OpaqueID
		}
		ex := &example{
			ID:        &OpaqueID{ID: 1, external: "1"},
			ProductID: &OpaqueID{ID: 2, external: "2"},
		}
		SetCodec(ex, &codec)

		checkCodec(t, ex.ID, &codec)
		checkCodec(t, ex.ProductID, &codec)

		assert.Equal(t, ex.ID.ID, 1)
		assert.Equal(t, ex.ProductID.ID, 2)

		assert.Equal(t, ex.ID.external, "1")
		assert.Equal(t, ex.ProductID.external, "2")
	})

	t.Run("multiple nested and unexported", func(t *testing.T) {
		type nestnest struct {
			id *OpaqueID
		}
		type nest struct {
			NestNest nestnest
		}
		type example struct {
			Sub nest
			ID  *OpaqueID
		}
		ex := &example{
			ID: &OpaqueID{ID: 2, external: "2"},
			Sub: nest{
				NestNest: nestnest{
					id: &OpaqueID{ID: 3, external: "3"},
				},
			},
		}
		SetCodec(ex, &codec)

		checkCodec(t, ex.ID, &codec)
		assert.Nil(t, ex.Sub.NestNest.id.codec)

		assert.Equal(t, ex.ID.ID, 2)
		assert.Equal(t, ex.Sub.NestNest.id.ID, 3)

		assert.Equal(t, ex.ID.external, "2")
		assert.Equal(t, ex.Sub.NestNest.id.external, "3")
	})
}

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
