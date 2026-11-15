package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var SuperCodeApiKey = ""

const ModelID = "nvidia/nemotron-3.5-lightning:free"
const EndPoint = "https://openrouter.ai/api/v1/chat/completions"
const envFilePath = "internal/config/.env"

func LoadKeys() (error, string) {

	err := godotenv.Load(envFilePath)

	if err != nil {
		log.Fatal("Error Reading keys check the .env file")
		return err, SuperCodeApiKey
	}

	SuperCodeApiKey = os.Getenv("SUPERCODE_API_KEY")

	if SuperCodeApiKey == "" {
		log.Fatal("The Key is empty enter a valid key")
	}

	return err, SuperCodeApiKey
}
