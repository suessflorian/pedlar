package paginate

import "github.com/suessflorian/pedlar/sales/pkg/keys"

type Input struct {
	Cursor *keys.OpaqueID `json:"cursor"`
	Limit  int            `json:"limit"`
}
