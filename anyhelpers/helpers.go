package anyhelpers

import "reflect"

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

// Generic function to get the pointer of any type
func Ptr[T any](v T) *T {
	return &v
}
