package options_test

import (
	"testing"

	prefab "github.com/prefab-cloud/prefab-cloud-go/pkg"

	"github.com/stretchr/testify/assert"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/options"
)

func TestGetDefaultOptions(t *testing.T) {
	// When ENV var PREFAB_API_URL_OVERRIDE is not set we use the default API URL.
	t.Setenv("PREFAB_API_URL_OVERRIDE", "")

	o := options.GetDefaultOptions()

	assert.Empty(t, o.APIKey)
	assert.Nil(t, o.APIURLs)
	assert.Equal(t, 10.0, o.InitializationTimeoutSeconds)
	assert.Equal(t, options.ReturnError, o.OnInitializationFailure)
	assert.NotNil(t, o.GlobalContext)
	assert.Len(t, o.Sources, 1)
	assert.Equal(t, options.ConfigSource{
		Store:   options.APIStore,
		Raw:     "api:prefab",
		Default: true,
	}, o.Sources[0])

	// When ENV var PREFAB_API_URL_OVERRIDE is set, that should be used instead
	// of the default API URL.
	desiredAPIURL := "https://api.staging-prefab.cloud"

	t.Setenv("PREFAB_API_URL_OVERRIDE", desiredAPIURL)

	o = options.GetDefaultOptions()
	assert.Equal(t, []string{desiredAPIURL}, o.APIURLs)

	t.Setenv("PREFAB_API_URL_OVERRIDE", "")
	t.Setenv("PREFAB_DATAFILE", "testdata/download.json")

	o = options.GetDefaultOptions()
	assert.Equal(t, []options.ConfigSource([]options.ConfigSource{{Store: "DataFile", Raw: "testdata/download.json", Path: "testdata/download.json", Default: false}}), o.Sources)
}

func TestOptions_TelemetryEnabledOverride(t *testing.T) {
	defaultOptions := options.GetDefaultOptions()
	assert.True(t, defaultOptions.TelemetryEnabled())
	_ = prefab.WithAllTelemetryDisabled()(&defaultOptions)
	assert.False(t, defaultOptions.TelemetryEnabled())
}

func TestOptions_TelemetryEnabledCalculatesBasedOnOptions(t *testing.T) {
	defaultOptions := options.GetDefaultOptions()
	assert.True(t, defaultOptions.TelemetryEnabled())

	_ = prefab.WithCollectEvaluationSummaries(false)(&defaultOptions)
	assert.True(t, defaultOptions.TelemetryEnabled())

	_ = prefab.WithContextTelemetryMode(options.ContextTelemetryModes.None)(&defaultOptions)
	assert.False(t, defaultOptions.TelemetryEnabled())
}
