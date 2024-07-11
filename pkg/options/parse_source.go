package options

type ConfigSource struct {
	// TODO: we can do better here
	Store   interface{}
	Raw     string
	Default bool
}

type StoreType string

const (
	APIStore StoreType = "API"
	DataFile StoreType = "DataFile"
	Poll     StoreType = "Poll"
	// TODO: more? maybe something for the Proto/ConfigWrapper store?
)

var DefaultConfigSources = []ConfigSource{
	{Raw: "sse:prefab", Store: APIStore, Default: true},
	{Raw: "https://api.prefab.cloud", Store: APIStore, Default: true},
}

func ParseConfigSource(rawSource string) (ConfigSource, error) {
	// TODO: validate this and figure out the store
	// We should probably check that a file exists for datafile, that the URL is valid for poll, etc.
	return ConfigSource{Raw: rawSource, Default: false}, nil
}
