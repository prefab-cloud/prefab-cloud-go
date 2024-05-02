package prefab

import prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"

// CompositeConfigStore is a config store that composes multiple config stores.
// It attempts to find a config by key in each store in the order they were provided.
// The first store that contains the requested config is used, and the search stops.
// If no store contains the requested config, the CompositeConfigStore returns false
// to indicate that the config does not exist.
type CompositeConfigStore struct {
	stores []ConfigStoreGetter
}

func BuildCompositeConfigStore(stores ...ConfigStoreGetter) *CompositeConfigStore {
	return &CompositeConfigStore{
		stores: stores,
	}
}

func (s *CompositeConfigStore) GetConfig(key string) (config *prefabProto.Config, exists bool) {
	for _, store := range s.stores {
		config, exists = store.GetConfig(key)
		if exists {
			return config, true
		}
	}
	return nil, false
}
