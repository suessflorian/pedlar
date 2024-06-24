package store

import (
	"context"
	"embed"
	"fmt"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed migrations/*.sql
	MIGRATIONS_DIR embed.FS
)

func Conn(ctx context.Context, defaultURL string, name string) (*pgxpool.Pool, error) {
	name = strings.ToLower(name)

	url, err := url.Parse(defaultURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse default database url: %w", err)
	}

	conn, err := pgx.Connect(ctx, defaultURL)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to default database: %w", err)
	}

	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", name))
	if err != nil && err.Error() != fmt.Sprintf("ERROR: database \"%s\" already exists (SQLSTATE 42P04)", name) {
		return nil, fmt.Errorf("failed to ensure database exists: %w", err)
	}

	if err := conn.Close(ctx); err != nil {
		return nil, fmt.Errorf("failed to close connection to default database: %w", err)
	}

	url.Path = "/" + name

	migrations, err := iofs.New(MIGRATIONS_DIR, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to get sub filesystem: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", migrations, url.String())
	if err != nil {
		return nil, fmt.Errorf("failed to establish migration connection: %w", err)
	}

	err = m.Up()
	if err != migrate.ErrNoChange && err != nil {
		return nil, fmt.Errorf("failed to perform migrations: %w", err)
	}

	config, err := pgxpool.ParseConfig(url.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool configuration: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection pool: %w", err)
	}

	return pool, nil
}
