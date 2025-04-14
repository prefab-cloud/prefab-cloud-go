package internal

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/anyhelpers"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/semver"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/utils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ConditionMatch struct {
	Match                 *prefabProto.ConfigValue
	RowIndex              *int
	ConditionalValueIndex *int
	IsMatch               bool
}

type ConfigRuleEvaluator struct {
	configStore          ConfigStoreGetter
	projectEnvIDSupplier ProjectEnvIDSupplier
}

func NewConfigRuleEvaluator(configStore ConfigStoreGetter, projectEnvIDSupplier ProjectEnvIDSupplier) *ConfigRuleEvaluator {
	return &ConfigRuleEvaluator{
		configStore:          configStore,
		projectEnvIDSupplier: projectEnvIDSupplier,
	}
}

func (cve *ConfigRuleEvaluator) EvaluateConfig(config *prefabProto.Config, contextSet ContextValueGetter) ConditionMatch {
	// find the right row for the env id, then the no-env id row
	// iterate over conditional values in rows
	// evaluate criterion
	noEnvRowIndex := 0
	envRow, envRowExists := rowWithMatchingEnvID(config, cve.projectEnvIDSupplier.GetProjectEnvID())

	if envRowExists {
		noEnvRowIndex = 1

		match := cve.EvaluateRow(envRow, contextSet, 0)
		if match.IsMatch {
			return match
		}
	}

	noEnvRow, noEnvRowExists := rowWithoutEnvID(config)
	if noEnvRowExists {
		match := cve.EvaluateRow(noEnvRow, contextSet, noEnvRowIndex)
		if match.IsMatch {
			return match
		}
	}

	return ConditionMatch{IsMatch: false}
}

func (cve *ConfigRuleEvaluator) EvaluateRow(row *prefabProto.ConfigRow, contextSet ContextValueGetter, rowIndex int) ConditionMatch {
	conditionMatch := ConditionMatch{}
	conditionMatch.IsMatch = false

	for conditionalValueIndex, conditionalValue := range row.GetValues() {
		matchedValue, matched := cve.EvaluateConditionalValue(conditionalValue, contextSet)
		if matched {
			conditionMatch.IsMatch = true
			conditionMatch.RowIndex = &rowIndex
			conditionMatch.ConditionalValueIndex = &conditionalValueIndex
			conditionMatch.Match = matchedValue

			break
		}
	}

	return conditionMatch
}

func (cve *ConfigRuleEvaluator) EvaluateConditionalValue(conditionalValue *prefabProto.ConditionalValue, contextSet ContextValueGetter) (*prefabProto.ConfigValue, bool) {
	for _, criterion := range conditionalValue.GetCriteria() {
		if !cve.EvaluateCriterion(criterion, contextSet) {
			return nil, false
		}
	}

	return conditionalValue.GetValue(), true
}

func contextValueToString(contextValue interface{}) string {
	if contextValue == nil {
		return ""
	}

	return fmt.Sprintf("%v", contextValue)
}

func contextValueToStringSlice(contextValue interface{}) []string {
	if contextValue == nil {
		return nil
	}

	if stringArray, ok := contextValue.([]string); ok {
		return stringArray
	}

	if reflect.TypeOf(contextValue).Kind() == reflect.Slice {
		stringSlice := make([]string, 0)
		for _, value := range contextValue.([]interface{}) {
			stringSlice = append(stringSlice, contextValueToString(value))
		}

		return stringSlice
	}

	return []string{contextValueToString(contextValue)}
}

func (cve *ConfigRuleEvaluator) EvaluateCriterion(criterion *prefabProto.Criterion, contextSet ContextValueGetter) bool {
	// get the value from context
	contextValue, contextValueExists := contextSet.GetContextValue(criterion.GetPropertyName())

	// Special handling for "prefab.current-time" property
	if criterion.GetPropertyName() == "prefab.current-time" {
		// Create a ConfigValue with the current UTC time in milliseconds since epoch
		currentTimeMillis := time.Now().UTC().UnixMilli()
		contextValue = currentTimeMillis
		contextValueExists = true
	}

	matchValue, _, err := utils.ExtractValue(criterion.GetValueToMatch())

	switch criterion.GetOperator() {
	case prefabProto.Criterion_NOT_SET:
		return false
	case prefabProto.Criterion_ALWAYS_TRUE:
		return true
	case prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF, prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF:
		if err == nil && contextValueExists {
			stringContextValue := contextValueToString(contextValue)

			if stringListMatchValue, ok := matchValue.([]string); ok {
				// Check if the property value ends with any of the specified suffixes.
				matchFound := endsWithAny(stringListMatchValue, stringContextValue)
				// For ENDS_WITH_ONE_OF, return the result of endsWithAny directly.
				// For DOES_NOT_END_WITH_ONE_OF, return the negation.
				return matchFound == (criterion.GetOperator() == prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF)
			}
		}

		return criterion.GetOperator() == prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF
	case prefabProto.Criterion_PROP_STARTS_WITH_ONE_OF, prefabProto.Criterion_PROP_DOES_NOT_START_WITH_ONE_OF:
		if err == nil && contextValueExists {
			stringContextValue := contextValueToString(contextValue)

			if stringListMatchValue, ok := matchValue.([]string); ok {
				// Check if the property value ends with any of the specified suffixes.
				matchFound := startsWithAny(stringListMatchValue, stringContextValue)
				// For STARTS_WITH_ONE_OF, return the result of startsWithAny directly.
				// For DOES_NOT_START_WITH_ONE_OF, return the negation.
				return matchFound == (criterion.GetOperator() == prefabProto.Criterion_PROP_STARTS_WITH_ONE_OF)
			}
		}

		return criterion.GetOperator() == prefabProto.Criterion_PROP_DOES_NOT_START_WITH_ONE_OF
	case prefabProto.Criterion_PROP_CONTAINS_ONE_OF, prefabProto.Criterion_PROP_DOES_NOT_CONTAIN_ONE_OF:
		if err == nil && contextValueExists {
			stringContextValue := contextValueToString(contextValue)

			if stringListMatchValue, ok := matchValue.([]string); ok {
				// Check if the property value ends with any of the specified suffixes.
				matchFound := containsAny(stringListMatchValue, stringContextValue)
				// For CONTAINS_ONE_OF, return the result of containsAny directly.
				// For DOES_NOT_CONTAIN_ONE_OF, return the negation.
				return matchFound == (criterion.GetOperator() == prefabProto.Criterion_PROP_CONTAINS_ONE_OF)
			}
		}

		return criterion.GetOperator() == prefabProto.Criterion_PROP_DOES_NOT_CONTAIN_ONE_OF
	case prefabProto.Criterion_PROP_IS_ONE_OF, prefabProto.Criterion_PROP_IS_NOT_ONE_OF:
		if err == nil && contextValueExists {
			sliceContextValue := contextValueToStringSlice(contextValue)

			// Type assertion for matchValue as []string
			if stringSliceMatchValue, matchValueIsStringSlice := matchValue.([]string); matchValueIsStringSlice {
				matchFound := false

				for _, stringContextValue := range sliceContextValue {
					// Check if stringContextValue is contained within stringSliceMatchValue
					if stringInSlice(stringContextValue, stringSliceMatchValue) {
						matchFound = true

						break
					}
				}

				return matchFound == (criterion.GetOperator() == prefabProto.Criterion_PROP_IS_ONE_OF)
			}
		}

		return criterion.GetOperator() == prefabProto.Criterion_PROP_IS_NOT_ONE_OF
	case prefabProto.Criterion_HIERARCHICAL_MATCH:
		if err == nil && contextValueExists {
			stringContextValue := contextValueToString(contextValue)

			if stringMatchValue, matchValueIsString := matchValue.(string); matchValueIsString {
				return strings.HasPrefix(stringContextValue, stringMatchValue)
			}
		}

		return false
	case prefabProto.Criterion_IN_INT_RANGE:
		if err == nil && contextValueExists {
			if intRange := criterion.GetValueToMatch().GetIntRange(); intRange != nil {
				if reflectedContextValue := reflect.ValueOf(contextValue); reflectedContextValue.IsValid() {
					if reflectedContextValue.CanInt() {
						intContextValue := reflectedContextValue.Int()

						return intContextValue >= defaultInt(intRange.Start, int64(math.MinInt64)) &&
							intContextValue < defaultInt(intRange.End, int64(math.MaxInt64))
					} else if reflectedContextValue.CanFloat() {
						floatContextValue := reflectedContextValue.Float()

						return floatContextValue >= float64(defaultInt(intRange.Start, math.MinInt64)) &&
							floatContextValue < float64(defaultInt(intRange.End, math.MaxInt64))
					}
				}
			}
		}

		return false
	case prefabProto.Criterion_IN_SEG, prefabProto.Criterion_NOT_IN_SEG:
		if err == nil {
			if segmentName, segmentValueIsString := matchValue.(string); segmentValueIsString {
				targetConfig, configExists := cve.configStore.GetConfig(segmentName)
				if !configExists {
					return criterion.GetOperator() == prefabProto.Criterion_NOT_IN_SEG
				}

				match := cve.EvaluateConfig(targetConfig, contextSet)
				if match.IsMatch {
					matchConfigValue := match.Match
					if _, boolExists := matchConfigValue.GetType().(*prefabProto.ConfigValue_Bool); boolExists {
						return matchConfigValue.GetBool() == (criterion.GetOperator() == prefabProto.Criterion_IN_SEG)
					}
				}
			}
		}

		return criterion.GetOperator() == prefabProto.Criterion_NOT_IN_SEG

	case prefabProto.Criterion_PROP_GREATER_THAN, prefabProto.Criterion_PROP_GREATER_THAN_OR_EQUAL, prefabProto.Criterion_PROP_LESS_THAN, prefabProto.Criterion_PROP_LESS_THAN_OR_EQUAL:
		if err == nil && contextValueExists && anyhelpers.IsNumber(contextValue) && anyhelpers.IsNumber(matchValue) {
			comparisonResult, comparisonErr := compareNumbers(contextValue, matchValue)
			if comparisonErr == nil {
				if handler, ok := numericCriterionHandlers[criterion.GetOperator()]; ok {
					return handler(comparisonResult)
				}
			}
		}

	case prefabProto.Criterion_PROP_BEFORE, prefabProto.Criterion_PROP_AFTER:
		if err == nil && contextValueExists && (anyhelpers.IsNumber(contextValue) || anyhelpers.IsString(contextValue)) && (anyhelpers.IsNumber(matchValue) || anyhelpers.IsString(matchValue)) {
			contextTimeMillis, contextTimeMillisErr := dateToMillis(contextValue)
			matchValueTimeMillis, matchValueTimeMillisErr := dateToMillis(matchValue)
			if contextTimeMillisErr == nil && matchValueTimeMillisErr == nil {
				return (criterion.GetOperator() == prefabProto.Criterion_PROP_AFTER && contextTimeMillis > matchValueTimeMillis) || (criterion.GetOperator() == prefabProto.Criterion_PROP_BEFORE && contextTimeMillis < matchValueTimeMillis)
			}
		}
	case prefabProto.Criterion_PROP_MATCHES, prefabProto.Criterion_PROP_DOES_NOT_MATCH:
		if err == nil && contextValueExists && anyhelpers.IsString(contextValue) && anyhelpers.IsString(matchValue) {
			if matchvalueRegex, matchValueRegexErr := regexp.Compile(matchValue.(string)); matchValueRegexErr == nil {
				matched := matchvalueRegex.MatchString(contextValue.(string))
				return matched == (criterion.GetOperator() == prefabProto.Criterion_PROP_MATCHES)
			}
		}
	case prefabProto.Criterion_PROP_SEMVER_LESS_THAN, prefabProto.Criterion_PROP_SEMVER_EQUAL, prefabProto.Criterion_PROP_SEMVER_GREATER_THAN:
		if err == nil && contextValueExists && anyhelpers.IsString(contextValue) && anyhelpers.IsString(matchValue) {
			semanticVersionFromMatch := semver.ParseQuietly(matchValue.(string))
			semanticVersionFromContext := semver.ParseQuietly(contextValue.(string))
			if semanticVersionFromMatch != nil && semanticVersionFromContext != nil {
				comparisonResult := semanticVersionFromContext.Compare(*semanticVersionFromMatch)
				if handler, ok := semverCriterionHandlers[criterion.GetOperator()]; ok {
					return handler(comparisonResult)
				}
			}

		}
	}
	return false
}

func dateToMillis(val any) (int64, error) {
	if anyhelpers.IsNumber(val) {
		if int64Value, err := anyhelpers.ToInt64(val); err == nil {
			return int64Value, nil
		} else {
			return 0, err
		}
	}
	if anyhelpers.IsString(val) {
		if parsedDate, err := parseDate(val.(string)); err == nil {
			return parsedDate.UnixMilli(), nil
		} else {
			return 0, err
		}
	}

	return 0, fmt.Errorf("unsupported type: %v", reflect.TypeOf(val))
}

var semverCriterionHandlers = map[prefabProto.Criterion_CriterionOperator]func(int) bool{
	prefabProto.Criterion_PROP_SEMVER_GREATER_THAN: func(result int) bool {
		return result > 0
	},
	prefabProto.Criterion_PROP_SEMVER_EQUAL: func(result int) bool {
		return result == 0
	},
	prefabProto.Criterion_PROP_SEMVER_LESS_THAN: func(result int) bool {
		return result < 0
	},
}

var numericCriterionHandlers = map[prefabProto.Criterion_CriterionOperator]func(int) bool{
	prefabProto.Criterion_PROP_GREATER_THAN: func(result int) bool {
		return result > 0
	},
	prefabProto.Criterion_PROP_GREATER_THAN_OR_EQUAL: func(result int) bool {
		return result >= 0
	},
	prefabProto.Criterion_PROP_LESS_THAN: func(result int) bool {
		return result < 0
	},
	prefabProto.Criterion_PROP_LESS_THAN_OR_EQUAL: func(result int) bool {
		return result <= 0
	},
}

func compareNumbers(a, b any) (int, error) {
	// Convert both numbers to float64 for comparison
	aFloat, err := anyhelpers.ToFloat64(a)
	if err != nil {
		return 0, err
	}
	bFloat, err := anyhelpers.ToFloat64(b)
	if err != nil {
		return 0, err
	}

	// Perform comparison with compareTo semantics
	if aFloat < bFloat {
		return -1, nil
	} else if aFloat > bFloat {
		return 1, nil
	}
	return 0, nil
}

func defaultInt(maybeInt *int64, defaultValue int64) int64 {
	if maybeInt != nil {
		return *maybeInt
	}

	return defaultValue
}

func rowWithMatchingEnvID(config *prefabProto.Config, envID int64) (*prefabProto.ConfigRow, bool) {
	for _, row := range config.GetRows() {
		if row.ProjectEnvId != nil && row.GetProjectEnvId() == envID {
			return row, true
		}
	}

	return nil, false
}

func rowWithoutEnvID(config *prefabProto.Config) (*prefabProto.ConfigRow, bool) {
	for _, row := range config.GetRows() {
		if row.ProjectEnvId == nil {
			return row, true
		}
	}

	return nil, false
}

func endsWithAny(suffixes []string, target string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(target, suffix) {
			return true // Found a suffix that matches.
		}
	}

	return false // No matching suffix found.
}

func startsWithAny(suffixes []string, target string) bool {
	for _, suffix := range suffixes {
		if strings.HasPrefix(target, suffix) {
			return true // Found a suffix that matches.
		}
	}

	return false // No matching suffix found.
}

func containsAny(suffixes []string, target string) bool {
	for _, suffix := range suffixes {
		if strings.Contains(target, suffix) {
			return true // Found a suffix that matches.
		}
	}

	return false // No matching suffix found.
}

// Helper function to check if a string is in a slice of strings.
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

func parseDate(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", value)
}
