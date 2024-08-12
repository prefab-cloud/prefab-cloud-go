package stores

import (
	"errors"
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"

	prefabInternalProto "github.com/prefab-cloud/prefab-cloud-go/internal-proto"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ConfigDumpConfigStore struct {
	configMap    map[string]*prefabProto.Config
	path         string
	ProjectEnvID int64
	Initialized  bool
}

func NewConfigDumpConfigStore(path string, projectEnvID int64) (*ConfigDumpConfigStore, error) {
	configMap, err := configDumpFileToConfigMap(path)
	if err != nil {
		return nil, fmt.Errorf("error creating ConfigDumpConfigStore: %w", err)
	}

	if projectEnvID == 0 {
		return nil, errors.New("projectEnvID must be provided for ConfigDumpConfigStore")
	}

	return &ConfigDumpConfigStore{configMap: configMap, Initialized: true, path: path, ProjectEnvID: projectEnvID}, nil
}

func configDumpFileToConfigMap(path string) (map[string]*prefabProto.Config, error) {
	configMap := make(map[string]*prefabProto.Config)

	var configDump prefabInternalProto.ConfigDump

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config dump file %s: %w", path, err)
	}

	err = proto.Unmarshal(data, &configDump)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config dump file %s: %w", path, err)
	}

	for _, wrapper := range configDump.GetWrappers() {
		if !wrapper.GetDeleted() {
			config := wrapper.GetConfig()
			configMap[config.GetKey()] = config
		}
	}

	return configMap, nil
}

func (s *ConfigDumpConfigStore) GetConfig(key string) (*prefabProto.Config, bool) {
	config, exists := s.configMap[key]

	return config, exists
}

func (s *ConfigDumpConfigStore) GetContextValue(_ string) (value interface{}, valueExists bool) {
	return nil, false
}

func (s *ConfigDumpConfigStore) Keys() []string {
	keys := make([]string, 0, len(s.configMap))
	for key := range s.configMap {
		keys = append(keys, key)
	}

	return keys
}

func (s *ConfigDumpConfigStore) GetProjectEnvID() int64 {
	return s.ProjectEnvID
}
