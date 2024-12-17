package semver

import (
	"testing"
)

func TestParseValidVersions(t *testing.T) {
	tests := []struct {
		input string
		major int
		minor int
		patch int
		pre   string
		build string
	}{
		{"1.2.3", 1, 2, 3, "", ""},
		{"0.0.0", 0, 0, 0, "", ""},
		{"1.2.3-alpha", 1, 2, 3, "alpha", ""},
		{"1.2.3+build.123", 1, 2, 3, "", "build.123"},
		{"1.2.3-alpha.1+build.123", 1, 2, 3, "alpha.1", "build.123"},
		{"10.20.30", 10, 20, 30, "", ""},
		{"1.0.0-alpha.1", 1, 0, 0, "alpha.1", ""},
		{"1.0.0-0.3.7", 1, 0, 0, "0.3.7", ""},
		{"1.0.0-x.7.z.92", 1, 0, 0, "x.7.z.92", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := Parse(tt.input)
			if err != nil {
				t.Errorf("Parse(%q) returned unexpected error: %v", tt.input, err)
			}
			if v.Major() != tt.major {
				t.Errorf("Major = %d; want %d", v.Major(), tt.major)
			}
			if v.Minor() != tt.minor {
				t.Errorf("Minor = %d; want %d", v.Minor(), tt.minor)
			}
			if v.Patch() != tt.patch {
				t.Errorf("Patch = %d; want %d", v.Patch(), tt.patch)
			}
			if v.Prerelease() != tt.pre {
				t.Errorf("Prerelease = %q; want %q", v.Prerelease(), tt.pre)
			}
			if v.BuildMetadata() != tt.build {
				t.Errorf("BuildMetadata = %q; want %q", v.BuildMetadata(), tt.build)
			}
		})
	}
}

func TestParseInvalidVersions(t *testing.T) {
	invalidVersions := []string{
		"",
		"1",
		"1.2",
		"1.2.3.4",
		"1.2.3-",
		"1.2.3+",
		"01.2.3",
		"1.02.3",
		"1.2.03",
		"-1.2.3",
		"1.-2.3",
		"1.2.-3",
	}

	for _, v := range invalidVersions {
		t.Run(v, func(t *testing.T) {
			if _, err := Parse(v); err == nil {
				t.Errorf("Parse(%q) succeeded, want error", v)
			}
		})
	}
}

func TestParseQuietly(t *testing.T) {
	tests := []struct {
		input   string
		wantNil bool
	}{
		{"1.2.3", false},
		{"invalid", true},
		{"", true},
		{"1.2.3-alpha+build", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseQuietly(tt.input)
			if (result == nil) != tt.wantNil {
				t.Errorf("ParseQuietly(%q) returned nil: %v, want nil: %v",
					tt.input, result == nil, tt.wantNil)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.2.3", "1.2.3", 0},
		{"2.0.0", "1.9.9", 1},
		{"1.2.3", "1.2.4", -1},
		{"1.2.3-alpha", "1.2.3", -1},
		{"1.2.3", "1.2.3-beta", 1},
		{"1.2.3-alpha", "1.2.3-beta", -1},
		{"1.0.0+build.1", "1.0.0+build.2", 0},
		{"2.1.1", "2.1.0", 1},
		{"2.1.1", "2.2.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_"+tt.v2, func(t *testing.T) {
			v1, err1 := Parse(tt.v1)
			v2, err2 := Parse(tt.v2)
			if err1 != nil || err2 != nil {
				t.Fatalf("Failed to parse test versions: %v, %v", err1, err2)
			}

			result := v1.Compare(*v2)
			if result != tt.expected {
				t.Errorf("Compare(%q, %q) = %d; want %d",
					tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected bool
	}{
		{"1.2.3", "1.2.3", true},
		{"1.2.3-alpha", "1.2.3-alpha", true},
		{"1.2.3+build.1", "1.2.3+build.2", true},
		{"1.2.3-alpha+build.1", "1.2.3-alpha+build.2", true},
		{"1.2.3", "1.2.4", false},
		{"1.2.3-alpha", "1.2.3-beta", false},
		{"1.2.3", "1.2.3-alpha", false},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_"+tt.v2, func(t *testing.T) {
			v1, err1 := Parse(tt.v1)
			v2, err2 := Parse(tt.v2)
			if err1 != nil || err2 != nil {
				t.Fatalf("Failed to parse test versions: %v, %v", err1, err2)
			}

			result := v1.Equal(*v2)
			if result != tt.expected {
				t.Errorf("Equal(%q, %q) = %v; want %v",
					tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1.2.3", "1.2.3"},
		{"1.2.3-alpha", "1.2.3-alpha"},
		{"1.2.3+build.123", "1.2.3+build.123"},
		{"1.2.3-alpha.1+build.123", "1.2.3-alpha.1+build.123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse version: %v", err)
			}

			if got := v.String(); got != tt.want {
				t.Errorf("String() = %q; want %q", got, tt.want)
			}
		})
	}
}
