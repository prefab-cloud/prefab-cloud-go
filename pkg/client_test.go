package prefab_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	prefab "github.com/prefab-cloud/prefab-cloud-go/pkg"
)

func TestWithConfig(t *testing.T) {
	configs := map[string]interface{}{
		"string.key": "value",
		"int.key":    int64(42),
		"bool.key":   true,
		"float.key":  3.14,
		"slice.key":  []string{"a", "b", "c"},
		"json.key": map[string]interface{}{
			"nested": "value",
		},
	}

	client, err := prefab.NewClient(
		prefab.WithConfigs(configs),
		prefab.WithInitializationTimeoutSeconds(0.5))

	require.NoError(t, err)

	str, ok, err := client.GetStringValue("string.key", prefab.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "value", str)

	i, ok, err := client.GetIntValue("int.key", prefab.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, int64(42), i)

	b, ok, err := client.GetBoolValue("bool.key", prefab.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, b)

	f, ok, err := client.GetFloatValue("float.key", prefab.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 3.14, f, 0.0001)

	slice, ok, err := client.GetStringSliceValue("slice.key", prefab.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b", "c"}, slice)

	json, ok, err := client.GetJSONValue("json.key", prefab.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{"nested": "value"}, json)
}

func TestCannotUseWithConfigAndOtherSources(t *testing.T) {
	configs := map[string]interface{}{
		"string.key": "value",
	}

	_, err := prefab.NewClient(
		prefab.WithConfigs(configs),
		prefab.WithSources([]string{}, false)) // Explicitly try to use online sources

	require.Error(t, err)
	assert.Equal(t, "cannot use WithConfigs with other sources", err.Error())

	_, err = prefab.NewClient(
		prefab.WithConfigs(configs),
		prefab.WithOfflineSources([]string{"dump://testdata/example.dump"}),
		prefab.WithProjectEnvID(8),
	)

	require.Error(t, err)
	assert.Equal(t, "cannot use WithConfigs with other sources", err.Error())
}
