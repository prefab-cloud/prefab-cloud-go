package prefab

import (
	"github.com/prefab-cloud/prefab-cloud-go/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigRuleTestSuite struct {
	suite.Suite
	contextSet *ContextSet
	evaluator  ConfigRuleEvaluator
}

func (suite *ConfigRuleTestSuite) SetupSuite() {
	suite.contextSet = NewContextSet()
	suite.contextSet.SetNamedContext(NewNamedContextWithValues("user", map[string]interface{}{
		"key":   "u123",
		"email": "me@example.com",
		"admin": true,
		"age":   int64(42),
	}))
	suite.contextSet.SetNamedContext(NewNamedContextWithValues("team", map[string]interface{}{
		"key":  "t123",
		"name": "dev ops",
	}))
	suite.contextSet.SetNamedContext(NewNamedContextWithValues("", map[string]interface{}{
		"key": "?234",
		"id":  int64(3456),
	}))
}

func (suite *ConfigRuleTestSuite) TestAlwaysTrueCriteria() {
	suite.Run("ends with one of returns true", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_ALWAYS_TRUE}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().True(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestPropEndsWithCriteriaEvaluation() {
	suite.Run("ends with one of returns true for email ending with example.com", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF, ValueToMatch: internal.CreateConfigValue([]string{"example.com", "splat"}), PropertyName: "user.email"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().True(isMatch)
	})

	suite.Run("ends with one of returns false for email ending with prefab.cloud", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF, ValueToMatch: internal.CreateConfigValue([]string{"prefab.cloud", "splat"}), PropertyName: "user.email"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().False(isMatch)
	})

	suite.Run("ends with one of returns false for property that does not exist", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF, ValueToMatch: internal.CreateConfigValue([]string{"prefab.cloud", "splat"}), PropertyName: "not.here"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().False(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestPropNotEndsWithCriteriaEvaluation() {
	suite.Run("ends with one of returns false for email ending with example.com", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF, ValueToMatch: internal.CreateConfigValue([]string{"example.com", "splat"}), PropertyName: "user.email"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().False(isMatch)
	})

	suite.Run("ends with one of returns true for email ending with prefab.cloud", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF, ValueToMatch: internal.CreateConfigValue([]string{"prefab.cloud", "splat"}), PropertyName: "user.email"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().True(isMatch)
	})

	suite.Run("ends with one of returns true for property that does not exist", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF, ValueToMatch: internal.CreateConfigValue([]string{"prefab.cloud", "splat"}), PropertyName: "not.here"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().True(isMatch)
	})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestConfigRuleTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigRuleTestSuite))
}
