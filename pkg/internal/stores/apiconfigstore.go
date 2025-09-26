package stores

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/options"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/sse"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

const maxRetries = 10

type APIConfigStore struct {
	configMap       map[string]*prefabProto.Config
	contextSet      *contexts.ContextSet
	httpClient      *internal.HTTPClient
	finishedLoading func()
	highWatermark   int64
	projectEnvID    int64
	sync.RWMutex
	Initialized bool
}

func NewAPIConfigStore(options options.Options, finishedLoading func()) (*APIConfigStore, error) {
	httpClient, err := internal.BuildHTTPClient(options)
	if err != nil {
		panic(err)
	}

	sseClient, err := sse.BuildSSEClient(options)
	if err != nil {
		panic(err)
	}

	store := &APIConfigStore{
		configMap:       make(map[string]*prefabProto.Config),
		Initialized:     false,
		highWatermark:   0,
		projectEnvID:    0,
		httpClient:      httpClient,
		finishedLoading: finishedLoading,
	}

	go func() {
		err = store.fetchFromServer(0, func() {
			go sse.StartSSEConnection(sseClient, store)
		})
		if err != nil {
			slog.Error(fmt.Sprintf("error fetching from server: %v", err))
		}
	}()

	return store, nil
}

func (cs *APIConfigStore) SetConfigs(configs []*prefabProto.Config, envID int64) {
	cs.Lock()
	defer cs.Unlock()
	cs.Initialized = true

	// Only update projectEnvID if we have configs to process
	// Empty config lists from SSE shouldn't change the environment context
	if len(configs) > 0 {
		cs.projectEnvID = envID
	}

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

func (cs *APIConfigStore) Keys() []string {
	cs.RLock()
	defer cs.RUnlock()

	keys := make([]string, 0, len(cs.configMap))
	for key := range cs.configMap {
		keys = append(keys, key)
	}

	return keys
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

func (cs *APIConfigStore) fetchFromServer(retriesAttempted int, then func()) error {
	configs, err := cs.httpClient.Load(cs.highWatermark)
	if err != nil {
		slog.Warn(fmt.Sprintf("unable to get data via http %v", err))

		if retriesAttempted < maxRetries {
			retryDelay := time.Duration(retriesAttempted) * time.Second

			slog.Debug(fmt.Sprintf("retrying in %d seconds, attempt %d/%d", int(retryDelay.Seconds()), retriesAttempted+1, maxRetries))

			time.Sleep(retryDelay)

			err = cs.fetchFromServer(retriesAttempted+1, then)
			if err != nil {
				return err
			}
		} else {
			slog.Error("max retries reached, giving up")
			then()

			return err
		}

		return nil
	}

	slog.Debug("Loaded configuration data")
	cs.SetFromConfigsProto(configs)

	cs.finishedLoading()

	then()

	return nil
}

func (cs *APIConfigStore) GetHighWatermark() int64 {
	cs.RLock()
	defer cs.RUnlock()

	return cs.highWatermark
}
