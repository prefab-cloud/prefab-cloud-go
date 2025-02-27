package stores

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

func TestNewMemoryConfigStore_WithNativeTypes(t *testing.T) {
	// Arrange
	projectEnvID := int64(123)
	rawConfigs := map[string]interface{}{
		"string_key": "string_value",
		"int_key":    42,
		"bool_key":   true,
		"float_key":  3.14,
		"slice_key":  []string{"a", "b", "c"},
		"map_key":    map[string]interface{}{"nested": "value"},
	}

	// Act
	store, err := NewMemoryConfigStore(projectEnvID, rawConfigs)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, store)
	assert.Equal(t, projectEnvID, store.ProjectEnvID)
	assert.Len(t, store.configMap, len(rawConfigs))

	// Check each key was properly converted
	stringConfig, ok := store.GetConfig("string_key")
	require.True(t, ok)
	assert.Equal(t, "string_key", stringConfig.Key)
	assert.Equal(t, "string_value", stringConfig.Rows[0].Values[0].Value.GetString_())

	intConfig, ok := store.GetConfig("int_key")
	require.True(t, ok)
	assert.Equal(t, "int_key", intConfig.Key)
	assert.Equal(t, int64(42), intConfig.Rows[0].Values[0].Value.GetInt())

	boolConfig, ok := store.GetConfig("bool_key")
	require.True(t, ok)
	assert.Equal(t, "bool_key", boolConfig.Key)
	assert.Equal(t, true, boolConfig.Rows[0].Values[0].Value.GetBool())

	floatConfig, ok := store.GetConfig("float_key")
	require.True(t, ok)
	assert.Equal(t, "float_key", floatConfig.Key)
	assert.Equal(t, 3.14, floatConfig.Rows[0].Values[0].Value.GetDouble())

	// Keys in configMap should match the input keys
	keys := store.Keys()
	assert.Len(t, keys, len(rawConfigs))
	for key := range rawConfigs {
		assert.Contains(t, keys, key)
	}
}

func TestNewMemoryConfigStore_WithConfigObjects(t *testing.T) {
	// Arrange
	projectEnvID := int64(123)

	// Create some config objects directly
	stringValue := "direct_config_value"
	directConfig := &prefabProto.Config{
		Key:        "direct_config_key",
		Id:         42,
		ConfigType: prefabProto.ConfigType_CONFIG,
		Rows: []*prefabProto.ConfigRow{
			{
				Values: []*prefabProto.ConditionalValue{
					{
						Value: &prefabProto.ConfigValue{
							Type: &prefabProto.ConfigValue_String_{
								String_: stringValue,
							},
						},
					},
				},
			},
		},
	}

	featureFlagValue := true
	featureFlagConfig := &prefabProto.Config{
		Key:        "feature_flag_key",
		Id:         43,
		ConfigType: prefabProto.ConfigType_FEATURE_FLAG,
		Rows: []*prefabProto.ConfigRow{
			{
				Values: []*prefabProto.ConditionalValue{
					{
						Value: &prefabProto.ConfigValue{
							Type: &prefabProto.ConfigValue_Bool{
								Bool: featureFlagValue,
							},
						},
					},
				},
			},
		},
	}

	// Mix of direct configs and native types
	rawConfigs := map[string]interface{}{
		"direct_config_key": directConfig,
		"feature_flag_key":  featureFlagConfig,
		"native_string_key": "native_value",
	}

	// Act
	store, err := NewMemoryConfigStore(projectEnvID, rawConfigs)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, store)
	assert.Equal(t, projectEnvID, store.ProjectEnvID)
	assert.Len(t, store.configMap, len(rawConfigs))

	// Check direct config was stored unchanged
	directResult, ok := store.GetConfig("direct_config_key")
	require.True(t, ok)
	assert.Equal(t, directConfig, directResult) // Should be the exact same object
	assert.Equal(t, int64(42), directResult.Id) // ID should be preserved
	assert.Equal(t, prefabProto.ConfigType_CONFIG, directResult.ConfigType)
	assert.Equal(t, stringValue, directResult.Rows[0].Values[0].Value.GetString_())

	// Check feature flag config was stored unchanged
	featureFlagResult, ok := store.GetConfig("feature_flag_key")
	require.True(t, ok)
	assert.Equal(t, featureFlagConfig, featureFlagResult) // Should be the exact same object
	assert.Equal(t, int64(43), featureFlagResult.Id)      // ID should be preserved
	assert.Equal(t, prefabProto.ConfigType_FEATURE_FLAG, featureFlagResult.ConfigType)
	assert.Equal(t, featureFlagValue, featureFlagResult.Rows[0].Values[0].Value.GetBool())

	// Check native type was converted
	nativeResult, ok := store.GetConfig("native_string_key")
	require.True(t, ok)
	assert.Equal(t, "native_string_key", nativeResult.Key)
	assert.Equal(t, int64(0), nativeResult.Id) // Default ID for converted values
	assert.Equal(t, "native_value", nativeResult.Rows[0].Values[0].Value.GetString_())
}

func TestNewMemoryConfigStore_WithEmptyConfigs(t *testing.T) {
	// Arrange
	projectEnvID := int64(123)
	rawConfigs := map[string]interface{}{}

	// Act
	store, err := NewMemoryConfigStore(projectEnvID, rawConfigs)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, store)
	assert.Equal(t, projectEnvID, store.ProjectEnvID)
	assert.Empty(t, store.configMap)
	assert.Empty(t, store.Keys())
}

func TestNewMemoryConfigStore_WithInvalidValue(t *testing.T) {
	// Arrange
	projectEnvID := int64(123)

	// Create an invalid value that can't be converted
	type customType struct{}
	invalidValue := customType{}

	rawConfigs := map[string]interface{}{
		"valid_key":   "valid_value",
		"invalid_key": invalidValue,
	}

	// Act
	store, err := NewMemoryConfigStore(projectEnvID, rawConfigs)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create config value for key invalid_key")
	assert.Nil(t, store)
}
