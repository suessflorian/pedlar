package main

import (
	"context"
	"fmt"
	"os"
	"reflect"

	env "github.com/joho/godotenv"
)

type cfg struct {
	DatabaseURL string `env:"DATABASE_URL"`
}

func config(ctx context.Context) (cfg, error) {
	err := env.Load()
	if err != nil {
		return cfg{}, fmt.Errorf("failed to load .env file: %w", err)
	}

	var cfg cfg
	v := reflect.ValueOf(&cfg).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("env")
		if tag == "" {
			continue
		}
		val := os.Getenv(tag)
		if val == "" {
			return cfg, fmt.Errorf("missing configurable %q", tag)
		}
		v.Field(i).SetString(val)
	}

	return cfg, nil
}
