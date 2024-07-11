package internal

import (
	"fmt"
	"log/slog"
	"os"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type LocalConfigStore struct {
	configMap   map[string]*prefabProto.Config
	Initialized bool
}

func NewLocalConfigStore(path string) (*LocalConfigStore, error) {
	configMap := make(map[string]*prefabProto.Config)

	err := loadFileIntoMap(path, &configMap)

	if err != nil {
		return nil, err
	}

	return &LocalConfigStore{configMap: configMap, Initialized: true}, nil
}

func loadFileIntoMap(filePath string, configmap *map[string]*prefabProto.Config) error {
	parser := &LocalConfigYamlParser{}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Debug(fmt.Sprintf("File %s does not exist\n", filePath))
		}

		return err
	}

	configs, err := parser.Parse(data)
	if err != nil {
		return err
	}

	for _, config := range configs {
		(*configmap)[config.GetKey()] = config
	}

	return nil
}

func (s *LocalConfigStore) GetConfig(key string) (*prefabProto.Config, bool) {
	config, exists := s.configMap[key]

	return config, exists
}

func (cs *LocalConfigStore) Keys() []string {
	keys := make([]string, 0, len(cs.configMap))
	for key := range cs.configMap {
		keys = append(keys, key)
	}

	return keys
}

func (cs *LocalConfigStore) GetProjectEnvID() int64 {
	return 0
}

func (cs *LocalConfigStore) GetContextValue(propertyName string) (value interface{}, valueExists bool) {
	return nil, false
}
