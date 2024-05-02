package prefab

import (
	"errors"
	"fmt"
	"os"
)

type Datasource int

const (
	ApiKeyEnvVar = "PREFAB_API_KEY"
	DomainEnvVar = "PREFAB_DOMAIN"
)

const (
	ALL Datasource = iota
	LOCAL_ONLY
)

type Options struct {
	ApiKey                       string
	Datasource                   Datasource
	PrefabDomain                 string
	InitializationTimeoutSeconds int
	OnInitializationFailure      OnInitializationFailure
}

var DefaultOptions = Options{ApiKey: "", PrefabDomain: "prefab.cloud", Datasource: ALL, InitializationTimeoutSeconds: 10, OnInitializationFailure: RAISE}

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

func (o *Options) PrefabDomainEnvVarOrSetting() (string, error) {
	domainFromEnvVar := os.Getenv(DomainEnvVar)
	if domainFromEnvVar != "" {
		o.PrefabDomain = domainFromEnvVar
	}
	if o.PrefabDomain == "" {
		return "", errors.New("no PrefabDomain set")
	}
	return o.PrefabDomain, nil
}

type OnInitializationFailure int

const (
	RAISE  OnInitializationFailure = iota // RAISE = 0
	UNLOCK                                // UNLOCK = 1
)
