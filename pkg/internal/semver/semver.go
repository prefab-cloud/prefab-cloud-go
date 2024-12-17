package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var semverPattern = regexp.MustCompile(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

type SemanticVersion struct {
	major         int
	minor         int
	patch         int
	prerelease    string
	buildMetadata string
}

// ParseQuietly attempts to parse a version string, returning nil if parsing fails
func ParseQuietly(version string) *SemanticVersion {
	semver, err := Parse(version)
	if err != nil {
		return nil
	}
	return semver
}

// Parse creates a new SemanticVersion from a version string
func Parse(version string) (*SemanticVersion, error) {
	if version == "" {
		return nil, fmt.Errorf("version string cannot be empty")
	}

	matches := findNamedMatches(semverPattern, version)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid semantic version format: %s", version)
	}

	major, err := strconv.Atoi(matches["major"])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", matches["major"])
	}

	minor, err := strconv.Atoi(matches["minor"])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", matches["minor"])
	}

	patch, err := strconv.Atoi(matches["patch"])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", matches["patch"])
	}

	return &SemanticVersion{
		major:         major,
		minor:         minor,
		patch:         patch,
		prerelease:    matches["prerelease"],
		buildMetadata: matches["buildmetadata"],
	}, nil
}

// Helper function to find named matches in regex
func findNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)
	if match == nil {
		return map[string]string{}
	}

	result := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}

// Getters - using value receivers since they don't modify state
func (s SemanticVersion) Major() int            { return s.major }
func (s SemanticVersion) Minor() int            { return s.minor }
func (s SemanticVersion) Patch() int            { return s.patch }
func (s SemanticVersion) Prerelease() string    { return s.prerelease }
func (s SemanticVersion) BuildMetadata() string { return s.buildMetadata }

// Compare implements comparison between two semantic versions
// Returns -1 if s < other, 0 if s == other, and 1 if s > other
func (s SemanticVersion) Compare(other SemanticVersion) int {
	if s.major != other.major {
		if s.major > other.major {
			return 1
		}
		return -1
	}

	if s.minor != other.minor {
		if s.minor > other.minor {
			return 1
		}
		return -1
	}

	if s.patch != other.patch {
		if s.patch > other.patch {
			return 1
		}
		return -1
	}

	// If one has prerelease and the other doesn't, the one without prerelease is greater
	if s.prerelease == "" && other.prerelease != "" {
		return 1
	}
	if s.prerelease != "" && other.prerelease == "" {
		return -1
	}
	if s.prerelease != "" && other.prerelease != "" {
		if s.prerelease > other.prerelease {
			return 1
		}
		if s.prerelease < other.prerelease {
			return -1
		}
	}

	return 0 // Build metadata doesn't affect precedence
}

func (s SemanticVersion) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d.%d.%d", s.major, s.minor, s.patch)
	if s.prerelease != "" {
		fmt.Fprintf(&sb, "-%s", s.prerelease)
	}
	if s.buildMetadata != "" {
		fmt.Fprintf(&sb, "+%s", s.buildMetadata)
	}
	return sb.String()
}

// Equal compares two semantic versions for equality
func (s SemanticVersion) Equal(other SemanticVersion) bool {
	return s.major == other.major &&
		s.minor == other.minor &&
		s.patch == other.patch &&
		s.prerelease == other.prerelease
	// Build metadata is ignored in equality checks
}
