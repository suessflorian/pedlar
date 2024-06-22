// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"github.com/suessflorian/pedlar/sales/pkg/keys"
)

type Mutation struct {
}

type NewSale struct {
	Items []*NewSaleLineItemSale `json:"items"`
}

type NewSaleLineItemSale struct {
	ProductID keys.OpaqueID `json:"productID"`
	Quantity  int           `json:"quantity"`
	UnitPrice int           `json:"unit_price"`
}

type Query struct {
}
