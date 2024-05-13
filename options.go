package prefab

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"
)

type Datasource int

const (
	ApiKeyEnvVar = "PREFAB_API_KEY"
	ApiUrlVar    = "PREFAB_API_URL"
)

const (
	ALL Datasource = iota
	LOCAL_ONLY
)

type Options struct {
	ApiKey                       string
	Datasource                   Datasource
	ApiUrl                       string
	InitializationTimeoutSeconds float64
	OnInitializationFailure      OnInitializationFailure
	EnvironmentNames             []string
	ConfigDirectory              *string
	ConfigOverrideDirectory      *string
}

var DefaultOptions = Options{
	ApiKey:                       "",
	ApiUrl:                       "https://api.prefab.cloud",
	Datasource:                   ALL,
	InitializationTimeoutSeconds: 10,
	OnInitializationFailure:      RAISE,
	EnvironmentNames:             getDefaultEnvironmentNames(),
	ConfigDirectory:              getHomeDir(),
	ConfigOverrideDirectory:      StringPtr("."),
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

func (o *Options) apiKeySettingOrEnvVar() (string, error) {
	if o.ApiKey == "" {
		// Attempt to read from an environment variable if APIKey is not directly set
		envAPIKey := os.Getenv(ApiKeyEnvVar)
		if envAPIKey == "" {
			return "", errors.New(fmt.Sprintf("API key is not set and not found in environment variable %s", ApiKeyEnvVar))
		}

		o.ApiKey = envAPIKey
	}

	return o.ApiKey, nil
}

func (o *Options) PrefabApiUrlEnvVarOrSetting() (string, error) {
	apiUrlFromEnvVar := os.Getenv(ApiUrlVar)
	if apiUrlFromEnvVar != "" {
		o.ApiUrl = apiUrlFromEnvVar
	}

	if o.ApiUrl == "" {
		return "", errors.New("no PrefabApiUrl set")
	}

	return o.ApiUrl, nil
}

func StringPtr(string string) *string {
	return &string
}

type OnInitializationFailure int

const (
	RAISE  OnInitializationFailure = iota // RAISE = 0
	UNLOCK                                // UNLOCK = 1
)
