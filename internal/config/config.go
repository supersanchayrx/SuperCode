package config

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	EndPoint      = "https://openrouter.ai/api/v1/chat/completions"
	DefaultModel  = "nvidia/nemotron-3.5-lightning:free"
	FallbackModel = "openrouter/free"
	envFilePath   = "internal/config/.env"
	keyPrefix     = "SUPERCODE_API_KEY"
)

var (
	APIKeys           []string
	OrchestratorModel = DefaultModel
	PlanModel         = DefaultModel
	EditModel         = DefaultModel
)

func LoadKeys() ([]string, error) {
	if err := godotenv.Load(envFilePath); err != nil {
		return nil, fmt.Errorf("error reading keys, check the .env file: %w", err)
	}

	APIKeys = collectKeys(os.Environ())

	if len(APIKeys) == 0 {
		return nil, errors.New("no API keys found, set SUPERCODE_API_KEY1, SUPERCODE_API_KEY2, ... in the .env file")
	}

	OrchestratorModel = envOr("ORCHESTRATOR_MODEL", DefaultModel)
	PlanModel = envOr("PLAN_MODEL", OrchestratorModel)
	EditModel = envOr("EDIT_MODEL", OrchestratorModel)

	return APIKeys, nil
}

func collectKeys(environ []string) []string {
	type numberedKey struct {
		number int
		value  string
	}

	var found []numberedKey
	seen := make(map[string]bool)

	for _, entry := range environ {
		name, value, ok := strings.Cut(entry, "=")
		if !ok || !strings.HasPrefix(strings.ToUpper(name), keyPrefix) {
			continue
		}

		number := 0
		if suffix := name[len(keyPrefix):]; suffix != "" {
			n, err := strconv.Atoi(suffix)
			if err != nil {
				continue
			}
			number = n
		}

		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}

		seen[value] = true
		found = append(found, numberedKey{number, value})
	}

	sort.Slice(found, func(i, j int) bool { return found[i].number < found[j].number })

	keys := make([]string, 0, len(found))
	for _, k := range found {
		keys = append(keys, k.value)
	}

	return keys
}

func envOr(name, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return fallback
}
