package options_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/options"
)

func TestGetDefaultConfigSources(t *testing.T) {
	sources := options.GetDefaultConfigSources()

	require.Len(t, sources, 1)

	assert.Equal(t, options.ConfigSource{
		Store:   options.APIStore,
		Raw:     "api:prefab",
		Default: true,
	}, sources[0])
}
