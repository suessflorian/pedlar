package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suessflorian/pedlar/sales/internal/items"
	"github.com/suessflorian/pedlar/sales/pkg/keys"
	"github.com/suessflorian/pedlar/sales/pkg/model/paginate"
)

type Items struct {
	Conn *pgxpool.Pool
}

func (i *Items) CreateItem(ctx context.Context, deets items.Details) (*items.Item, error) {
	var assigned int
	err := i.Conn.QueryRow(ctx, `INSERT INTO items (name, description, unit_scale) VALUES ($1, $2, $3) RETURNING id`, deets.Name, deets.Description, deets.UnitScale).Scan(&assigned)
	if err != nil {
		return nil, fmt.Errorf("failed to insert into items: %w", err)
	}

	return &items.Item{
		ID: &keys.OpaqueID{
			ID: assigned,
		},
		Details: deets,
	}, nil
}

func (i *Items) UpdateItemDetails(ctx context.Context, id *keys.OpaqueID, deets items.Details) error {
	_, err := i.Conn.Exec(ctx, `UPDATE items SET name = $1, description = $2, unit_scale = $3 WHERE id = $4`, deets.Name, deets.Description, deets.UnitScale, id.ID)
	return err
}

func (i *Items) GetItem(ctx context.Context, id int) (*items.Item, error) {
	var (
		name        string
		description string
		scale       items.UnitScale
	)
	err := i.Conn.QueryRow(ctx, `SELECT name, description, unit_scale FROM items WHERE id = $1`, id).Scan(&name, &description, &scale)
	if err != nil {
		return nil, fmt.Errorf("failed to select from items: %w", err)
	}

	return &items.Item{
		ID: &keys.OpaqueID{ID: id},
		Details: items.Details{
			Name:        name,
			Description: description,
			UnitScale:   scale,
		},
	}, nil
}

func (i *Items) GetItems(ctx context.Context, ids ...int) ([]*items.Item, error) {
	rows, err := i.Conn.Query(ctx, `SELECT id, name, description, unit_scale FROM items WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to select many from items: %w", err)
	}

	var results = make([]*items.Item, 0, len(ids))
	for rows.Next() {
		var (
			id          keys.OpaqueID
			name        string
			description string
			scale       items.UnitScale
		)
		err := rows.Scan(&id, &name, &description, &scale)
		if err != nil {
			return nil, fmt.Errorf("failed to scan from rows result when selecting from items: %w", err)
		}
		results = append(results, &items.Item{
			ID: &id,
			Details: items.Details{
				Name:        name,
				Description: description,
				UnitScale:   scale,
			},
		})
	}

	return results, nil
}

func (i *Items) PageItems(ctx context.Context, page paginate.Paginate) ([]*keys.OpaqueID, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if page.Cursor == nil {
		rows, err = i.Conn.Query(ctx, `SELECT id FROM items LIMIT $1`, page.Limit)
		if err != nil {
			return nil, fmt.Errorf("failed to select a page from items: %w", err)
		}
	} else {
		rows, err = i.Conn.Query(ctx, `SELECT id FROM items WHERE id > $1 LIMIT $2`, page.Cursor, page.Limit)
		if err != nil {
			return nil, fmt.Errorf("failed to select a page from items with cursor condition: %w", err)
		}
	}

	var results = make([]*keys.OpaqueID, 0, page.Limit)
	for rows.Next() {
		var (
			id keys.OpaqueID
		)
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to scan from rows result when selecting from items: %w", err)
		}
		results = append(results, &id)
	}

	return results, nil
}
