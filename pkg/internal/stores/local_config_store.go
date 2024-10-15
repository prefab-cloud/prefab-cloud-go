package stores

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type LocalConfigStore struct {
	configMap    map[string]*prefabProto.Config
	Initialized  bool
	projectEnvID int64
}

func NewLocalConfigStore(path string) (*LocalConfigStore, error) {
	configMap := make(map[string]*prefabProto.Config)

	projectEnvID, err := loadFileIntoMap(path, &configMap)
	if err != nil {
		return nil, err
	}

	return &LocalConfigStore{configMap: configMap, projectEnvID: projectEnvID, Initialized: true}, nil
}

func parserFor(filePath string) (internal.ConfigParser, error) {
	switch {
	case strings.HasSuffix(filePath, ".json"):
		return &internal.LocalConfigJSONParser{}, nil
	case strings.HasSuffix(filePath, ".yaml"), strings.HasSuffix(filePath, ".yml"):
		return &internal.LocalConfigYamlParser{}, nil
	default:
		return nil, errors.New("unknown file extension. Supported extensions are .json, .yaml, .yml")
	}
}

func loadFileIntoMap(filePath string, configmap *map[string]*prefabProto.Config) (int64, error) {
	parser, err := parserFor(filePath)
	if err != nil {
		return 0, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Debug(fmt.Sprintf("File %s does not exist\n", filePath))
		}

		return 0, err
	}

	configs, projectEnvID, err := parser.Parse(data)
	if err != nil {
		fmt.Println("Error parsing config file", err)
		return 0, err
	}

	for _, config := range configs {
		fmt.Println("Adding config", config.GetKey())
		(*configmap)[config.GetKey()] = config
	}

	return projectEnvID, nil
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
	return s.projectEnvID
}

func (s *LocalConfigStore) GetContextValue(_ string) (value interface{}, valueExists bool) {
	return nil, false
}
