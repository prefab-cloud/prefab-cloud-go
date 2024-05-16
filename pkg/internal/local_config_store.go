package internal

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/options"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type LocalConfigStore struct {
	configMap       map[string]*prefabProto.Config
	sourceDirectory string
	Initialized     bool
}

func NewLocalConfigStore(sourceDirectory string, options *options.Options) *LocalConfigStore {
	configMap := make(map[string]*prefabProto.Config)

	envNames := append([]string{"default"}, options.EnvironmentNames...)
	for _, envName := range envNames {
		file := filepath.Join(sourceDirectory, ".prefab."+envName+".config.yaml")
		loadFileIntoMap(file, &configMap)
	}

	return &LocalConfigStore{sourceDirectory: sourceDirectory, configMap: configMap, Initialized: true}
}

func loadFileIntoMap(filePath string, configmap *map[string]*prefabProto.Config) {
	parser := &LocalConfigYamlParser{}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Debug(fmt.Sprintf("File %s does not exist\n", filePath))
		}

		return
	}

	configs, err := parser.Parse(data)
	if err != nil {
		return
	}

	for _, config := range configs {
		(*configmap)[config.GetKey()] = config
	}
}

func (s *LocalConfigStore) GetConfig(key string) (*prefabProto.Config, bool) {
	config, exists := s.configMap[key]
	return config, exists
}
