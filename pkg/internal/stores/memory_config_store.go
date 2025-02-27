package stores

import (
	"fmt"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/utils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

// MemoryConfigStore is a simple in-memory key-value store.
type MemoryConfigStore struct {
	configMap    map[string]*prefabProto.Config
	ProjectEnvID int64
}

func NewMemoryConfigStore(projectEnvID int64, rawConfigs map[string]interface{}) (*MemoryConfigStore, error) {
	configs := make(map[string]*prefabProto.Config, len(rawConfigs))

	for key, rawValue := range rawConfigs {
		// Check if rawValue is already a prefabProto.Config
		if existingConfig, ok := rawValue.(*prefabProto.Config); ok {
			configs[key] = existingConfig
			continue
		}

		configValue, ok := utils.Create(rawValue)
		if !ok {
			return nil, fmt.Errorf("failed to create config value for key %s", key)
		}

		config := prefabProto.Config{
			Key: key,
			Id:  0,
			Rows: []*prefabProto.ConfigRow{
				{
					Values: []*prefabProto.ConditionalValue{
						{
							Value: configValue,
						},
					},
				},
			},
		}

		configs[key] = &config
	}

	return &MemoryConfigStore{
		ProjectEnvID: projectEnvID,
		configMap:    configs,
	}, nil
}

func (s *MemoryConfigStore) GetConfig(key string) (*prefabProto.Config, bool) {
	config, ok := s.configMap[key]

	return config, ok
}

func (s *MemoryConfigStore) GetContextValue(_ string) (interface{}, bool) {
	return nil, false
}

func (s *MemoryConfigStore) GetProjectEnvID() int64 {
	return s.ProjectEnvID
}

func (s *MemoryConfigStore) Keys() []string {
	keys := make([]string, 0, len(s.configMap))

	for key := range s.configMap {
		keys = append(keys, key)
	}

	return keys
}
