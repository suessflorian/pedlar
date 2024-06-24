package items

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/suessflorian/pedlar/sales/pkg/keys"
	"github.com/suessflorian/pedlar/sales/pkg/model/paginate"
)

type store interface {
	CreateItem(context.Context, Details) (*Item, error)
	UpdateItemDetails(context.Context, *keys.OpaqueID, Details) error

	GetItem(context.Context, int) (*Item, error)
	GetItems(context.Context, ...int) ([]*Item, error)

	PageItems(context.Context, paginate.Paginate) ([]*keys.OpaqueID, error)
	// TODO: SearchItems(context.Context, search) ([]*Item, error)
}

type ItemManager struct {
	Store store
}

func (i *ItemManager) GetItem(ctx context.Context, externalID *keys.OpaqueID) (*Item, error) {
	id, err := externalID.Decode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to decode id: %w", err)
	}

	return i.Store.GetItem(ctx, id)
}

func (i *ItemManager) GetItems(ctx context.Context, externalIDs ...*keys.OpaqueID) ([]*Item, error) {
	var ids = make([]int, 0, len(externalIDs))
	for _, externalID := range externalIDs {
		id, err := externalID.Decode(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to decode id: %w", err)
		}
		ids = append(ids, id)
	}
	return i.Store.GetItems(ctx, ids...)
}

type ItemSearch struct {
	Page *paginate.Paginate
	// TODO: embed a search type
}

func (i *ItemManager) SearchItems(ctx context.Context, search *ItemSearch) ([]*Item, error) {
	// TODO: if search pattern found, route request to SearchItems instead

	if search.Page == nil {
		search.Page = &paginate.Paginate{
			Cursor: nil,
			Limit:  20,
		}
	}

	items, err := i.Store.PageItems(ctx, *search.Page)
	if err != nil {
		return nil, fmt.Errorf("failed to page through items: %w", err)
	}

	return i.GetItems(ctx, items...)
}

var (
	ErrNoNameItem = errors.New("item must have a name")
)

func (i *ItemManager) CreateItem(ctx context.Context, deets Details) (*Item, error) {
	if strings.TrimSpace(deets.Name) == "" {
		return nil, ErrNoNameItem
	}

	item, err := i.Store.CreateItem(ctx, deets)
	if err != nil {
		return nil, fmt.Errorf("failed to create new item from details provided: %w", err)
	}

	return item, nil
}

func (i *ItemManager) UpdateItemDetails(ctx context.Context, id *keys.OpaqueID, deets Details) (*Item, error) {
	if strings.TrimSpace(deets.Name) == "" {
		return nil, ErrNoNameItem
	}

	err := i.Store.UpdateItemDetails(ctx, id, deets)
	if err != nil {
		return nil, fmt.Errorf("failed to create new item from details provided: %w", err)
	}

	return i.GetItem(ctx, id)
}
