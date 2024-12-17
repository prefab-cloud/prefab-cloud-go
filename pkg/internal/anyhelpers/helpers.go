package anyhelpers

import (
	"errors"
	"fmt"
	"math"
	"reflect"
)

// DetectAndReturnStringListIfPresent checks if the input interface{} is a slice with only string elements
// and returns the slice and a bool indicating if a string slice was found.
func DetectAndReturnStringListIfPresent(input interface{}) ([]string, bool) {
	canonical, ok := CanonicalizeSlice(input)
	if !ok {
		return nil, false
	}

	if strSlice, ok := canonical.([]string); ok {
		return strSlice, true
	}

	return nil, false
}

// CanonicalizeSlice checks if the input is a slice and if all elements are of the same type.
// If so, it returns the slice or converts to a more specific slice type, otherwise returns the original slice.
func CanonicalizeSlice(input any) (any, bool) {
	v := reflect.ValueOf(input)
	if v.Kind() != reflect.Slice {
		return input, false
	}

	if v.Len() == 0 {
		return input, false // Empty slice, no type uniformity to check
	}

	elementType := v.Type().Elem()
	if elementType.Kind() == reflect.Interface {
		// Handle []interface{} by checking each element's type
		firstElemType := reflect.TypeOf(v.Index(0).Interface())
		for i := 1; i < v.Len(); i++ {
			if reflect.TypeOf(v.Index(i).Interface()) != firstElemType {
				return input, false // Not all elements are of the same type
			}
		}
		// Convert all to the discovered type
		result := reflect.MakeSlice(reflect.SliceOf(firstElemType), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			result.Index(i).Set(reflect.ValueOf(v.Index(i).Interface()).Convert(firstElemType))
		}

		return result.Interface(), true
	}

	return input, true // Return as is if already a specific slice type
}

func IsNumber(val any) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}

func ToFloat64(val any) (float64, error) {
	switch v := val.(type) {
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Uint()), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		if v, ok := val.(float64); ok {
			return v, nil
		}
	}
	return 0, fmt.Errorf("unsupported type: %v", reflect.TypeOf(val))
}

func ToInt64(val any) (int64, error) {
	switch v := val.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		u := reflect.ValueOf(v).Uint()
		if u > math.MaxInt64 { // Use constant from math package
			return 0, errors.New("value too large to fit in int64")
		}
		return int64(u), nil
	case float32, float64:
		return int64(reflect.ValueOf(v).Float()), nil // Truncate float
	default:
		return 0, fmt.Errorf("unsupported type: %v", reflect.TypeOf(val))
	}
}

func IsString(val any) bool {
	_, ok := val.(string)
	return ok
}
