package store

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestCyclesDisallowedOnItemRelationships(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	err := godotenv.Load()
	require.NoError(t, err)

	url := os.Getenv("TEST_DATABASE_URL")
	require.NotEmpty(t, url)

	conn, err := Conn(ctx, url, t.Name())
	require.NoError(t, err)
	defer conn.Close()

	var (
		first  int
		second int
	)
	err = conn.QueryRow(ctx, `INSERT INTO items (name) VALUES ('some_item') RETURNING id`).Scan(&first)
	require.NoError(t, err)

	err = conn.QueryRow(ctx, `INSERT INTO items (name) VALUES ('another_item') RETURNING id`).Scan(&second)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `INSERT INTO item_relationships (parent_id, child_id) VALUES ($1, $2)`, first, second)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `INSERT INTO item_relationships (parent_id, child_id) VALUES ($1, $2)`, second, first)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cyclical relationship detected")
}
