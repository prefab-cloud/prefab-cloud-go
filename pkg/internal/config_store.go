package internal

import (
	"fmt"

	opts "github.com/prefab-cloud/prefab-cloud-go/pkg/options"
)

func BuildConfigStore(options opts.Options, source opts.ConfigSource, apiSourceFinishedLoading func()) (ConfigStoreGetter, error) {
	switch source.Store {
	case opts.APIStore:
		return NewAPIConfigStore(options, apiSourceFinishedLoading)
	case opts.DataFile:
		// TODO: handle JSON here too (based on file extension)
		return NewLocalConfigStore(source.Path)
	case opts.ConfigDump:
		return NewConfigDumpConfigStore(source.Path)
	default:
		return nil, fmt.Errorf("unknown store type %v", source.Store)
	}
}
