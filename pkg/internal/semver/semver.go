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

// isNumeric returns true if the string contains only digits
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// comparePreReleaseIdentifiers compares individual prerelease identifiers
func comparePreReleaseIdentifiers(id1, id2 string) int {
	// If both are numeric, compare numerically
	if isNumeric(id1) && isNumeric(id2) {
		num1, _ := strconv.Atoi(id1)
		num2, _ := strconv.Atoi(id2)
		if num1 < num2 {
			return -1
		}
		if num1 > num2 {
			return 1
		}
		return 0
	}

	// If only one is numeric, numeric ones have lower precedence
	if isNumeric(id1) {
		return -1
	}
	if isNumeric(id2) {
		return 1
	}

	// Neither is numeric, compare as strings
	if id1 < id2 {
		return -1
	}
	if id1 > id2 {
		return 1
	}
	return 0
}

// comparePreRelease compares prerelease strings according to semver rules
func comparePreRelease(pre1, pre2 string) int {
	// If both are empty, they're equal
	if pre1 == "" && pre2 == "" {
		return 0
	}

	// A version without prerelease has higher precedence
	if pre1 == "" {
		return 1
	}
	if pre2 == "" {
		return -1
	}

	// Split into identifiers
	ids1 := strings.Split(pre1, ".")
	ids2 := strings.Split(pre2, ".")

	// Compare each identifier until we find a difference
	minLen := len(ids1)
	if len(ids2) < minLen {
		minLen = len(ids2)
	}

	for i := 0; i < minLen; i++ {
		if cmp := comparePreReleaseIdentifiers(ids1[i], ids2[i]); cmp != 0 {
			return cmp
		}
	}

	// If all identifiers match up to the length of the shorter one,
	// the longer one has higher precedence
	if len(ids1) < len(ids2) {
		return -1
	}
	if len(ids1) > len(ids2) {
		return 1
	}
	return 0
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

	return comparePreRelease(s.prerelease, other.prerelease)
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
