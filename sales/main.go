package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	cfg, err := config(ctx)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	_, err = conn(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to establish connection: %v", err)
	}
}
