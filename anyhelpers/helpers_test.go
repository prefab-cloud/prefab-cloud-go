package anyhelpers

import (
	"reflect"
	"testing"
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
			got, ok := DetectAndReturnStringListIfPresent(tc.input)
			if !reflect.DeepEqual(got, tc.expected) || ok != tc.wantOk {
				t.Errorf("DetectAndReturnStringListIfPresent(%v) = %v, %v; want %v, %v", tc.input, got, ok, tc.expected, tc.wantOk)
			}
		})
	}
}

func TestCanonicalizeSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
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
			got, ok := CanonicalizeSlice(tc.input)
			if ok != tc.wantOk || !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("canonicalize(%v) got %v, %v; want %v, %v", tc.input, got, ok, tc.expected, tc.wantOk)
			}
		})
	}
}
