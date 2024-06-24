package items

import "github.com/suessflorian/pedlar/sales/pkg/keys"

type Item struct {
	ID *keys.OpaqueID
	Details

	Children []*Item
}

type Details struct {
	Name        string
	Description string
	UnitScale   UnitScale
}

type UnitScale string

const (
	Unit UnitScale = "unit" // DEFAULT unit scale value of all created items
)
