package anyhelpers_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/anyhelpers"
)

func TestDetectAndReturnStringListIfPresent(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []string
		wantOk   bool
	}{
		{
			name:     "Slice of strings",
			input:    []string{"one", "two", "three"},
			expected: []string{"one", "two", "three"},
			wantOk:   true,
		},
		{
			name:     "Slice of interfaces with all strings",
			input:    []interface{}{"one", "two", "three"},
			expected: []string{"one", "two", "three"},
			wantOk:   true,
		},
		{
			name:     "Slice of interfaces with not all strings",
			input:    []interface{}{"one", 2, "three"},
			expected: nil,
			wantOk:   false,
		},
		{
			name:     "Not a slice",
			input:    "not a slice",
			expected: nil,
			wantOk:   false,
		},
		{
			name:     "Slice of integers",
			input:    []int{1, 2, 3},
			expected: nil,
			wantOk:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := anyhelpers.DetectAndReturnStringListIfPresent(tc.input)
			if !reflect.DeepEqual(got, tc.expected) || ok != tc.wantOk {
				t.Errorf("DetectAndReturnStringListIfPresent(%v) = %v, %v; want %v, %v", tc.input, got, ok, tc.expected, tc.wantOk)
			}
		})
	}
}

func TestCanonicalizeSlice(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected interface{}
		name     string
		wantOk   bool
	}{
		{
			name:     "All strings",
			input:    []string{"apple", "banana", "cherry"},
			expected: []string{"apple", "banana", "cherry"},
			wantOk:   true,
		},
		{
			name:     "All integers",
			input:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
			wantOk:   true,
		},
		{
			name:     "All floats",
			input:    []float64{1.1, 2.2, 3.3},
			expected: []float64{1.1, 2.2, 3.3},
			wantOk:   true,
		},
		{
			name:     "Slice of interfaces all strings",
			input:    []interface{}{"apple", "banana", "cherry"},
			expected: []string{"apple", "banana", "cherry"},
			wantOk:   true,
		},
		{
			name:     "Mixed types",
			input:    []interface{}{"apple", 2, "cherry"},
			expected: []interface{}{"apple", 2, "cherry"},
			wantOk:   false,
		},
		{
			name:     "Empty slice of interfaces",
			input:    []interface{}{},
			expected: []interface{}{},
			wantOk:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := anyhelpers.CanonicalizeSlice(tc.input)
			if ok != tc.wantOk || !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("canonicalize(%v) got %v, %v; want %v, %v", tc.input, got, ok, tc.expected, tc.wantOk)
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"Int", 42, true},
		{"Int8", int8(42), true},
		{"Int16", int16(42), true},
		{"Int32", int32(42), true},
		{"Int64", int64(42), true},
		{"Uint", uint(42), true},
		{"Uint8", uint8(42), true},
		{"Uint16", uint16(42), true},
		{"Uint32", uint32(42), true},
		{"Uint64", uint64(42), true},
		{"Float32", float32(42.0), true},
		{"Float64", float64(42.0), true},
		{"Negative Int", -42, true},
		{"Negative Float", -42.0, true},
		{"Zero Int", 0, true},
		{"Zero Float", 0.0, true},
		{"String", "42", false},
		{"Bool", true, false},
		{"Nil", nil, false},
		{"Struct", struct{}{}, false},
		{"Slice", []int{42}, false},
		{"Map", map[string]int{"key": 42}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := anyhelpers.IsNumber(tt.input)
			if result != tt.expected {
				t.Errorf("IsNumber(%v) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to compare floats with a small tolerance
func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) <= epsilon
}

func TestToFloat64(t *testing.T) {
	const epsilon = 1e-4 // Tolerance for floating-point comparisons

	tests := []struct {
		name      string
		input     any
		expected  float64
		expectErr bool
	}{
		{"Int", 42, 42.0, false},
		{"Int8", int8(42), 42.0, false},
		{"Int16", int16(42), 42.0, false},
		{"Int32", int32(42), 42.0, false},
		{"Int64", int64(42), 42.0, false},
		{"Uint", uint(42), 42.0, false},
		{"Uint8", uint8(42), 42.0, false},
		{"Uint16", uint16(42), 42.0, false},
		{"Uint32", uint32(42), 42.0, false},
		{"Uint64", uint64(42), 42.0, false},
		{"Float32", float32(42.42), 42.42, false},
		{"Float64", float64(42.42), 42.42, false},
		{"Negative Int", -42, -42.0, false},
		{"Negative Float", -42.42, -42.42, false},
		{"Zero Int", 0, 0.0, false},
		{"Zero Float", 0.0, 0.0, false},
		{"String", "42", 0.0, true},
		{"Bool", true, 0.0, true},
		{"Nil", nil, 0.0, true},
		{"Struct", struct{}{}, 0.0, true},
		{"Slice", []int{42}, 0.0, true},
		{"Map", map[string]int{"key": 42}, 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := anyhelpers.ToFloat64(tt.input)
			if (err != nil) != tt.expectErr {
				t.Fatalf("ToFloat64(%v) error = %v; expectErr = %v", tt.input, err, tt.expectErr)
			}
			if !tt.expectErr && !almostEqual(result, tt.expected, epsilon) {
				t.Errorf("ToFloat64(%v) = %v; expected %v (±%v)", tt.input, result, tt.expected, epsilon)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  int64
		expectErr bool
	}{
		{"Int", 42, 42, false},
		{"Int8", int8(42), 42, false},
		{"Int16", int16(42), 42, false},
		{"Int32", int32(42), 42, false},
		{"Int64", int64(42), 42, false},
		{"Negative Int", -42, -42, false},
		{"Uint", uint(42), 42, false},
		{"Uint8", uint8(42), 42, false},
		{"Uint16", uint16(42), 42, false},
		{"Uint32", uint32(42), 42, false},
		{"Uint64", uint64(42), 42, false},
		{"Uint64 Too Large", uint64(math.MaxUint64), 0, true},
		{"Float32", float32(42.99), 42, false}, // Truncate float
		{"Float64", float64(42.99), 42, false}, // Truncate float
		{"Negative Float", -42.99, -42, false},
		{"Zero Int", 0, 0, false},
		{"Zero Float", 0.0, 0, false},
		{"String", "42", 0, true},
		{"Bool", true, 0, true},
		{"Nil", nil, 0, true},
		{"Struct", struct{}{}, 0, true},
		{"Slice", []int{42}, 0, true},
		{"Map", map[string]int{"key": 42}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := anyhelpers.ToInt64(tt.input)
			if (err != nil) != tt.expectErr {
				t.Fatalf("ToInt64(%v) error = %v; expectErr = %v", tt.input, err, tt.expectErr)
			}
			if !tt.expectErr && result != tt.expected {
				t.Errorf("ToInt64(%v) = %v; expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
