package config

import (
	"context"
	"fmt"
	"os"
	"reflect"

	env "github.com/joho/godotenv"
)

type Cfg struct {
	DatabaseURL string `env:"DATABASE_URL"`
}

func Config(ctx context.Context) (Cfg, error) {
	err := env.Load()
	if err != nil {
		return Cfg{}, fmt.Errorf("failed to load .env file: %w", err)
	}

	var cfg Cfg
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
