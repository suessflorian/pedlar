package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/suessflorian/pedlar/sales/pkg/keys"
)

type Keys struct {
	Conn *pgx.Conn
}

func (k *Keys) GetActiveKeySet(ctx context.Context) (*keys.KeySet, error) {
	var key keys.KeySet
	row := k.Conn.QueryRow(ctx, "SELECT kid, encryption_key, signing_key, public_key, expiry FROM keys WHERE revoked = false AND expiry > NOW() ORDER BY expiry DESC LIMIT 1")

	err := row.Scan(&key.ID, &key.EncryptionKey, &key.PrivateKey, &key.PublicKey, &key.Expiry)
	if err == pgx.ErrNoRows {
		return &keys.KeySet{}, keys.ErrNoActiveKeySet
	} else if err != nil {
		return &keys.KeySet{}, fmt.Errorf("failed to get keyset: %w", err)
	} else {
		return &key, nil
	}
}

func (k *Keys) RevokeKeySet(ctx context.Context, ID uuid.UUID) error {
	_, err := k.Conn.Exec(ctx, "UPDATE keys SET revoked = true WHERE id = $1", ID)
	if err != nil {
		return fmt.Errorf("failed to update keyset to revoked: %w", err)
	}
	return nil
}

func (k *Keys) KeySets(ctx context.Context, IDs ...uuid.UUID) ([]*keys.KeySet, error) {
	rows, err := k.Conn.Query(ctx, "SELECT id, encryption_key, signing_key, public_key, expiry FROM keys WHERE id = ANY($1)", IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve keys: %w", err)
	}

	var results []*keys.KeySet
	for rows.Next() {
		var key keys.KeySet
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("failed to scan in key: %w", err)
		}
		results = append(results, &key)
	}

	return results, nil
}

func (k *Keys) RegisterKeySet(ctx context.Context, signingKey, publicKey, encryptionKey string) (*keys.KeySet, error) {
	expiry := time.Now().Add(7 * 24 * time.Hour)

	var id uuid.UUID
	err := k.Conn.QueryRow(ctx,
		"INSERT INTO keys (signing_key, public_key, encryption_key, expiry) VALUES ($1, $2, $3, $4) RETURNING kid",
		signingKey, publicKey, encryptionKey, expiry).Scan(&id)
	if err != nil {
		return &keys.KeySet{}, fmt.Errorf("failed to insert into keys: %w", err)
	}

	return &keys.KeySet{
		ID:            id,
		EncryptionKey: encryptionKey,
		PrivateKey:    signingKey,
		PublicKey:     publicKey,
		Expiry:        expiry,
		Revoked:       false,
	}, nil
}
