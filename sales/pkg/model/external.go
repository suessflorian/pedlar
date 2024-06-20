package model

import (
	"context"
	"fmt"
	"log/slog"
)

type ExternelIDFunc = func(ctx context.Context, object string, internalID int) (string, error)

func Must(f ExternelIDFunc) func(ctx context.Context, object string, internalID int) string {
	return func(ctx context.Context, object string, internalID int) string {
		external, err := f(ctx, object, internalID)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("failed to create external facing id: %v", err))
			panic("failed to created external facing id")
		}
		return external
	}
}
