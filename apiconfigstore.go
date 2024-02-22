package prefab

import (
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"sync"
)

//TODO: support the api-default-context once we've worked out how we want to store/merge contexts
// may want to store it other than in original proto form?

type ApiConfigStore struct {
	configMap     map[string]*prefabProto.Config
	initialized   bool
	highWatermark int64
	sync.RWMutex
}

func BuildApiConfigStore() *ApiConfigStore {
	return &ApiConfigStore{
		configMap:     make(map[string]*prefabProto.Config),
		initialized:   false,
		highWatermark: 0,
	}
}

func (cs *ApiConfigStore) SetConfigs(configs []*prefabProto.Config) {
	cs.Lock()
	defer cs.Unlock()
	cs.initialized = true
	for _, config := range configs {
		cs.setConfig(config)
	}
}

func (cs *ApiConfigStore) SetFromConfigsProto(configs *prefabProto.Configs) {
	cs.SetConfigs(configs.Configs)
}

func (cs *ApiConfigStore) Len() int {
	return len(cs.configMap)
}

func (cs *ApiConfigStore) setConfig(newConfig *prefabProto.Config) {
	newConfigIsEmpty := len(newConfig.Rows) == 0
	currentConfig, exists := cs.configMap[newConfig.Key]
	switch {
	case newConfigIsEmpty && exists && newConfig.Id > currentConfig.Id:
		delete(cs.configMap, newConfig.Key)
	case !newConfigIsEmpty && (exists && newConfig.Id > currentConfig.Id) || (!exists):
		cs.configMap[newConfig.Key] = newConfig
	}
	if newConfig.Id > cs.highWatermark {
		cs.highWatermark = newConfig.Id
	}
}

// GetConfig retrieves a Config associated with the given key.
// It returns a pointer to the Config and a boolean value.
// The Config pointer is nil if the key does not exist in the store.
// The boolean value is true if the key exists, and false otherwise.
func (cs *ApiConfigStore) GetConfig(key string) (*prefabProto.Config, bool) {
	cs.RLock()
	defer cs.RUnlock()
	config, exists := cs.configMap[key]
	return config, exists
}
