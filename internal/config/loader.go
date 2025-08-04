package config

import (
	"log"

	"github.com/caarlos0/env/v10"
)

func Load() *Config {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        log.Fatalf("failed to parse env: %v", err)
    }
    return cfg
}
