// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type Mutation struct {
}

type NewSale struct {
	Items []*NewSaleLineItemSale `json:"items"`
}

type NewSaleLineItemSale struct {
	ProductID string `json:"productID"`
	Quantity  int    `json:"quantity"`
	UnitPrice int    `json:"unit_price"`
}

type PaginationInput struct {
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
}

type Query struct {
}
