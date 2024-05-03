package prefab

import (
	"fmt"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"log/slog"
	"os"
	"path/filepath"
)

type LocalConfigStore struct {
	source_directory string
	configMap        map[string]*prefabProto.Config
	initialized      bool
}

func NewLocalConfigStore(source_directory string, options *Options) *LocalConfigStore {
	configMap := make(map[string]*prefabProto.Config)
	env_names := append([]string{"default"}, options.EnvironmentNames...)
	for _, env_name := range env_names {
		file := filepath.Join(source_directory, ".prefab."+env_name+".config.yaml")
		loadFileIntoMap(file, &configMap)
	}
	return &LocalConfigStore{source_directory: source_directory, configMap: configMap, initialized: true}

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
	configs, err := parser.parse(data)
	if err != nil {
		return
	}
	for _, config := range configs {
		(*configmap)[config.Key] = config
	}

}

func (s *LocalConfigStore) GetConfig(key string) (config *prefabProto.Config, exists bool) {
	config, exists = s.configMap[key]
	return config, exists
}
