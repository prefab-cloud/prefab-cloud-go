package options

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
)

type Datasource int

const (
	ALL Datasource = iota
	// TODO: support LocalOnly
	// LocalOnly
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
	ConfigDirectory              *string
	ConfigOverrideDirectory      *string
	GlobalContext                *contexts.ContextSet
	APIKey                       string
	APIUrl                       string
	EnvironmentNames             []string
	Datasource                   Datasource
	InitializationTimeoutSeconds float64
	OnInitializationFailure      OnInitializationFailure
}

const timeoutDefault = 10.0

var DefaultOptions = Options{
	APIKey:                       "",
	APIUrl:                       "https://api.prefab.cloud",
	Datasource:                   ALL,
	InitializationTimeoutSeconds: timeoutDefault,
	OnInitializationFailure:      RAISE,
	EnvironmentNames:             getDefaultEnvironmentNames(),
	ConfigDirectory:              getHomeDir(),
	ConfigOverrideDirectory:      StringPtr("."),
	GlobalContext:                contexts.NewContextSet(),
}

func getHomeDir() *string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		usr, err := user.Current()
		if err != nil {
			fmt.Println(err)

			return nil
		}

		homeDir = usr.HomeDir
	}

	return StringPtr(homeDir)
}

func getDefaultEnvironmentNames() []string {
	envVar := os.Getenv("PREFAB_ENVS")
	if envVar == "" {
		return []string{}
	}

	return strings.Split(envVar, ",")
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
