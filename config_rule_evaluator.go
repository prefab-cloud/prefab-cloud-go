package prefab

import (
	"math"
	"reflect"
	"strings"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/prefab-cloud/prefab-cloud-go/utils"
)

type ConditionMatch struct {
	isMatch                  bool
	match                    *prefabProto.ConfigValue
	rowIndex                 int
	conditionalValueIndex    int
	selectedConditionalValue *prefabProto.ConditionalValue
}

type ConfigRuleEvaluator struct {
	configStore          ConfigStoreGetter
	projectEnvIdSupplier ProjectEnvIdSupplier
}

func NewConfigRuleEvaluator(configStore ConfigStoreGetter, supplier ProjectEnvIdSupplier) *ConfigRuleEvaluator {
	return &ConfigRuleEvaluator{
		configStore:          configStore,
		projectEnvIdSupplier: supplier,
	}
}

func (cve *ConfigRuleEvaluator) EvaluateConfig(config *prefabProto.Config, contextSet ContextGetter) ConditionMatch {
	// find the right row for the env id, then the no-env id row
	// iterate over conditional values in rows
	// evaluate criterion
	noEnvRowIndex := 0
	envRow, envRowExists := rowWithMatchingEnvId(config, cve.projectEnvIdSupplier.GetProjectEnvId())
	if envRowExists {
		noEnvRowIndex = 1
		match := cve.EvaluateRow(envRow, contextSet, 0)
		if match.isMatch {
			return match
		}
	}
	noEnvRow, noEnvRowExists := rowWithoutEnvId(config)
	if noEnvRowExists {
		match := cve.EvaluateRow(noEnvRow, contextSet, noEnvRowIndex)
		if match.isMatch {
			return match
		}
	}
	return ConditionMatch{isMatch: false}
}

func (cve *ConfigRuleEvaluator) EvaluateRow(row *prefabProto.ConfigRow, contextSet ContextGetter, rowIndex int) ConditionMatch {
	conditionMatch := ConditionMatch{}
	conditionMatch.isMatch = false

	for conditionalValueIndex, conditionalValue := range row.Values {
		matchedValue, matched := cve.EvaluateConditionalValue(conditionalValue, contextSet)
		if matched {
			conditionMatch.isMatch = true
			conditionMatch.rowIndex = rowIndex
			conditionMatch.conditionalValueIndex = conditionalValueIndex
			conditionMatch.match = matchedValue
			break
		}
	}
	return conditionMatch
}

func (cve *ConfigRuleEvaluator) EvaluateConditionalValue(conditionalValue *prefabProto.ConditionalValue, contextSet ContextGetter) (*prefabProto.ConfigValue, bool) {
	for _, criterion := range conditionalValue.Criteria {
		if !cve.EvaluateCriterion(criterion, contextSet) {
			return nil, false
		}
	}
	// TODO handle weighted value here? or handle when we check if this is encrypted, env-var?
	return conditionalValue.GetValue(), true
}

func (cve *ConfigRuleEvaluator) EvaluateCriterion(criterion *prefabProto.Criterion, contextSet ContextGetter) bool {
	// get the value from context
	contextValue, contextValueExists := contextSet.GetValue(criterion.GetPropertyName())
	matchValue, _ := utils.UnpackConfigValue(criterion.GetValueToMatch())

	switch criterion.GetOperator() {
	case prefabProto.Criterion_NOT_SET:
		return false
	case prefabProto.Criterion_ALWAYS_TRUE:
		return true
	case prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF, prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF:
		if !contextValueExists {
			// If the property doesn't exist, return false for ENDS_WITH_ONE_OF,
			// true for DOES_NOT_END_WITH_ONE_OF.
			return criterion.GetOperator() == prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF
		}
		if stringContextValue, contextValueIsString := contextValue.(string); contextValueIsString {
			if stringListMatchValue, ok := matchValue.([]string); ok {
				// Check if the property value ends with any of the specified suffixes.
				matchFound := endsWithAny(stringListMatchValue, stringContextValue)
				// For ENDS_WITH_ONE_OF, return the result of endsWithAny directly.
				// For DOES_NOT_END_WITH_ONE_OF, return the negation.
				return matchFound == (criterion.GetOperator() == prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF)
			}
		}
		return criterion.GetOperator() == prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF
	case prefabProto.Criterion_PROP_IS_ONE_OF, prefabProto.Criterion_PROP_IS_NOT_ONE_OF:
		if stringContextValue, contextValueIsString := contextValue.(string); contextValueIsString {
			// Type assertion for matchValue as []string
			if stringSliceMatchValue, matchValueIsStringSlice := matchValue.([]string); matchValueIsStringSlice {
				// Check if stringContextValue is contained within stringSliceMatchValue
				return stringInSlice(stringContextValue, stringSliceMatchValue) == (criterion.GetOperator() == prefabProto.Criterion_PROP_IS_ONE_OF)
			}
		}
		return criterion.GetOperator() == prefabProto.Criterion_PROP_IS_NOT_ONE_OF
	case prefabProto.Criterion_HIERARCHICAL_MATCH:
		if stringContextValue, contextValueIsString := contextValue.(string); contextValueIsString {
			if stringMatchValue, matchValueIsString := matchValue.(string); matchValueIsString {
				return strings.HasPrefix(stringContextValue, stringMatchValue)
			}
		}
		return false
	case prefabProto.Criterion_IN_INT_RANGE:
		if !contextValueExists {
			return false
		}
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
		return false
	case prefabProto.Criterion_IN_SEG, prefabProto.Criterion_NOT_IN_SEG:
		if segmentName, segmentValueIsString := matchValue.(string); segmentValueIsString {
			targetConfig, configExists := cve.configStore.GetConfig(segmentName)
			if !configExists {
				return criterion.GetOperator() == prefabProto.Criterion_NOT_IN_SEG
			}
			match := cve.EvaluateConfig(targetConfig, contextSet)
			if match.isMatch {
				matchConfigValue := match.match
				if _, boolExists := matchConfigValue.GetType().(*prefabProto.ConfigValue_Bool); boolExists {
					return matchConfigValue.GetBool() == (criterion.GetOperator() == prefabProto.Criterion_IN_SEG)
				}
			}
			return criterion.GetOperator() == prefabProto.Criterion_NOT_IN_SEG
		}
	}

	return false
}

func defaultInt(maybeInt *int64, defaultValue int64) int64 {
	if maybeInt != nil {
		return *maybeInt
	}
	return defaultValue
}

func rowWithMatchingEnvId(config *prefabProto.Config, envId int64) (row *prefabProto.ConfigRow, rowWasFound bool) {
	for _, row := range config.GetRows() {
		if row.ProjectEnvId != nil && *row.ProjectEnvId == envId {
			return row, true
		}
	}
	return nil, false
}

func rowWithoutEnvId(config *prefabProto.Config) (row *prefabProto.ConfigRow, rowWasFound bool) {
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
