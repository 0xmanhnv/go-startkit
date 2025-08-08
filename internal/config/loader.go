package config

import (
	"sync"

	"github.com/caarlos0/env/v10"
)

var (
	once   sync.Once
	global *Config
)

// Load parses environment variables into a global singleton config and returns it.
// Subsequent calls return the same instance without reparsing.
func Load() *Config {
	once.Do(func() {
		cfg := &Config{}
		if err := env.Parse(cfg); err != nil {
			panic("failed to parse env: " + err.Error())
		}
		global = cfg
	})
	return global
}

// Get returns the already loaded config, or loads it once if not yet loaded.
func Get() *Config { return Load() }
