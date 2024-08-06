package stores

import (
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

// CompositeConfigStore is a config store that composes multiple config stores.
// It attempts to find a config by key in each store in the order they were provided.
// The first store that contains the requested config is used, and the search stops.
// If no store contains the requested config, the CompositeConfigStore returns false
// to indicate that the config does not exist.
type CompositeConfigStore struct {
	stores []internal.ConfigStoreGetter
}

func BuildCompositeConfigStore(stores ...internal.ConfigStoreGetter) *CompositeConfigStore {
	return &CompositeConfigStore{
		stores: stores,
	}
}

func (s *CompositeConfigStore) GetConfig(key string) (*prefabProto.Config, bool) {
	for _, store := range s.stores {
		config, exists := store.GetConfig(key)
		if exists {
			return config, true
		}
	}

	return nil, false
}

func (s *CompositeConfigStore) GetContextValue(propertyName string) (interface{}, bool) {
	for _, store := range s.stores {
		value, valueExists := store.GetContextValue(propertyName)

		if valueExists {
			return value, valueExists
		}
	}

	return nil, false
}

func (s *CompositeConfigStore) Keys() []string {
	uniqueKeys := make(map[string]struct{})

	for _, store := range s.stores {
		for _, key := range store.Keys() {
			uniqueKeys[key] = struct{}{}
		}
	}

	keys := make([]string, 0, len(uniqueKeys))
	for key := range uniqueKeys {
		keys = append(keys, key)
	}

	return keys
}

func (s *CompositeConfigStore) GetProjectEnvID() int64 {
	for _, store := range s.stores {
		projectEnvID := store.GetProjectEnvID()

		if projectEnvID != 0 {
			return projectEnvID
		}
	}

	return 0
}
