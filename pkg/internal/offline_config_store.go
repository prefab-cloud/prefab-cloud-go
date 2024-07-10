package internal

import (
	"github.com/prefab-cloud/prefab-cloud-go/pkg/options"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type OfflineConfigStore struct {
	configMap    map[string]*prefabProto.Config
	Initialized  bool
	ProjectEnvID int64
}

func NewOfflineConfigStore(offlineConfig options.OfflineConfig) *OfflineConfigStore {
	configMap := make(map[string]*prefabProto.Config)

	configs := offlineConfig.Configs

	for i := range configs {
		config := configs[i]
		configMap[config.GetKey()] = config
	}

	return &OfflineConfigStore{configMap: configMap, ProjectEnvID: offlineConfig.ProjectEnvID, Initialized: true}
}

func (s *OfflineConfigStore) GetConfig(key string) (*prefabProto.Config, bool) {
	config, exists := s.configMap[key]

	return config, exists
}

func (cs *OfflineConfigStore) GetContextValue(propertyName string) (interface{}, bool) {
	return nil, false
}

func (cs *OfflineConfigStore) GetProjectEnvID() int64 {
	return cs.ProjectEnvID
}

func (cs *OfflineConfigStore) Keys() []string {
	keys := make([]string, 0, len(cs.configMap))
	for key := range cs.configMap {
		keys = append(keys, key)
	}

	return keys
}
