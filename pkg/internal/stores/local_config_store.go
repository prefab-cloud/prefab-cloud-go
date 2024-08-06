package stores

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
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
	parser := &internal.LocalConfigYamlParser{}

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

func (s *LocalConfigStore) Keys() []string {
	keys := make([]string, 0, len(s.configMap))
	for key := range s.configMap {
		keys = append(keys, key)
	}

	return keys
}

func (s *LocalConfigStore) GetProjectEnvID() int64 {
	return 0
}

func (s *LocalConfigStore) GetContextValue(_ string) (value interface{}, valueExists bool) {
	return nil, false
}
