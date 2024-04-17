package internal

import (
	"errors"
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
	case *prefabProto.ConfigValue_LogLevel:
		configValue.Type = v
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

// ValueTypeFromConfigValue returns the value type we'd expect the config containing this value to have
func ValueTypeFromConfigValue(cv *prefabProto.ConfigValue) (prefabProto.Config_ValueType, error) {
	if cv == nil {
		return prefabProto.Config_NOT_SET_VALUE_TYPE, errors.New("config value was nil")
	}

	switch cv.Type.(type) {
	case *prefabProto.ConfigValue_Int:
		return prefabProto.Config_INT, nil
	case *prefabProto.ConfigValue_String_:
		return prefabProto.Config_STRING, nil
	case *prefabProto.ConfigValue_Bytes:
		return prefabProto.Config_BYTES, nil
	case *prefabProto.ConfigValue_Double:
		return prefabProto.Config_DOUBLE, nil
	case *prefabProto.ConfigValue_Bool:
		return prefabProto.Config_BOOL, nil
	case *prefabProto.ConfigValue_StringList:
		// StringList is considered a simple type for this example, returning the slice of strings directly.
		return prefabProto.Config_STRING_LIST, nil
	case *prefabProto.ConfigValue_LogLevel:
		return prefabProto.Config_LOG_LEVEL, nil
	default:
		// For other types, return the protobuf value itself and false.
		return prefabProto.Config_NOT_SET_VALUE_TYPE, nil
	}
}

type ConfigValueParseFunction func(*prefabProto.ConfigValue) (interface{}, error)

func ParseIntValue(cv *prefabProto.ConfigValue) (interface{}, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Int:
		return v.Int, nil
	default:
		return nil, errors.New("config did not produce the correct type for Int")
	}
}

func ParseStringValue(cv *prefabProto.ConfigValue) (interface{}, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_String_:
		return v.String_, nil
	default:
		return nil, errors.New("config did not produce the correct type for String")
	}
}

func ParseFloatValue(cv *prefabProto.ConfigValue) (interface{}, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Double:
		return v.Double, nil
	default:
		return nil, errors.New("config did not produce the correct type for Float")
	}
}

func ParseBoolValue(cv *prefabProto.ConfigValue) (interface{}, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Bool:
		return v.Bool, nil
	default:
		return nil, errors.New("config did not produce the correct type for Bool")
	}
}

func parseStringListValue(cv *prefabProto.ConfigValue) (interface{}, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_StringList:
		return v.StringList.Values, nil
	default:
		return nil, errors.New("config did not produce the correct type for StringList")
	}
}

func parseBytesValue(cv *prefabProto.ConfigValue) (interface{}, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Bytes:
		return v.Bytes, nil
	default:
		return nil, errors.New("config did not produce the correct type for Bytes")
	}
}
