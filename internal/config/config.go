package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

var SuperCodeApiKey = ""

const ModelID = "nvidia/nemotron-3.5-lightning:free"
const EndPoint = "https://openrouter.ai/api/v1/chat/completions"
const envFilePath = "internal/config/.env"

func LoadKeys() (string, error) {

	err := godotenv.Load(envFilePath)

	if err != nil {
		return "", fmt.Errorf("error reading keys, check the .env file: %w", err)
	}

	SuperCodeApiKey = os.Getenv("SUPERCODE_API_KEY")

	if SuperCodeApiKey == "" {
		return "", errors.New("the key is empty, set SUPERCODE_API_KEY in the .env file")
	}

	return SuperCodeApiKey, nil
}
