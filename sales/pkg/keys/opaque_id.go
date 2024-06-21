package keys

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
)

type EncoderDecoder interface {
	Encode(ctx context.Context, id int) (string, error)
	Decode(ctx context.Context, id string) (int, error)
}

// OpaqueID is the type that is used as an integration tool to any application that wants to
// use the Holder service. It satisfies marshaller interfaces for API.
type OpaqueID struct {
	ID       int
	external string

	codec EncoderDecoder
}

// WithCodec sets the codec for the key. It is nessecary for encoding an internal id
// or decoding an external id.
func (k *OpaqueID) WithCodec(c EncoderDecoder) *OpaqueID {
	return &OpaqueID{
		ID:    k.ID,
		codec: c,
	}
}

func (k *OpaqueID) Decode(ctx context.Context) (int, error) {
	if k.codec == nil {
		return -1, fmt.Errorf("no codec set for HiddenID")
	}

	if k.external == "" {
		return -1, fmt.Errorf("no stored external id within HiddenID to decode")
	}

	return k.codec.Decode(ctx, k.external)
}

func (k *OpaqueID) UnmarshalGQLContext(ctx context.Context, v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("`ID` must be a string")
	}
	k.external = s
	return nil
}

func (k OpaqueID) MarshalGQLContext(ctx context.Context, w io.Writer) error {
	if k.codec == nil {
		return fmt.Errorf("no codec set for HiddenID")
	}

	encoded, err := k.codec.Encode(ctx, k.ID)
	if err != nil {
		return fmt.Errorf("failed to encode HiddenID: %w", err)
	}

	_, err = w.Write([]byte(`"` + encoded + `"`))
	return err
}

func (oid OpaqueID) Value() (driver.Value, error) {
	return oid.ID, nil
}

func (oid *OpaqueID) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		oid.ID = int(v)
	case int:
		oid.ID = v
	case int32:
		oid.ID = int(v)
	default:
		return fmt.Errorf("cannot scan type %T into OpaqueID", value)
	}
	return nil
}
