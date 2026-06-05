package main

import (
	"fmt"
	"os"

	"supercode/internal/config"
)

func main() {
	// Load configuration (API key, etc.)
	cfg := config.Load()

	if cfg.APIKey == "" {
		fmt.Println("⚠  SUPERCODE_API_KEY is not set. Set it before using the agent.")
	}

	fmt.Println("⚡ SuperCode — AI Coding Agent")
	fmt.Println("Ready to go. Run with a task to get started.")
	os.Exit(0)
}
