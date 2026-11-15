package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"supercode/internal/config"

	"github.com/spf13/cobra"
)

func loadConfig(cmd *cobra.Command, args []string) error {
	_, err := config.LoadKeys()
	if err == nil {
		return nil
	}
	if !errors.Is(err, config.ErrNoKeys) {
		return err
	}
	return firstRunSetup()
}

func firstRunSetup() error {
	path, err := config.EnsureUserEnvFile()
	if err != nil {
		return fmt.Errorf("no API key found, and could not create the settings file: %w", err)
	}

	fmt.Println("No API key found.")
	fmt.Println("SuperCode reads its keys from:", path)
	fmt.Println("Get a free OpenRouter key at https://openrouter.ai/keys")
	fmt.Print("Paste your key here to save it into that file (or press Enter to edit the file yourself): ")

	key := strings.TrimSpace(readLine())
	if key == "" {
		return fmt.Errorf("open %s, set SUPERCODE_API_KEY1=<your key>, then run supercode again", path)
	}

	if _, err := config.SaveKey(key); err != nil {
		return fmt.Errorf("could not save the key: %w", err)
	}
	fmt.Printf("Saved to %s\n\n", path)

	_, err = config.LoadKeys()
	return err
}

func readLine() string {
	var line []byte
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if n == 1 {
			if buf[0] == '\n' {
				break
			}
			line = append(line, buf[0])
		}
		if err != nil {
			break
		}
	}
	return string(line)
}
