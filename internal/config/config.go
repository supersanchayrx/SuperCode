package config

import "os"

// Config holds all configuration for SuperCode.
type Config struct {
	APIKey string
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		APIKey: os.Getenv("SUPERCODE_API_KEY"),
	}
}
