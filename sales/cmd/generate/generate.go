package main

import (
	"context"
	"log"

	"github.com/suessflorian/pedlar/sales/internal/config"
	"github.com/suessflorian/pedlar/sales/internal/db"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Config(ctx)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	_, err = db.Conn(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to establish connection: %v", err)
	}
}
