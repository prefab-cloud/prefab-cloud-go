package internal

import (
	"fmt"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"reflect"
)

func CreateConfigValue(value interface{}) *prefabProto.ConfigValue {
	configValue := &prefabProto.ConfigValue{}

	switch v := value.(type) {
	case int:
		configValue.Type = &prefabProto.ConfigValue_Int{Int: int64(v)}
	case int8:
		configValue.Type = &prefabProto.ConfigValue_Int{Int: int64(v)}
	case int16:
		configValue.Type = &prefabProto.ConfigValue_Int{Int: int64(v)}
	case int32:
		configValue.Type = &prefabProto.ConfigValue_Int{Int: int64(v)}
	case int64:
		configValue.Type = &prefabProto.ConfigValue_Int{Int: v}
	case string:
		configValue.Type = &prefabProto.ConfigValue_String_{String_: v}
	case bool:
		configValue.Type = &prefabProto.ConfigValue_Bool{Bool: v}
	case float64:
		configValue.Type = &prefabProto.ConfigValue_Double{Double: v}
	case float32:
		configValue.Type = &prefabProto.ConfigValue_Double{Double: float64(v)}
	default:
		panic(fmt.Sprintf("Unsupported type: %v\n", reflect.TypeOf(value)))
	}

	return configValue
}
