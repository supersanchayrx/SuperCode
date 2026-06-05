package config

import "os"

// Config holds all configuration for SuperCode.
type Config struct {
	APIKey  string
	Model   string
	BaseURL string
	Stream  bool
}

// Load reads configuration from environment variables.
func Load() Config {
	model := os.Getenv("SUPERCODE_MODEL")
	if model == "" {
		model = "nvidia/nemotron-3-nano-30b-a3b:free"
	}

	baseurl := os.Getenv("SUPERCODE_BASE_URL")
	if baseurl == "" {
		baseurl = "https://openrouter.ai/api/v1"
	}

	var streamMode bool = true

	return Config{
		APIKey:  os.Getenv("SUPERCODE_API_KEY"),
		Model:   model,
		BaseURL: baseurl,
		Stream:  streamMode,
	}
}
