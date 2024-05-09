package prefab

import (
	"sync"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

// TODO: support the api-default-context once we've worked out how we want to store/merge contexts
// may want to store it other than in original proto form?

type ApiConfigStore struct {
	configMap     map[string]*prefabProto.Config
	initialized   bool
	highWatermark int64
	projectEnvId  int64
	contextSet    *ContextSet
	sync.RWMutex
}

func BuildApiConfigStore() *ApiConfigStore {
	return &ApiConfigStore{
		configMap:     make(map[string]*prefabProto.Config),
		initialized:   false,
		highWatermark: 0,
		projectEnvId:  0,
	}
}

func (cs *ApiConfigStore) SetConfigs(configs []*prefabProto.Config, envId int64) {
	cs.Lock()
	defer cs.Unlock()
	cs.initialized = true
	cs.projectEnvId = envId

	for _, config := range configs {
		cs.setConfig(config)
	}
}

func (cs *ApiConfigStore) SetFromConfigsProto(configs *prefabProto.Configs) {
	cs.contextSet = NewContextSetFromProto(configs.DefaultContext)
	cs.SetConfigs(configs.Configs, configs.ConfigServicePointer.GetProjectEnvId())
}

func (cs *ApiConfigStore) GetContextValue(propertyName string) (value interface{}, valueExists bool) {
	value, valueExists = cs.contextSet.GetContextValue(propertyName)
	return value, valueExists
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

func (cs *ApiConfigStore) GetProjectEnvId() int64 {
	cs.RLock()
	defer cs.RUnlock()

	return cs.projectEnvId
}
