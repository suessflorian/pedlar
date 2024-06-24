package main

import (
	"context"
	"log"

	"github.com/suessflorian/pedlar/sales/internal/config"
	"github.com/suessflorian/pedlar/sales/internal/store"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Config(ctx)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	_, err = store.Conn(ctx, cfg.DatabaseURL, "sales")
	if err != nil {
		log.Fatalf("failed to establish connection: %v", err)
	}
}
