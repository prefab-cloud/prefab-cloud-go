package options

import (
	"errors"
	"fmt"
	"os"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
)

type OnInitializationFailure int

const (
	RAISE  OnInitializationFailure = iota // RAISE = 0
	UNLOCK                                // UNLOCK = 1
)

const (
	// #nosec G101 -- This is just the env var name
	APIKeyEnvVar = "PREFAB_API_KEY"
	APIURLVar    = "PREFAB_API_URL"
)

type Options struct {
	Sources                      []ConfigSource
	GlobalContext                *contexts.ContextSet
	APIKey                       string
	APIUrl                       string
	EnvironmentNames             []string
	InitializationTimeoutSeconds float64
	OnInitializationFailure      OnInitializationFailure
}

const timeoutDefault = 10.0

var DefaultOptions = Options{
	APIKey:                       "",
	APIUrl:                       "https://api.prefab.cloud",
	InitializationTimeoutSeconds: timeoutDefault,
	OnInitializationFailure:      RAISE,
	GlobalContext:                contexts.NewContextSet(),
	Sources:                      DefaultConfigSources,
}

func NewOptions(modifyFn func(*Options)) Options {
	opts := DefaultOptions
	modifyFn(&opts)

	return opts
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

func (o *Options) PrefabAPIURLEnvVarOrSetting() (string, error) {
	apiURLFromEnvVar := os.Getenv(APIURLVar)
	if apiURLFromEnvVar != "" {
		o.APIUrl = apiURLFromEnvVar
	}

	if o.APIUrl == "" {
		return "", errors.New("no PrefabApiUrl set")
	}

	return o.APIUrl, nil
}

func StringPtr(str string) *string {
	return &str
}
