package utils

import (
	"errors"
	"fmt"
	"github.com/prefab-cloud/prefab-cloud-go/anyhelpers"
	durationParser "github.com/sosodev/duration"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

func Create(value any) (*prefabProto.ConfigValue, bool) {
	// Check if value is already a *prefabProto.ConfigValue
	if cv, ok := value.(*prefabProto.ConfigValue); ok {
		return cv, true // Return early if it's already the correct type
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
	case []any:
		stringList, ok := anyhelpers.DetectAndReturnStringListIfPresent(v)
		if ok {
			configValue.Type = &prefabProto.ConfigValue_StringList{StringList: &prefabProto.StringList{Values: stringList}}
		}
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
		configValue.Type = &prefabProto.ConfigValue_Duration{Duration: &prefabProto.IsoDuration{Definition: durationToISO8601(v)}}
	case *any:
		if v != nil {
			return Create(*v)
		}
		slog.Warn("create unable to handle nil")
		return nil, false
	default:
		slog.Warn(fmt.Sprintf("Unsupported type: %T", value))
		return nil, false
	}

	return configValue, true
}

// ExtractValue returns the value of the oneof field and a boolean indicating
// if it's one of the simple types (fields 1-5, 10). returns nil, false if the configvalue is nil
func ExtractValue(cv *prefabProto.ConfigValue) (value any, simpleTypeAvailable bool, err error) {
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
	case *prefabProto.ConfigValue_Provided:
		val, ok := handleProvided(v.Provided)
		return val, ok, nil
	default:
		// For other types, return the protobuf value itself and false.
		return v, false, nil
	}
}

func handleProvided(provided *prefabProto.Provided) (value string, ok bool) {
	switch provided.GetSource() {
	case prefabProto.ProvidedSource_ENV_VAR:
		if provided.Lookup != nil {
			envValue, envValueExists := os.LookupEnv(provided.GetLookup())
			return envValue, envValueExists
		}
	}

	return "", false
}

// GetValueType returns the value type we'd expect the config containing this value to have
func GetValueType(cv *prefabProto.ConfigValue) prefabProto.Config_ValueType {
	if cv == nil {
		return prefabProto.Config_NOT_SET_VALUE_TYPE
	}

	switch cv.Type.(type) {
	case *prefabProto.ConfigValue_Int:
		return prefabProto.Config_INT
	case *prefabProto.ConfigValue_String_:
		return prefabProto.Config_STRING
	case *prefabProto.ConfigValue_Bytes:
		return prefabProto.Config_BYTES
	case *prefabProto.ConfigValue_Double:
		return prefabProto.Config_DOUBLE
	case *prefabProto.ConfigValue_Bool:
		return prefabProto.Config_BOOL
	case *prefabProto.ConfigValue_StringList:
		// StringList is considered a simple type for this example, returning the slice of strings directly.
		return prefabProto.Config_STRING_LIST
	case *prefabProto.ConfigValue_LogLevel:
		return prefabProto.Config_LOG_LEVEL
	case *prefabProto.ConfigValue_Duration:
		return prefabProto.Config_DURATION
	case *prefabProto.ConfigValue_Json:
		return prefabProto.Config_JSON
	}
	//For other types, return the protobuf value itself and false.
	return prefabProto.Config_NOT_SET_VALUE_TYPE
}

type ExtractValueFunction func(*prefabProto.ConfigValue) (any, bool)

func ExtractIntValue(cv *prefabProto.ConfigValue) (int64, bool) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Int:
		return v.Int, true
	default:
		return 0, false
	}
}

func ExtractStringValue(cv *prefabProto.ConfigValue) (string, bool) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_String_:
		return v.String_, true
	default:
		return "", false
	}
}

func ExtractFloatValue(cv *prefabProto.ConfigValue) (float64, bool) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Double:
		return v.Double, true
	default:
		return 0, false
	}
}

func ExtractBoolValue(cv *prefabProto.ConfigValue) (bool, bool) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Bool:
		return v.Bool, true
	default:
		return false, false
	}
}

func ExtractStringListValue(cv *prefabProto.ConfigValue) ([]string, bool) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_StringList:
		return v.StringList.Values, true
	default:
		var zero []string
		return zero, false
	}
}

func ExtractDurationValue(cv *prefabProto.ConfigValue) (time.Duration, bool) {
	switch v := cv.Type.(type) {
	case *prefabProto.ConfigValue_Duration:
		{

			duration, err := durationParser.Parse(v.Duration.Definition)
			if err != nil {
				slog.Debug(fmt.Sprintf("Failed to parse duration value: %s", v.Duration.Definition))
				return time.Duration(0), false
			}
			return duration.ToTimeDuration(), true
		}
	default:
		var zero time.Duration
		return zero, false
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

func durationToISO8601(d time.Duration) string {
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
