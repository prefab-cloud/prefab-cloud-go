package internal

import (
	"fmt"

	opts "github.com/prefab-cloud/prefab-cloud-go/pkg/options"
)

func BuildConfigStore(options opts.Options, source opts.ConfigSource, apiSourceFinishedLoading func()) (ConfigStoreGetter, bool, error) {
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
	default:
		return nil, false, fmt.Errorf("unknown store type %v", source.Store)
	}
}
