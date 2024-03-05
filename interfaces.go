package prefab

import prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"

type ContextGetter interface {
	GetValue(propertyName string) (value interface{}, valueExists bool)
}

type ConfigStoreGetter interface {
	GetConfig(key string) (config *prefabProto.Config, exists bool)
}

type Decrypter interface {
	DecryptValue(secretKeyString, value string) (decryptedValue string, err error)
}

type Randomer interface {
	Float64() float64
}

type Hasher interface {
	HashZeroToOne(value string) (zeroToOne float64, ok bool)
}
