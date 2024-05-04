package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
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
	case time.Duration:
		configValue.Type = &prefabProto.ConfigValue_Duration{Duration: &prefabProto.IsoDuration{Definition: DurationToISO8601(v)}}
	case *interface{}:
		if v != nil {
			return CreateConfigValue(*v)
		}
		return nil
	default:
		fmt.Printf("Unsupported type: %v\n", reflect.TypeOf(value))
	}

	return configValue
}

// UnpackConfigValue returns the value of the oneof field and a boolean indicating
// if it's one of the simple types (fields 1-5, 10). returns nil, false if the configvalue is nil
func UnpackConfigValue(cv *prefabProto.ConfigValue) (value interface{}, simpleTypeAvailable bool, err error) {
	if cv == nil {
		return nil, false, nil
	}

	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Int:
		return v.Int, true, nil
	case *prefabProto.ConfigValue_String_:
		return v.String_, true, nil
	case *prefabProto.ConfigValue_Bytes:
		return v.Bytes, true, nil // Bytes are returned as []byte, which is a simple Go type
	case *prefabProto.ConfigValue_Double:
		return v.Double, true, nil
	case *prefabProto.ConfigValue_Bool:
		return v.Bool, true, nil
	case *prefabProto.ConfigValue_StringList:
		// StringList is considered a simple type, returning the slice of strings directly.
		return v.StringList.Values, true, nil
	case *prefabProto.ConfigValue_Duration:
		duration, err := time.ParseDuration(v.Duration.Definition)
		if err != nil {
			return duration, true, nil
		}
		return nil, false, err

	default:
		// For other types, return the protobuf value itself and false.
		return v, false, nil
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
	case *prefabProto.ConfigValue_Duration:
		return prefabProto.Config_DURATION, nil
	default:
		// For other types, return the protobuf value itself and false.
		return prefabProto.Config_NOT_SET_VALUE_TYPE, nil
	}
}

type ConfigValueParseFunction func(*prefabProto.ConfigValue) (interface{}, error)

func ParseIntValue(cv *prefabProto.ConfigValue) (int64, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Int:
		return v.Int, nil
	default:
		return 0, errors.New("config did not produce the correct type for Int")
	}
}

func ParseStringValue(cv *prefabProto.ConfigValue) (string, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_String_:
		return v.String_, nil
	default:
		return "", errors.New("config did not produce the correct type for String")
	}
}

func ParseFloatValue(cv *prefabProto.ConfigValue) (float64, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Double:
		return v.Double, nil
	default:
		return 0, errors.New("config did not produce the correct type for Float")
	}
}

func ParseBoolValue(cv *prefabProto.ConfigValue) (bool, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Bool:
		return v.Bool, nil
	default:
		return false, errors.New("config did not produce the correct type for Bool")
	}
}

func ParseStringListValue(cv *prefabProto.ConfigValue) ([]string, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_StringList:
		return v.StringList.Values, nil
	default:
		return nil, errors.New("config did not produce the correct type for StringList")
	}
}

func ParseDurationValue(cv *prefabProto.ConfigValue) (time.Duration, error) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Duration:
		{
			duration, err := time.ParseDuration(v.Duration.Definition)
			if err != nil {
				return time.Duration(0), err
			}
			return duration, nil
		}
	default:
		return time.Duration(0), errors.New("config did not produce the correct type for Duration")
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

func DurationToISO8601(d time.Duration) string {
	var (
		years   = int64(d / (365 * 24 * time.Hour))
		months  = int64(d % (365 * 24 * time.Hour) / (30 * 24 * time.Hour))
		days    = int64(d % (30 * 24 * time.Hour) / (24 * time.Hour))
		hours   = int64(d % (24 * time.Hour) / time.Hour)
		minutes = int64(d % time.Hour / time.Minute)
		seconds = float64(d%time.Minute) / float64(time.Second)
	)

	var sb strings.Builder
	sb.WriteString("P")

	if years > 0 {
		sb.WriteString(strconv.FormatInt(years, 10))
		sb.WriteString("Y")
	}
	if months > 0 {
		sb.WriteString(strconv.FormatInt(months, 10))
		sb.WriteString("M")
	}
	if days > 0 {
		sb.WriteString(strconv.FormatInt(days, 10))
		sb.WriteString("D")
	}

	hasTimePart := hours > 0 || minutes > 0 || seconds > 0
	if hasTimePart {
		sb.WriteString("T")
	}

	if hours > 0 {
		sb.WriteString(strconv.FormatInt(hours, 10))
		sb.WriteString("H")
	}
	if minutes > 0 {
		sb.WriteString(strconv.FormatInt(minutes, 10))
		sb.WriteString("M")
	}
	if seconds > 0 {
		if hours == 0 && minutes == 0 {
			sb.WriteString(fmt.Sprintf("%.9f", seconds))
			sb.WriteString("S")
		} else {
			sb.WriteString(fmt.Sprintf("%02.0f", seconds))
			sb.WriteString("S")
		}
	} else if hasTimePart {
		sb.WriteString("0S")
	}

	return sb.String()
}
