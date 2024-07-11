package internal

import prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"

type ContextValueGetter interface {
	GetContextValue(propertyName string) (value interface{}, valueExists bool)
}

// TODO: add a `freshen` or similar method
type ConfigStoreGetter interface {
	GetConfig(key string) (config *prefabProto.Config, exists bool)
	Keys() []string
	ContextValueGetter
	ProjectEnvIDSupplier
}

type ProjectEnvIDSupplier interface {
	GetProjectEnvID() int64
}

type ConfigEvaluator interface {
	EvaluateConfig(config *prefabProto.Config, contextSet ContextValueGetter) (match ConditionMatch)
}

type Decrypter interface {
	DecryptValue(secretKey string, value string) (decryptedValue string, err error)
}

type Randomer interface {
	Float64() float64
}

type Hasher interface {
	HashZeroToOne(value string) (zeroToOne float64, ok bool)
}

type WeightedValueResolverIF interface {
	Resolve(weightedValues *prefabProto.WeightedValues, propertyName string, contextGetter ContextValueGetter) (valueResult *prefabProto.ConfigValue, index int)
}
type EnvLookup interface {
	LookupEnv(key string) (string, bool)
}
