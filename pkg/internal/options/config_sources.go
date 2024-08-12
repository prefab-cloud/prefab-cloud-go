package options

import (
	"fmt"
	"strings"
)

type ConfigSource struct {
	Store   StoreType
	Raw     string
	Path    string
	Default bool
}

type StoreType string

const (
	APIStore   StoreType = "API"
	DataFile   StoreType = "DataFile"
	ConfigDump StoreType = "ConfigDump"
	Memory     StoreType = "Memory"
	// TODO: Support polling
	// Poll       StoreType = "Poll"

	MemoryStoreKey = "memory://configs"
)

func GetDefaultConfigSources() []ConfigSource {
	return []ConfigSource{
		{Raw: "api:prefab", Store: APIStore, Default: true},
	}
}

func ParseConfigSource(rawSource string) (ConfigSource, error) {
	parts := strings.Split(rawSource, "://")

	if len(parts) < 2 {
		return ConfigSource{}, fmt.Errorf("invalid source (protocol is missing): %s", rawSource)
	}

	protocol := parts[0]
	path := parts[1]

	switch protocol {
	case "datafile":
		return ConfigSource{Raw: rawSource, Store: DataFile, Default: false, Path: path}, nil
	case "dump":
		return ConfigSource{Raw: rawSource, Store: ConfigDump, Default: false, Path: path}, nil
	case "memory":
		return ConfigSource{Raw: rawSource, Store: Memory, Default: false, Path: path}, nil
	}

	return ConfigSource{}, fmt.Errorf("unknown protocol %s", protocol)
}
