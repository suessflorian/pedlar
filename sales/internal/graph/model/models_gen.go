// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"github.com/suessflorian/pedlar/sales/internal/items"
)

type ConfirmCreateItem struct {
	Similar []*items.Item  `json:"similar"`
	Details *items.Details `json:"details"`
	Confirm *items.Item    `json:"confirm"`
}

type Mutation struct {
}

type Query struct {
}
