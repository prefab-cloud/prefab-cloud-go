package internal

import (
	"fmt"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"reflect"
)

func CreateConfigValue(value interface{}) *prefabProto.ConfigValue {
	// Check if value is already a *prefabProto.ConfigValue
	if cv, ok := value.(*prefabProto.ConfigValue); ok {
		return cv // Return early if it's already the correct type
	}

	configValue := &prefabProto.ConfigValue{}

	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		val := reflect.ValueOf(value)
		configValue.Type = &prefabProto.ConfigValue_Int{Int: val.Int()}
	case string:
		configValue.Type = &prefabProto.ConfigValue_String_{String_: v}
	case []string:
		// Assuming ConfigValue has a StringList field to handle a list of strings
		configValue.Type = &prefabProto.ConfigValue_StringList{StringList: &prefabProto.StringList{Values: v}}
	case []byte:
		configValue.Type = &prefabProto.ConfigValue_Bytes{Bytes: v}
	case bool:
		configValue.Type = &prefabProto.ConfigValue_Bool{Bool: v}
	case float64, float32:
		val := reflect.ValueOf(value)
		configValue.Type = &prefabProto.ConfigValue_Double{Double: val.Float()}
	default:
		fmt.Printf("Unsupported type: %v\n", reflect.TypeOf(value))
	}

	return configValue
}

// UnpackConfigValue returns the value of the oneof field and a boolean indicating
// if it's one of the simple types (fields 1-5, 10). returns nil, false if the configvalue is nil
func UnpackConfigValue(cv *prefabProto.ConfigValue) (interface{}, bool) {
	if cv == nil {
		return nil, false
	}

	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Int:
		return v.Int, true
	case *prefabProto.ConfigValue_String_:
		return v.String_, true
	case *prefabProto.ConfigValue_Bytes:
		return v.Bytes, true // Bytes are returned as []byte, which is a simple Go type
	case *prefabProto.ConfigValue_Double:
		return v.Double, true
	case *prefabProto.ConfigValue_Bool:
		return v.Bool, true
	case *prefabProto.ConfigValue_StringList:
		// StringList is considered a simple type for this example, returning the slice of strings directly.
		return v.StringList.Values, true
	default:
		// For other types, return the protobuf value itself and false.
		return v, false
	}
}
