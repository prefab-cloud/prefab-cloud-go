package prefab

import (
	prefabInternalProto "github.com/prefab-cloud/prefab-cloud-go/internal-proto"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/options"
)

type Options = options.Options

type Option func(*options.Options) error

type ConfigDump = prefabInternalProto.ConfigDump

func WithOfflineSources(sources []string) Option {
	return WithSources(sources, true)
}

func WithSources(sources []string, excludeDefault bool) Option {
	// some of these are their own datastore, some map to the same datastore (e.g. poll and sse)

	// e.g.
	// "datafile://secret-submodule/secrets.json",
	// "datafile://this-project/project-config.json",
	// "sse:sse.prefab.cloud"

	return func(o *options.Options) error {
		configSources := make([]options.ConfigSource, 0, len(sources))

		for _, source := range sources {
			configSource, err := options.ParseConfigSource(source)
			if err != nil {
				return err
			}

			configSources = append(configSources, configSource)
		}

		if !excludeDefault {
			for _, defaultSource := range options.DefaultConfigSources {
				configSources = append(configSources, defaultSource)
			}
		}

		o.Sources = configSources

		return nil
	}
}

func WithGlobalContext(globalContext *ContextSet) Option {
	return func(o *options.Options) error {
		o.GlobalContext = globalContext

		return nil
	}
}

func WithEnvironmentNames(environmentNames []string) Option {
	return func(o *options.Options) error {
		o.EnvironmentNames = environmentNames

		return nil
	}
}

func WithAPIKey(apiKey string) Option {
	return func(o *options.Options) error {
		o.APIKey = apiKey

		return nil
	}
}

func WithAPIURL(apiURL string) Option {
	return func(o *options.Options) error {
		o.APIUrl = apiURL

		return nil
	}
}

func WithInitializationTimeoutSeconds(timeoutSeconds float64) Option {
	return func(o *options.Options) error {
		o.InitializationTimeoutSeconds = timeoutSeconds

		return nil
	}
}

func WithOnInitializationFailure(onInitializationFailure options.OnInitializationFailure) Option {
	return func(o *options.Options) error {
		o.OnInitializationFailure = onInitializationFailure

		return nil
	}
}
