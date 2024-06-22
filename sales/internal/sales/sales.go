package sales

import (
	"context"

	"github.com/suessflorian/pedlar/sales/pkg/keys"
)

type store interface {
	CreateSale(context.Context, Sale)
}

type Sale struct {
	ID        *keys.OpaqueID
	LineItems []LineItem
}

type LineItem struct {
	ID        *keys.OpaqueID
	ProductID *keys.OpaqueID
	Quantity  int
	UnitPrice int
}
