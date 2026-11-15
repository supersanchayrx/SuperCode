package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	EndPoint      = "https://openrouter.ai/api/v1/chat/completions"
	DefaultModel  = "nvidia/nemotron-3.5-lightning:free"
	FallbackModel = "openrouter/free"
	devEnvPath    = "internal/config/.env"
	envFileName   = ".env"
	configDirName = ".supercode"
	keyPrefix     = "SUPERCODE_API_KEY"
)

var ErrNoKeys = errors.New("no API keys found")

var (
	APIKeys           []string
	OrchestratorModel = DefaultModel
	PlanModel         = DefaultModel
	EditModel         = DefaultModel
)

func LoadKeys() ([]string, error) {
	for _, path := range EnvFiles() {
		err := godotenv.Load(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error reading %s: %w", path, err)
		}
	}

	APIKeys = collectKeys(os.Environ())

	if len(APIKeys) == 0 {
		return nil, ErrNoKeys
	}

	OrchestratorModel = envOr("ORCHESTRATOR_MODEL", DefaultModel)
	PlanModel = envOr("PLAN_MODEL", OrchestratorModel)
	EditModel = envOr("EDIT_MODEL", OrchestratorModel)

	return APIKeys, nil
}

func UserEnvFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDirName, envFileName), nil
}

func EnvFiles() []string {
	var files []string
	if exe, err := os.Executable(); err == nil {
		files = append(files, filepath.Join(filepath.Dir(exe), envFileName))
	}
	if path, err := UserEnvFile(); err == nil {
		files = append(files, path)
	}
	return append(files, devEnvPath)
}

const envTemplate = `# SuperCode settings
# Get a free key at https://openrouter.ai/keys and put it after the = sign.
# You can add more keys as SUPERCODE_API_KEY2, SUPERCODE_API_KEY3, ...
# SUPERCODE_API_KEY1=paste-your-key-here

# Optional: models per task (defaults to nvidia/nemotron-3.5-lightning:free)
# ORCHESTRATOR_MODEL=
# PLAN_MODEL=
# EDIT_MODEL=
`

func EnsureUserEnvFile() (string, error) {
	path, err := UserEnvFile()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, []byte(envTemplate), 0o600)
}

func SaveKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("the key is empty")
	}

	path, err := EnsureUserEnvFile()
	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n%s1=%s\n", keyPrefix, key); err != nil {
		return "", err
	}

	os.Setenv(keyPrefix+"1", key)
	return path, nil
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
