package utils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/utils"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

func TestExtractDurationValue(t *testing.T) {
	t.Run("parsing durations", func(t *testing.T) {
		value := &prefabProto.ConfigValue_Duration{Duration: &prefabProto.IsoDuration{Definition: "PT1M"}}
		configValue := &prefabProto.ConfigValue{Type: value}

		duration, ok := utils.ExtractDurationValue(configValue)

		assert.True(t, ok)
		assert.Equal(t, time.Minute, duration)
	})
}

func TestExtractValue(t *testing.T) {
	t.Run("parsing durations", func(t *testing.T) {
		value := &prefabProto.ConfigValue_Duration{Duration: &prefabProto.IsoDuration{Definition: "PT1M"}}
		configValue := &prefabProto.ConfigValue{Type: value}

		extracted, ok, err := utils.ExtractValue(configValue)
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, time.Minute, extracted)
	})
}
