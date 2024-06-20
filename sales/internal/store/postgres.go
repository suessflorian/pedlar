package store

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
)

const (
	MIGRATIONS_DIR = "file://db/migrations"
)

// TODO: connection pool
func Conn(ctx context.Context, url string) (*pgx.Conn, error) {
	m, err := migrate.New(MIGRATIONS_DIR, url)
	if err != nil {
		return nil, fmt.Errorf("failed to establish migration connection: %w", err)
	}

	err = m.Up()
	if err != migrate.ErrNoChange && err != nil {
		return nil, fmt.Errorf("failed to perform migrations: %w", err)
	}

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to establish database connection: %w", err)
	}

	return conn, nil
}
