package prefab

import (
	"github.com/prefab-cloud/prefab-cloud-go/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"strings"
)

type ConditionMatch struct {
	isMatch               bool
	match                 *prefabProto.ConfigValue
	rowIndex              int
	conditionalvalueIndex int
}

type ConfigRuleEvaluator struct {
}

func (cve *ConfigRuleEvaluator) Evaluate(config *prefabProto.Config, contextSet *ContextSet, projectEnvId int64) ConditionMatch {

	// find the right row for the env id, then the no-env id row
	// iterate over conditional values in rows
	// evaluate criterion
	noEnvRowIndex := 0
	envRow, envRowExists := rowWithMatchingEnvId(config, projectEnvId)
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

func (cve *ConfigRuleEvaluator) EvaluateRow(row *prefabProto.ConfigRow, contextSet *ContextSet, rowIndex int) ConditionMatch {
	conditionMatch := ConditionMatch{}
	conditionMatch.isMatch = false

	for conditionalValueIndex, conditionalValue := range row.Values {
		matchedValue, matched := cve.EvaluateConditionalValue(conditionalValue, contextSet)
		if matched {
			conditionMatch.isMatch = true
			conditionMatch.rowIndex = rowIndex
			conditionMatch.conditionalvalueIndex = conditionalValueIndex
			conditionMatch.match = matchedValue
		}
	}
	return conditionMatch
}

func (cve *ConfigRuleEvaluator) EvaluateConditionalValue(conditionalValue *prefabProto.ConditionalValue, contextSet *ContextSet) (*prefabProto.ConfigValue, bool) {
	for _, criterion := range conditionalValue.Criteria {
		if !cve.EvaluateCriterion(criterion, contextSet) {
			return nil, false
		}
	}
	// TODO handle weighted value here? or handle when we check if this is encrypted, env-var?
	return conditionalValue.GetValue(), true
}

func (cve *ConfigRuleEvaluator) EvaluateCriterion(criterion *prefabProto.Criterion, contextSet *ContextSet) bool {
	// get the value from context
	propertyValue, exists := contextSet.getValue(criterion.GetPropertyName())
	matchValue, _ := internal.UnpackConfigValue(criterion.GetValueToMatch())

	switch criterion.GetOperator() {
	case prefabProto.Criterion_ALWAYS_TRUE:
		return true
	case prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF, prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF:
		if !exists {
			// If the property doesn't exist, return false for ENDS_WITH_ONE_OF,
			// true for DOES_NOT_END_WITH_ONE_OF.
			return criterion.GetOperator() == prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF
		}
		if stringPropertyValue, propValueOk := propertyValue.(string); propValueOk {
			if stringListMatchValue, ok := matchValue.([]string); ok {
				// Check if the property value ends with any of the specified suffixes.
				matchFound := endsWithAny(stringListMatchValue, stringPropertyValue)
				// For ENDS_WITH_ONE_OF, return the result of endsWithAny directly.
				// For DOES_NOT_END_WITH_ONE_OF, return the negation.
				return matchFound == (criterion.GetOperator() == prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF)
			}
		}
	}
	return false
}

func rowWithMatchingEnvId(config *prefabProto.Config, envId int64) (*prefabProto.ConfigRow, bool) {
	for _, row := range config.GetRows() {
		if row.ProjectEnvId != nil && *row.ProjectEnvId == envId {
			return row, true
		}
	}
	return nil, false
}

func rowWithoutEnvId(config *prefabProto.Config) (*prefabProto.ConfigRow, bool) {
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
