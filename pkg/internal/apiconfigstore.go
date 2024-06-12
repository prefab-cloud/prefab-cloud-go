package internal

import (
	"sync"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

// TODO: support the api-default-context once we've worked out how we want to store/merge contexts
// may want to store it other than in original proto form?

type APIConfigStore struct {
	configMap     map[string]*prefabProto.Config
	contextSet    *contexts.ContextSet
	highWatermark int64
	projectEnvID  int64
	sync.RWMutex
	Initialized bool
}

func BuildAPIConfigStore() *APIConfigStore {
	return &APIConfigStore{
		configMap:     make(map[string]*prefabProto.Config),
		Initialized:   false,
		highWatermark: 0,
		projectEnvID:  0,
	}
}

func (cs *APIConfigStore) SetConfigs(configs []*prefabProto.Config, envID int64) {
	cs.Lock()
	defer cs.Unlock()
	cs.Initialized = true
	cs.projectEnvID = envID

	for _, config := range configs {
		cs.setConfig(config)
	}
}

func (cs *APIConfigStore) SetFromConfigsProto(configs *prefabProto.Configs) {
	cs.contextSet = contexts.NewContextSetFromProto(configs.GetDefaultContext())
	cs.SetConfigs(configs.GetConfigs(), configs.GetConfigServicePointer().GetProjectEnvId())
}

func (cs *APIConfigStore) GetContextValue(propertyName string) (interface{}, bool) {
	value, valueExists := cs.contextSet.GetContextValue(propertyName)

	return value, valueExists
}

func (cs *APIConfigStore) Len() int {
	return len(cs.configMap)
}

func (cs *APIConfigStore) setConfig(newConfig *prefabProto.Config) {
	newConfigIsEmpty := len(newConfig.GetRows()) == 0
	currentConfig, exists := cs.configMap[newConfig.GetKey()]

	switch {
	case newConfigIsEmpty && exists && newConfig.GetId() > currentConfig.GetId():
		delete(cs.configMap, newConfig.GetKey())
	case !newConfigIsEmpty && (exists && newConfig.GetId() > currentConfig.GetId()) || (!exists):
		cs.configMap[newConfig.GetKey()] = newConfig
	}

	if newConfig.GetId() > cs.highWatermark {
		cs.highWatermark = newConfig.GetId()
	}
}

// GetConfig retrieves a Config associated with the given key.
// It returns a pointer to the Config and a boolean value.
// The Config pointer is nil if the key does not exist in the store.
// The boolean value is true if the key exists, and false otherwise.
func (cs *APIConfigStore) GetConfig(key string) (*prefabProto.Config, bool) {
	cs.RLock()
	defer cs.RUnlock()
	config, exists := cs.configMap[key]

	return config, exists
}

func (cs *APIConfigStore) GetProjectEnvID() int64 {
	cs.RLock()
	defer cs.RUnlock()

	return cs.projectEnvID
}
