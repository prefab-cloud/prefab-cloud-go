package options

import (
	"fmt"
	"os"
	"strings"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
)

type OnInitializationFailure int

const (
	ReturnError    OnInitializationFailure = iota // ReturnError = 0
	ReturnNilMatch                                // UNLOCK = 1
)

const (
	// #nosec G101 -- This is just the env var name
	APIKeyEnvVar = "PREFAB_API_KEY"
	APIURLVar    = "PREFAB_API_URL"
)

func GetDefaultAPIURLs() []string {
	return []string{
		"https://belt.prefab.cloud",
		"https://suspenders.prefab.cloud",
	}
}

type Options struct {
	GlobalContext                *contexts.ContextSet
	APIKey                       string
	APIURLs                      []string
	Sources                      []ConfigSource
	EnvironmentNames             []string
	ProjectEnvID                 int64
	InitializationTimeoutSeconds float64
	OnInitializationFailure      OnInitializationFailure
}

const timeoutDefault = 10.0

func GetDefaultOptions() Options {
	return Options{
		APIKey:                       "",
		APIURLs:                      nil,
		InitializationTimeoutSeconds: timeoutDefault,
		OnInitializationFailure:      ReturnError,
		GlobalContext:                contexts.NewContextSet(),
		Sources:                      GetDefaultConfigSources(),
	}
}

func (o *Options) APIKeySettingOrEnvVar() (string, error) {
	if o.APIKey == "" {
		// Attempt to read from an environment variable if APIKey is not directly set
		envAPIKey := os.Getenv(APIKeyEnvVar)
		if envAPIKey == "" {
			return "", fmt.Errorf("API key is not set and not found in environment variable %s", APIKeyEnvVar)
		}

		o.APIKey = envAPIKey
	}

	return o.APIKey, nil
}

func (o *Options) PrefabAPIURLEnvVarOrSetting() ([]string, error) {
	apiURLs := []string{}

	if os.Getenv(APIURLVar) != "" {
		for _, url := range strings.Split(os.Getenv(APIURLVar), ",") {
			if url != "" {
				apiURLs = append(apiURLs, url)
			}
		}

		if len(apiURLs) == 0 {
			return nil, fmt.Errorf("environment variable %s is blank", APIURLVar)
		}

		return apiURLs, nil
	}

	for _, url := range o.APIURLs {
		if url != "" {
			apiURLs = append(apiURLs, url)
		}
	}

	if o.APIURLs == nil {
		apiURLs = GetDefaultAPIURLs()
	}

	return apiURLs, nil
}
