// TODO: need to figure out ProjectEnvID here. If we don't have an API key, we can't parse it so it'll need to be provided?
package internal

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"os"

	prefabInternalProto "github.com/prefab-cloud/prefab-cloud-go/internal-proto"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ConfigDumpConfigStore struct {
	configMap   map[string]*prefabProto.Config
	path        string
	Initialized bool
}

func NewConfigDumpConfigStore(path string) (*ConfigDumpConfigStore, error) {
	configMap, err := configDumpFileToConfigMap(path)

	if err != nil {
		return nil, fmt.Errorf("error creating ConfigDumpConfigStore: %v", err)
	}

	return &ConfigDumpConfigStore{configMap: configMap, Initialized: true, path: path}, nil
}

func configDumpFileToConfigMap(path string) (map[string]*prefabProto.Config, error) {
	configMap := make(map[string]*prefabProto.Config)

	var configDump prefabInternalProto.ConfigDump

	data, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("error reading config dump file %s: %v", path, err)
	}

	err = proto.Unmarshal(data, &configDump)

	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config dump file %s: %v", path, err)
	}

	for _, wrapper := range configDump.Wrappers {
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

func (cs *ConfigDumpConfigStore) GetContextValue(propertyName string) (value interface{}, valueExists bool) {
	return nil, false
}

func (cs *ConfigDumpConfigStore) Keys() []string {
	keys := make([]string, 0, len(cs.configMap))
	for key := range cs.configMap {
		keys = append(keys, key)
	}

	return keys
}

func (cs *ConfigDumpConfigStore) GetProjectEnvID() int64 {
	return 0
}
