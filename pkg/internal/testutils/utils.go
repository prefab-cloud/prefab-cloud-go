package testutils

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/utils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

func CreateConfigValueAndAssertOk(t *testing.T, value any) *prefabProto.ConfigValue {
	t.Helper()

	val, ok := utils.Create(value)
	if !ok {
		t.Fatalf("Unable to create a config value given %v", value)
	}

	return val
}

// Recursively sort JSON objects and arrays by keys and values
func sortJSONData(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// Sort map keys
		sortedKeys := make([]string, 0, len(v))
		for key := range v {
			sortedKeys = append(sortedKeys, key)
		}

		sort.Strings(sortedKeys)

		// Create a new sorted map
		sortedMap := make(map[string]interface{})
		for _, key := range sortedKeys {
			sortedMap[key] = sortJSONData(v[key])
		}

		return sortedMap
	case []interface{}:
		// Sort array elements based on their string representation
		sort.Slice(v, func(i, j int) bool {
			return strings.Compare(fmt.Sprint(v[i]), fmt.Sprint(v[j])) < 0
		})

		// Recursively sort the array elements (if they're objects or arrays)
		for i := range v {
			v[i] = sortJSONData(v[i])
		}

		return v
	default:
		// For primitive types, return as is
		return v
	}
}

// Marshal and sort both keys and arrays in the JSON output
func sortedJSON(message proto.Message) (string, error) {
	// Use protojson to marshal the message
	options := protojson.MarshalOptions{
		EmitUnpopulated: true,
	}

	jsonBytes, err := options.Marshal(message)
	if err != nil {
		return "", err
	}

	// Unmarshal into a map so we can sort it
	var jsonData interface{}
	if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
		return "", err
	}

	// Sort the JSON data (both objects and arrays)
	sortedData := sortJSONData(jsonData)

	// Marshal sorted data back to JSON
	sortedJSON, err := json.Marshal(sortedData)
	if err != nil {
		return "", err
	}

	return string(sortedJSON), nil
}

func AssertJSONEqual(t *testing.T, expected proto.Message, actualData proto.Message) {
	expectedJSON, err := sortedJSON(expected)
	require.NoError(t, err)

	actualJSON, err := sortedJSON(actualData)
	require.NoError(t, err)

	assert.Equal(t, expectedJSON, actualJSON)
}
