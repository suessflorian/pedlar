package sale

import (
	"context"

	"github.com/suessflorian/pedlar/sales/pkg/model"
)

type Sale struct {
	ID        int
	LineItems []LineItem
}

func (s *Sale) External(ctx context.Context, obfuscate model.ExternelIDFunc) *ExternalSale {
	ext := model.Must(obfuscate)
	var lines = make([]*ExternalLineItem, 0, len(s.LineItems))
	for _, line := range s.LineItems {
		lines = append(lines, line.External(ctx, obfuscate))
	}

	return &ExternalSale{
		ID:        ext(ctx, "sale", s.ID),
		LineItems: lines,
	}
}

type LineItem struct {
	ID        int
	ProductID int
	Quantity  int
	UnitPrice int
}

func (s *LineItem) External(ctx context.Context, obfuscate model.ExternelIDFunc) *ExternalLineItem {
	ext := model.Must(obfuscate)
	return &ExternalLineItem{
		ID:        ext(ctx, "line_item", s.ID),
		ProductID: ext(ctx, "product", s.ProductID),
		Quantity:  s.Quantity,
		UnitPrice: s.UnitPrice,
	}
}
