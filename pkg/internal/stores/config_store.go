package stores

import (
	"fmt"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	opts "github.com/prefab-cloud/prefab-cloud-go/pkg/internal/options"
)

func BuildConfigStore(options opts.Options, source opts.ConfigSource, apiSourceFinishedLoading func()) (internal.ConfigStoreGetter, bool, error) {
	switch source.Store {
	case opts.APIStore:
		store, err := NewAPIConfigStore(options, apiSourceFinishedLoading)

		return store, true, err
	case opts.DataFile:
		// TODO: handle JSON here too (based on file extension)
		store, err := NewLocalConfigStore(source.Path)

		return store, false, err
	case opts.ConfigDump:
		store, err := NewConfigDumpConfigStore(source.Path, options.ProjectEnvID)

		return store, false, err
	case opts.Memory:
		store, err := NewMemoryConfigStore(options.ProjectEnvID, options.Configs)

		return store, false, err
	default:
		return nil, false, fmt.Errorf("unknown store type %v", source.Store)
	}
}
