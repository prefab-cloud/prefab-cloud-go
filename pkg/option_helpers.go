package prefab

import (
	"time"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/options"
)

// Option is a function that modifies the options for the prefab client.
type Option func(*options.Options) error

// WithProjectEnvID sets the project environment ID for the prefab client.
// You only need to set this if you are using a config dump source.
func WithProjectEnvID(projectEnvID int64) Option {
	return func(o *options.Options) error {
		o.ProjectEnvID = projectEnvID

		return nil
	}
}

// WithOfflineSources allows providing customthe sources for the prefab client.
// Using this option will exclude the default (API + SSE) sources.
//
// Do not use this option if you are using the standard API or SSE sources.
// Example:
//
//	client, err := prefab.NewClient(
//		prefab.WithProjectEnvID(projectEnvID),
//		prefab.WithOfflineSources([]string{
//			"dump://" + fileName,
//		}))
func WithOfflineSources(sources []string) Option {
	return func(o *options.Options) error {
		o.ContextTelemetryMode = options.ContextTelemetryModes.None

		err := WithSources(sources, true)(o)
		if err != nil {
			return err
		}

		return nil
	}
}

// WithSources allows providing custom sources for the prefab client to use.
// This prepends your sources to the default sources (API + SSE).
func WithSources(sources []string, excludeDefault bool) Option {
	// some of these are their own datastore, some map to the same datastore (e.g. poll and sse)
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
			configSources = append(configSources, options.GetDefaultConfigSources()...)
		}

		o.Sources = configSources

		return nil
	}
}

// WithGlobalContext sets the global context for the prefab client.
func WithGlobalContext(globalContext *ContextSet) Option {
	return func(o *options.Options) error {
		o.GlobalContext = globalContext

		return nil
	}
}

// WithAPIKey sets the API key for the prefab client.
func WithAPIKey(apiKey string) Option {
	return func(o *options.Options) error {
		o.APIKey = apiKey

		return nil
	}
}

// WithAPIURLs sets the API URLs for the prefab client.
//
// You likely will never need to use this option.
func WithAPIURLs(apiURL []string) Option {
	return func(o *options.Options) error {
		o.APIURLs = apiURL

		return nil
	}
}

// WithInitializationTimeoutSeconds sets the initialization timeout for the prefab client. After this time, the client will either raise or continue depending on the OnInitializationFailure option.
func WithInitializationTimeoutSeconds(timeoutSeconds float64) Option {
	return func(o *options.Options) error {
		o.InitializationTimeoutSeconds = timeoutSeconds

		return nil
	}
}

// WithOnInitializationFailure sets the behavior for the prefab client when initialization fails.
//
// The default behavior is to return an error when a GetConfig (or similar) call is made (prefab.ReturnError).
//
// If you want to ignore the error and continue, use prefab.ReturnNilMatch.
func WithOnInitializationFailure(onInitializationFailure options.OnInitializationFailure) Option {
	return func(o *options.Options) error {
		o.OnInitializationFailure = onInitializationFailure

		return nil
	}
}

// WithConfigs lets you provide a map of configs to the prefab client to aid in
// testing. This is not compatible with other sources.
//
//	configs := map[string]interface{}{
//		"string.key": "value",
//		"int.key":    int64(42),
//		"bool.key":   true,
//		"float.key":  3.14,
//		"slice.key":  []string{"a", "b", "c"},
//		"json.key": map[string]interface{}{
//			"nested": "value",
//		},
//	}
//
//	client, err := prefab.NewClient(prefab.WithConfigs(configs))
func WithConfigs(configs map[string]interface{}) Option {
	return func(o *options.Options) error {
		err := WithOfflineSources([]string{options.MemoryStoreKey})(o)
		if err != nil {
			return err
		}

		o.Configs = configs

		return nil
	}
}

// WithTelemetrySyncInterval sets the interval at which the client will send telemetry data to the server.
func WithTelemetrySyncInterval(interval time.Duration) Option {
	return func(o *options.Options) error {
		o.TelemetrySyncInterval = interval

		return nil
	}
}

// WithTelemetryHost sets the host to which the client will send telemetry data. You likely will never need to use this option.
func WithTelemetryHost(host string) Option {
	return func(o *options.Options) error {
		o.TelemetryHost = host

		return nil
	}
}

// WithContextTelemetryMode sets the context telemetry mode for the prefab client.
func WithContextTelemetryMode(mode options.ContextTelemetryMode) Option {
	return func(o *options.Options) error {
		o.ContextTelemetryMode = mode

		return nil
	}
}

// WithCollectEvaluationSummaries sets whether the client should collect evaluation summaries.
//
// The default is true
func WithCollectEvaluationSummaries(collect bool) Option {
	return func(o *options.Options) error {
		o.CollectEvaluationSummaries = collect

		return nil
	}
}

func WithAllTelemetryDisabled() Option {
	return func(o *options.Options) error {
		o.ContextTelemetryMode = options.ContextTelemetryModes.None
		o.CollectEvaluationSummaries = false
		return nil
	}
}
