package internal

import (
	"math"
	"reflect"
	"strings"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/utils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ConditionMatch struct {
	Match                    *prefabProto.ConfigValue
	SelectedConditionalValue *prefabProto.ConditionalValue
	RowIndex                 int
	ConditionalValueIndex    int
	IsMatch                  bool
}

type ConfigRuleEvaluator struct {
	configStore          ConfigStoreGetter
	projectEnvIDSupplier ProjectEnvIDSupplier
}

func NewConfigRuleEvaluator(configStore ConfigStoreGetter, supplier ProjectEnvIDSupplier) *ConfigRuleEvaluator {
	return &ConfigRuleEvaluator{
		configStore:          configStore,
		projectEnvIDSupplier: supplier,
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
			conditionMatch.RowIndex = rowIndex
			conditionMatch.ConditionalValueIndex = conditionalValueIndex
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

func (cve *ConfigRuleEvaluator) EvaluateCriterion(criterion *prefabProto.Criterion, contextSet ContextValueGetter) bool {
	// get the value from context
	contextValue, contextValueExists := contextSet.GetContextValue(criterion.GetPropertyName())
	matchValue, _, err := utils.ExtractValue(criterion.GetValueToMatch())

	switch criterion.GetOperator() {
	case prefabProto.Criterion_NOT_SET:
		return false
	case prefabProto.Criterion_ALWAYS_TRUE:
		return true
	case prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF, prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF:
		if err == nil && contextValueExists {
			if stringContextValue, contextValueIsString := contextValue.(string); contextValueIsString {
				if stringListMatchValue, ok := matchValue.([]string); ok {
					// Check if the property value ends with any of the specified suffixes.
					matchFound := endsWithAny(stringListMatchValue, stringContextValue)
					// For ENDS_WITH_ONE_OF, return the result of endsWithAny directly.
					// For DOES_NOT_END_WITH_ONE_OF, return the negation.
					return matchFound == (criterion.GetOperator() == prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF)
				}
			}
		}

		return criterion.GetOperator() == prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF
	case prefabProto.Criterion_PROP_IS_ONE_OF, prefabProto.Criterion_PROP_IS_NOT_ONE_OF:
		if err == nil && contextValueExists {
			if stringContextValue, contextValueIsString := contextValue.(string); contextValueIsString {
				// Type assertion for matchValue as []string
				if stringSliceMatchValue, matchValueIsStringSlice := matchValue.([]string); matchValueIsStringSlice {
					// Check if stringContextValue is contained within stringSliceMatchValue
					return stringInSlice(stringContextValue, stringSliceMatchValue) == (criterion.GetOperator() == prefabProto.Criterion_PROP_IS_ONE_OF)
				}
			}
		}

		return criterion.GetOperator() == prefabProto.Criterion_PROP_IS_NOT_ONE_OF
	case prefabProto.Criterion_HIERARCHICAL_MATCH:
		if err == nil && contextValueExists {
			if stringContextValue, contextValueIsString := contextValue.(string); contextValueIsString {
				if stringMatchValue, matchValueIsString := matchValue.(string); matchValueIsString {
					return strings.HasPrefix(stringContextValue, stringMatchValue)
				}
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
	}

	return false
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

// Helper function to check if a string is in a slice of strings.
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}
