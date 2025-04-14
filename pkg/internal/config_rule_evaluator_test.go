package internal_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/mocks"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/testutils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"

	"github.com/stretchr/testify/suite"
)

const teamSize = "team.size"

type MockContextGetter struct {
	mock.Mock
}

func (m *MockContextGetter) GetContextValue(propertyName string) (interface{}, bool) {
	var value interface{}

	args := m.Called(propertyName)
	if args.Get(0) != nil {
		value = args.Get(0)
	} else {
		value = nil
	}

	valueExists := args.Bool(1)

	return value, valueExists
}

type ConfigRuleTestSuite struct {
	suite.Suite
	evaluator                 *internal.ConfigRuleEvaluator
	nonBoolReturnSimpleConfig *prefabProto.Config
	mockConfigStoreGetter     mocks.MockConfigStoreGetter
	mockProjectEnvIDSupplier  mocks.MockProjectEnvIDSupplier
	projectEnvID              int64
}

func (suite *ConfigRuleTestSuite) SetupTest() {
	suite.mockConfigStoreGetter = mocks.MockConfigStoreGetter{}
	suite.projectEnvID = 101
	suite.mockProjectEnvIDSupplier = *mocks.NewMockProjectEnvIDSupplier(suite.projectEnvID)
	suite.evaluator = internal.NewConfigRuleEvaluator(&suite.mockConfigStoreGetter, &suite.mockProjectEnvIDSupplier)
}

func (suite *ConfigRuleTestSuite) TestFullRuleEvaluation() {
	matchingProjectEnvID := int64(101)
	mismatchingProjectEnvID := int64(102)
	departmentNameEndsWithIngCriterion := &prefabProto.Criterion{
		Operator:     prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF,
		ValueToMatch: testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"ing"}),
		PropertyName: "department.name",
	}

	departmentNameIsComplianceOrSecurity := &prefabProto.Criterion{
		Operator:     prefabProto.Criterion_PROP_IS_ONE_OF,
		ValueToMatch: testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"compliance", "security"}),
		PropertyName: "department.name",
	}

	departmentNameIsAliens := &prefabProto.Criterion{
		Operator:     prefabProto.Criterion_PROP_IS_ONE_OF,
		ValueToMatch: testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"aliens"}),
		PropertyName: "department.name",
	}

	securityClearance := &prefabProto.Criterion{
		Operator:     prefabProto.Criterion_PROP_IS_ONE_OF,
		ValueToMatch: testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"top secret", "super top secret"}),
		PropertyName: "security.clearance",
	}

	config := &prefabProto.Config{
		Id:              1,
		ProjectId:       1,
		Key:             "test.rule",
		ConfigType:      prefabProto.ConfigType_CONFIG,
		ValueType:       prefabProto.Config_INT,
		SendToClientSdk: false,
		Rows: []*prefabProto.ConfigRow{
			{
				ProjectEnvId: internal.Int64Ptr(matchingProjectEnvID),
				Values: []*prefabProto.ConditionalValue{
					{
						Criteria: []*prefabProto.Criterion{departmentNameEndsWithIngCriterion, securityClearance},
						Value:    testutils.CreateConfigValueAndAssertOk(suite.T(), 1),
					},
					{
						Criteria: []*prefabProto.Criterion{departmentNameEndsWithIngCriterion},
						Value:    testutils.CreateConfigValueAndAssertOk(suite.T(), 2),
					},
					{
						Criteria: []*prefabProto.Criterion{departmentNameIsComplianceOrSecurity},
						Value:    testutils.CreateConfigValueAndAssertOk(suite.T(), 3),
					},
				},
			},
			{
				Values: []*prefabProto.ConditionalValue{
					{
						Criteria: []*prefabProto.Criterion{departmentNameIsAliens},
						Value:    testutils.CreateConfigValueAndAssertOk(suite.T(), 10),
					},
					{
						Value: testutils.CreateConfigValueAndAssertOk(suite.T(), 11),
					},
				},
			},
		},
	}

	tests := []struct {
		name                          string
		projectEnvID                  int64
		expectedValue                 *prefabProto.ConfigValue
		expectedRowIndex              int
		expectedConditionalValueIndex int
		contextMockings               []ContextMocking
	}{
		{"returns 1 for high security clearance mining department with matching projectEnv", matchingProjectEnvID, testutils.CreateConfigValueAndAssertOk(suite.T(), 1), 0, 0, []ContextMocking{{contextPropertyName: "department.name", value: "mining", exists: true}, {contextPropertyName: "security.clearance", value: "top secret", exists: true}}},
		{"returns 2 for no security clearance mining department with matching projectEnv", matchingProjectEnvID, testutils.CreateConfigValueAndAssertOk(suite.T(), 2), 0, 1, []ContextMocking{{contextPropertyName: "department.name", value: "mining", exists: true}, {contextPropertyName: "security.clearance", value: nil, exists: false}}},
		{"returns 3 for security department with matching projectEnv", matchingProjectEnvID, testutils.CreateConfigValueAndAssertOk(suite.T(), 3), 0, 2, []ContextMocking{{contextPropertyName: "department.name", value: "security", exists: true}}},
		{"returns 10 for aliens department with matching projectEnv", matchingProjectEnvID, testutils.CreateConfigValueAndAssertOk(suite.T(), 10), 1, 0, []ContextMocking{{contextPropertyName: "department.name", value: "aliens", exists: true}}},
		{"returns 11 for cleanup department with matching projectEnv", matchingProjectEnvID, testutils.CreateConfigValueAndAssertOk(suite.T(), 11), 1, 1, []ContextMocking{{contextPropertyName: "department.name", value: "cleanup", exists: true}}},
		{"returns 10 for aliens department with mismatching projectEnv", mismatchingProjectEnvID, testutils.CreateConfigValueAndAssertOk(suite.T(), 10), 0, 0, []ContextMocking{{contextPropertyName: "department.name", value: "aliens", exists: true}}},
		{"returns 11 for cleanup department with mismatching projectEnv", mismatchingProjectEnvID, testutils.CreateConfigValueAndAssertOk(suite.T(), 11), 0, 1, []ContextMocking{{contextPropertyName: "department.name", value: "cleanup", exists: true}}},
		{"returns 11 for high security clearance mining department with mismatching projectEnv", mismatchingProjectEnvID, testutils.CreateConfigValueAndAssertOk(suite.T(), 11), 0, 1, []ContextMocking{{contextPropertyName: "department.name", value: "mining", exists: true}, {contextPropertyName: "security.clearance", value: "top secret", exists: true}}},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			suite.evaluator = internal.NewConfigRuleEvaluator(&suite.mockConfigStoreGetter, mocks.NewMockProjectEnvIDSupplier(testCase.projectEnvID))
			mockContext, _ := suite.setupMockContextWithMultipleValues(testCase.contextMockings)
			conditionMatch := suite.evaluator.EvaluateConfig(config, mockContext)
			suite.Equal(testCase.expectedValue, conditionMatch.Match, "value should match")
			suite.Equal(testCase.expectedRowIndex, *conditionMatch.RowIndex, "rowIndex should match")
			suite.Equal(testCase.expectedConditionalValueIndex, *conditionMatch.ConditionalValueIndex, "conditionalValueIndex should match")
		})
	}
}

func (suite *ConfigRuleTestSuite) TestAlwaysTrueCriteria() {
	suite.Run("returns true", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_ALWAYS_TRUE}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, contexts.NewContextSet())
		suite.True(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestNotSetCriteria() {
	suite.Run("returns false", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_NOT_SET}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, contexts.NewContextSet())
		suite.False(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestPropEndsWithCriteriaEvaluation() {
	operator := prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF
	contextPropertyName := "user.email"
	defaultValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"example.com", "splat"})

	tests := []struct {
		name               string
		valueToMatch       *prefabProto.ConfigValue
		contextValue       interface{}
		contextValueExists bool
		expected           bool
	}{
		{"returns true for email ending with example.com", defaultValueToMatch, "me@example.com", true, true},
		{"returns false for email ending with prefab.cloud", defaultValueToMatch, "me@prefab.cloud", true, false},
		{"returns false when context does not exist", defaultValueToMatch, nil, false, false},
		{"returns false when valueToMatch is not a string slice", testutils.CreateConfigValueAndAssertOk(suite.T(), "example.com"), "doesn't matter", true, false},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()

			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: testCase.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestPropNotEndsWithCriteriaEvaluation() {
	operator := prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF
	contextPropertyName := "user.email"
	defaultValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"example.com", "splat"})

	tests := []struct {
		name               string
		valueToMatch       *prefabProto.ConfigValue
		contextValue       interface{}
		contextValueExists bool
		expected           bool
	}{
		{"returns false for email ending with example.com", defaultValueToMatch, "me@example.com", true, false},
		{"returns true for email ending with prefab.cloud", defaultValueToMatch, "me@prefab.cloud", true, true},
		{"returns true when context does not exist", defaultValueToMatch, nil, false, true},
		{"returns true when valueToMatch is not a string slice", testutils.CreateConfigValueAndAssertOk(suite.T(), "example.com"), "doesn't matter", true, true},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()

			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: testCase.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestPropStartsWithCriteriaEvaluation() {
	operator := prefabProto.Criterion_PROP_STARTS_WITH_ONE_OF
	contextPropertyName := "user.testProperty"
	defaultValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"one", "two"})

	tests := []struct {
		name               string
		valueToMatch       *prefabProto.ConfigValue
		contextValue       interface{}
		contextValueExists bool
		expected           bool
	}{
		{"returns true for testProperty starting with one", defaultValueToMatch, "one tree", true, true},
		{"returns false for testProperty starting with three", defaultValueToMatch, "three dogs", true, false},
		{"returns false when context does not exist", defaultValueToMatch, nil, false, false},
		{"returns false when valueToMatch is not a string slice", testutils.CreateConfigValueAndAssertOk(suite.T(), "example.com"), "doesn't matter", true, false},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()

			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: testCase.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestPropDoesNotStartsWithCriteriaEvaluation() {
	operator := prefabProto.Criterion_PROP_DOES_NOT_START_WITH_ONE_OF
	contextPropertyName := "user.testProperty"
	defaultValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"one", "two"})

	tests := []struct {
		name               string
		valueToMatch       *prefabProto.ConfigValue
		contextValue       interface{}
		contextValueExists bool
		expected           bool
	}{
		{"returns false for testProperty does not start with one", defaultValueToMatch, "one tree", true, false},
		{"returns true for testProperty does not start with three", defaultValueToMatch, "three dogs", true, true},
		{"returns true when context does not exist", defaultValueToMatch, nil, false, true},
		{"returns true when valueToMatch is not a string slice", testutils.CreateConfigValueAndAssertOk(suite.T(), "example.com"), "doesn't matter", true, true},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()

			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: testCase.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestPropContainsCriteriaEvaluation() {
	operator := prefabProto.Criterion_PROP_CONTAINS_ONE_OF
	contextPropertyName := "user.testProperty"
	defaultValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"one", "two"})

	tests := []struct {
		name               string
		valueToMatch       *prefabProto.ConfigValue
		contextValue       interface{}
		contextValueExists bool
		expected           bool
	}{
		{"returns true for testProperty contains one", defaultValueToMatch, "just one tree", true, true},
		{"returns false for testProperty contains three", defaultValueToMatch, "three dogs", true, false},
		{"returns false when context does not exist", defaultValueToMatch, nil, false, false},
		{"returns false when valueToMatch is not a string slice", testutils.CreateConfigValueAndAssertOk(suite.T(), "example.com"), "doesn't matter", true, false},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()

			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: testCase.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestPropDoesNotContainCriteriaEvaluation() {
	operator := prefabProto.Criterion_PROP_DOES_NOT_CONTAIN_ONE_OF
	contextPropertyName := "user.testProperty"
	defaultValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"one", "two"})

	tests := []struct {
		name               string
		valueToMatch       *prefabProto.ConfigValue
		contextValue       interface{}
		contextValueExists bool
		expected           bool
	}{
		{"returns fa for testProperty does not contain with one", defaultValueToMatch, "one tree", true, false},
		{"returns false for testProperty does not contain three", defaultValueToMatch, "three dogs", true, true},
		{"returns true when context does not exist", defaultValueToMatch, nil, false, true},
		{"returns true when valueToMatch is not a string slice", testutils.CreateConfigValueAndAssertOk(suite.T(), "example.com"), "doesn't matter", true, true},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()

			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: testCase.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestNumericOps() {
	contextPropertyName := "user.orderCount"
	tests := []struct {
		name               string
		contextValue       any
		operator           prefabProto.Criterion_CriterionOperator
		valueToMatch       any
		contextValueExists bool
		matchExpected      bool
	}{
		{"12 is less than 14", 12, prefabProto.Criterion_PROP_LESS_THAN, 14, true, true},
		{"12.1 is less than 14", 12.1, prefabProto.Criterion_PROP_LESS_THAN, 14, true, true},
		{"12.1 is less than 14.2", 12.1, prefabProto.Criterion_PROP_LESS_THAN, 14.2, true, true},
		{"12 is not less than 12", 12, prefabProto.Criterion_PROP_LESS_THAN, 12, true, false},
		{"14.2 is not less than 12", 14, prefabProto.Criterion_PROP_LESS_THAN, 12, true, false},
		{"12 is less than or equal 14", 12, prefabProto.Criterion_PROP_LESS_THAN_OR_EQUAL, 14, true, true},
		{"12.1 is less than or equal 14", 12.1, prefabProto.Criterion_PROP_LESS_THAN_OR_EQUAL, 14, true, true},
		{"12.1 is less than or equal 14.2", 12.1, prefabProto.Criterion_PROP_LESS_THAN_OR_EQUAL, 14.2, true, true},
		{"12 is less than or equal 12", 12, prefabProto.Criterion_PROP_LESS_THAN_OR_EQUAL, 12, true, true},
		{"14.2 is not less than or equal 12", 14, prefabProto.Criterion_PROP_LESS_THAN_OR_EQUAL, 12, true, false},
		{"14 is greater than 12", 14, prefabProto.Criterion_PROP_GREATER_THAN, 12, true, true},
		{"14.1 is greater than 12", 14.1, prefabProto.Criterion_PROP_GREATER_THAN, 12, true, true},
		{"14.1 is greater than 12.2", 14.1, prefabProto.Criterion_PROP_GREATER_THAN, 12.2, true, true},
		{"14 is not greater than 14", 14, prefabProto.Criterion_PROP_GREATER_THAN, 14, true, false},
		{"13 is not greater than 14", 13, prefabProto.Criterion_PROP_GREATER_THAN, 14, true, false},
		{"14 is greater than or equal 12", 14, prefabProto.Criterion_PROP_GREATER_THAN_OR_EQUAL, 12, true, true},
		{"14.1 is greater than or equal 12", 14.1, prefabProto.Criterion_PROP_GREATER_THAN_OR_EQUAL, 12, true, true},
		{"14.1 is greater than or equal 12.2", 14.1, prefabProto.Criterion_PROP_GREATER_THAN_OR_EQUAL, 12.2, true, true},
		{"14 is greater than or equal 14", 14, prefabProto.Criterion_PROP_GREATER_THAN_OR_EQUAL, 14, true, true},
		{"13 is not greater than or equal 14", 13, prefabProto.Criterion_PROP_GREATER_THAN_OR_EQUAL, 14, true, false},
		{"return false for missing context value", nil, prefabProto.Criterion_PROP_GREATER_THAN_OR_EQUAL, 14, false, false},
		{"return false for string context value", "12", prefabProto.Criterion_PROP_GREATER_THAN_OR_EQUAL, 14, true, false},
		{"return false for string match value", 14, prefabProto.Criterion_PROP_GREATER_THAN_OR_EQUAL, "12", true, false},
	}
	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()
			criterion := &prefabProto.Criterion{Operator: testCase.operator, ValueToMatch: testutils.CreateConfigValueAndAssertOk(suite.T(), testCase.valueToMatch), PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.matchExpected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestDateOps() {
	decFirst24Str := "2024-12-01T00:00:00Z"
	decFirst24Time, err := time.Parse(time.RFC3339, decFirst24Str)
	suite.Assert().Nil(err)
	decFirst24Millis := decFirst24Time.UnixMilli()
	janFirst25Str := "2025-01-01T00:00:00Z"
	janFirst25Time, err := time.Parse(time.RFC3339, janFirst25Str)
	suite.Assert().Nil(err)
	janFirst25Millis := janFirst25Time.UnixMilli()

	contextPropertyName := "user.orderCount"
	tests := []struct {
		name               string
		contextValue       any
		operator           prefabProto.Criterion_CriterionOperator
		valueToMatch       any
		contextValueExists bool
		matchExpected      bool
	}{
		{"dec 24 before jan 25", decFirst24Str, prefabProto.Criterion_PROP_BEFORE, janFirst25Str, true, true},
		{"dec 24 millis before jan 25", decFirst24Millis, prefabProto.Criterion_PROP_BEFORE, janFirst25Str, true, true},
		{"dec 24 millis before jan 25 millis", decFirst24Millis, prefabProto.Criterion_PROP_BEFORE, janFirst25Millis, true, true},
		{"dec 24 before jan 25 millis", decFirst24Millis, prefabProto.Criterion_PROP_BEFORE, janFirst25Millis, true, true},
		{"dec 24 not before dec 24", decFirst24Str, prefabProto.Criterion_PROP_BEFORE, decFirst24Str, true, false},
		{"jan 25 not before dec 24", janFirst25Str, prefabProto.Criterion_PROP_BEFORE, decFirst24Str, true, false},
		{"no context date not before dec 24", nil, prefabProto.Criterion_PROP_BEFORE, janFirst25Str, false, false},
		{"bad context date returns false", "not a date", prefabProto.Criterion_PROP_BEFORE, janFirst25Str, true, false},
		{"bad match date returns false", decFirst24Time, prefabProto.Criterion_PROP_BEFORE, "not a date", true, false},
		{"jan 25 after dec 24", janFirst25Str, prefabProto.Criterion_PROP_AFTER, decFirst24Str, true, true},
		{"jan 25 not after jan 25", janFirst25Str, prefabProto.Criterion_PROP_AFTER, janFirst25Str, true, false},
		{"dec 24 not after jan 25", decFirst24Str, prefabProto.Criterion_PROP_AFTER, janFirst25Str, true, false},
	}
	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()
			criterion := &prefabProto.Criterion{Operator: testCase.operator, ValueToMatch: testutils.CreateConfigValueAndAssertOk(suite.T(), testCase.valueToMatch), PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.matchExpected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestSemverOps() {
	contextPropertyName := "app.version"
	tests := []struct {
		name               string
		contextValue       any
		operator           prefabProto.Criterion_CriterionOperator
		valueToMatch       any
		contextValueExists bool
		matchExpected      bool
	}{
		{"2.0.0 equal to 2.0.0", "2.0.0", prefabProto.Criterion_PROP_SEMVER_EQUAL, "2.0.0", true, true},
		{"2.0.0 not greater than 2.0.0", "2.0.0", prefabProto.Criterion_PROP_SEMVER_GREATER_THAN, "2.0.0", true, false},
		{"2.0.0 not less than 2.0.0", "2.0.0", prefabProto.Criterion_PROP_SEMVER_LESS_THAN, "2.0.0", true, false},
		{"2.0.1 is greater than 2.0.0", "2.0.1", prefabProto.Criterion_PROP_SEMVER_GREATER_THAN, "2.0.0", true, true},
		{"2.0.0 is less than 2.0.1", "2.0.0", prefabProto.Criterion_PROP_SEMVER_LESS_THAN, "2.0.1", true, true},
		{"missing context value isn't equal", "", prefabProto.Criterion_PROP_SEMVER_EQUAL, "2.0.1", false, false},
		{"v2.0.0 is not equal to 2.0.0 because v prefix is invalid", "v2.0.0", prefabProto.Criterion_PROP_SEMVER_EQUAL, "2.0.1", true, false},
		{"bad value in match rule returns false", "v2.0.0", prefabProto.Criterion_PROP_SEMVER_EQUAL, "foobar", true, false},
	}
	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()
			criterion := &prefabProto.Criterion{Operator: testCase.operator, ValueToMatch: testutils.CreateConfigValueAndAssertOk(suite.T(), testCase.valueToMatch), PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.matchExpected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestRegexOps() {
	contextPropertyName := "user.code"
	tests := []struct {
		name               string
		contextValue       any
		operator           prefabProto.Criterion_CriterionOperator
		valueToMatch       any
		contextValueExists bool
		matchExpected      bool
	}{
		{"ab matches a+b+", "ab", prefabProto.Criterion_PROP_MATCHES, "a+b+", true, true},
		{"ab does not match a+b+ is false", "ab", prefabProto.Criterion_PROP_DOES_NOT_MATCH, "a+b+", true, false},
		{"ab matches a+b+c is false", "ab", prefabProto.Criterion_PROP_MATCHES, "a+b+c", true, false},
		{"ab does not match a+b+c is true", "ab", prefabProto.Criterion_PROP_DOES_NOT_MATCH, "a+b+c", true, true},
		{"bad regex returns false", "ab", prefabProto.Criterion_PROP_MATCHES, "[a+b+", true, false},
		{"missing context returns false", "", prefabProto.Criterion_PROP_MATCHES, "[a+b+", false, false},
	}
	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()
			criterion := &prefabProto.Criterion{Operator: testCase.operator, ValueToMatch: testutils.CreateConfigValueAndAssertOk(suite.T(), testCase.valueToMatch), PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.matchExpected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestPropIsOneOf() {
	operator := prefabProto.Criterion_PROP_IS_ONE_OF
	contextPropertyName := "user.email.domain"
	defaultValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"yahoo.com", "example.com"})
	alternateValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"1", "2", "3"})

	tests := []struct {
		name               string
		valueToMatch       *prefabProto.ConfigValue
		contextValue       interface{}
		contextValueExists bool
		expected           bool
	}{
		{"returns true when in set", defaultValueToMatch, "yahoo.com", true, true},
		{"returns false when not in set", defaultValueToMatch, "gmail.com", true, false},
		{"returns false when context value is not a string", defaultValueToMatch, 1, true, false},
		{"returns false when context value does not exist", defaultValueToMatch, nil, false, false},
		{"returns false when valueToMatch is not a string slice", testutils.CreateConfigValueAndAssertOk(suite.T(), "example.com"), "doesn't matter", true, false},
		{"returns true when the string matches the slice", alternateValueToMatch, "2", true, true},
		{"returns true when the non-string matches the slice (it is coerced)", alternateValueToMatch, 2, true, true},
		{"returns true when the context array overlaps the set", defaultValueToMatch, []string{"yahoo.com", "hats"}, true, true},
		{"returns false when the context array does not overlap the set", defaultValueToMatch, []string{"pumpkins", "hats", "shoes"}, true, false},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()

			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: testCase.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestPropIsNotOneOf() {
	operator := prefabProto.Criterion_PROP_IS_NOT_ONE_OF
	contextPropertyName := "user.email.domain"
	defaultValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"yahoo.com", "example.com"})
	alternateValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), []string{"1", "2", "3"})

	tests := []struct {
		name               string
		valueToMatch       *prefabProto.ConfigValue
		contextValue       interface{}
		contextValueExists bool
		expected           bool
	}{
		{"returns false when in set", defaultValueToMatch, "yahoo.com", true, false},
		{"returns true when not in set", defaultValueToMatch, "gmail.com", true, true},
		{"returns true when context value is not a string", defaultValueToMatch, 1, true, true},
		{"returns true when context value does not exist", defaultValueToMatch, nil, false, true},
		{"returns true when valueToMatch is not a string slice", testutils.CreateConfigValueAndAssertOk(suite.T(), "example.com"), "doesn't matter", true, true},
		{"returns false when the string matches the slice", alternateValueToMatch, "2", true, false},
		{"returns false when the non-string matches the slice (it is coerced)", alternateValueToMatch, 2, true, false},
		{"returns false when the context array overlaps the set", defaultValueToMatch, []string{"yahoo.com", "hats"}, true, false},
		{"returns true when the context array does not overlap the set", defaultValueToMatch, []string{"pumpkins", "hats", "shoes"}, true, true},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()

			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: testCase.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestHierarchicalMatch() {
	operator := prefabProto.Criterion_HIERARCHICAL_MATCH
	contextPropertyName := "team.path"
	defaultValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), "a.b.c")

	tests := []struct {
		name               string
		valueToMatch       *prefabProto.ConfigValue
		contextValue       interface{}
		contextValueExists bool
		expected           bool
	}{
		{"returns true when matched", defaultValueToMatch, "a.b.c.d", true, true},
		{"returns false when not matched", defaultValueToMatch, "a.b", true, false},
		{"returns false when context value is not a string", defaultValueToMatch, 1, true, false},
		{"returns false when context value does not exist", defaultValueToMatch, nil, false, false},
		{"returns false when valueToMatch is not a string", testutils.CreateConfigValueAndAssertOk(suite.T(), 100), "a.b.c.d", true, false},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()

			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: testCase.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestInIntRangeCriterion() {
	operator := prefabProto.Criterion_IN_INT_RANGE
	contextPropertyName := teamSize
	defaultValueToMatch := &prefabProto.ConfigValue{Type: &prefabProto.ConfigValue_IntRange{IntRange: &prefabProto.IntRange{Start: internal.Int64Ptr(0), End: internal.Int64Ptr(100)}}}

	tests := []struct {
		name               string
		valueToMatch       *prefabProto.ConfigValue
		contextValue       interface{}
		contextValueExists bool
		expected           bool
	}{
		{"returns true when in range", defaultValueToMatch, 10, true, true},
		{"returns true when in range -- equal to range start", defaultValueToMatch, 0, true, true},
		{"returns false when not range -- equal to range end", defaultValueToMatch, 100, true, false},
		{"returns false when not range -- above range end", defaultValueToMatch, 101, true, false},
		{"returns false when not range -- below range start", defaultValueToMatch, -10, true, false},
		{"returns true when in range -- float", defaultValueToMatch, 10.4, true, true},
		{"returns true when in range treating empty start as long-min", &prefabProto.ConfigValue{Type: &prefabProto.ConfigValue_IntRange{IntRange: &prefabProto.IntRange{End: internal.Int64Ptr(100)}}}, 10, true, true},
		{"returns true when in range treating empty end as long-min", &prefabProto.ConfigValue{Type: &prefabProto.ConfigValue_IntRange{IntRange: &prefabProto.IntRange{Start: internal.Int64Ptr(0)}}}, 10, true, true},
		{"returns true when in range treating empty start and end as all numbers < long-min", &prefabProto.ConfigValue{Type: &prefabProto.ConfigValue_IntRange{IntRange: &prefabProto.IntRange{}}}, 10, true, true},
		{"returns false when context value is a string", defaultValueToMatch, "foo", true, false},
		{"returns false when context value does not exist", defaultValueToMatch, nil, false, false},
		{"returns false when value to match isn't an int range", testutils.CreateConfigValueAndAssertOk(suite.T(), "what's up doc"), nil, false, false},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			defer assertMockCalled()

			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: testCase.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestInSegmentCriterion() {
	// this test is based on in-int range as the target segment
	contextPropertyName := teamSize
	segmentTargetIntRangeConfig := suite.createInIntRangeSegmentTarget(prefabProto.Criterion_IN_INT_RANGE, contextPropertyName, internal.Int64Ptr(0), internal.Int64Ptr(100))
	targetConfigKey := segmentTargetIntRangeConfig.GetKey()
	defaultValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), targetConfigKey)

	tests := []struct {
		name                string
		valueToMatch        *prefabProto.ConfigValue
		contextValue        interface{}
		contextValueExists  bool
		targetSegmentConfig *prefabProto.Config
		targetSegmentExists bool
		expected            bool
	}{
		{"returns true in segment", defaultValueToMatch, 10, true, segmentTargetIntRangeConfig, true, true},
		{"returns false not in segment", defaultValueToMatch, 1000, true, segmentTargetIntRangeConfig, true, false},
		{"returns false if target segment returns does not exist", defaultValueToMatch, 1000, true, nil, false, false},
		{"returns false if target segment returns non boolean value", defaultValueToMatch, 1000, true, suite.nonBoolReturnSimpleConfig, true, false},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			mockContext.On("GetContextValue", "").Return(nil, false)

			defer assertMockCalled()

			assertMockConfigStoreGetterCalled := suite.setupMockConfigStoreGetter(targetConfigKey, testCase.targetSegmentConfig, testCase.targetSegmentExists)
			defer assertMockConfigStoreGetterCalled()

			criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_IN_SEG, ValueToMatch: testCase.valueToMatch}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestNotInSegmentCriterion() {
	// this test is based on in-int range as the target segment
	contextPropertyName := teamSize
	segmentTargetIntRangeConfig := suite.createInIntRangeSegmentTarget(prefabProto.Criterion_IN_INT_RANGE, contextPropertyName, internal.Int64Ptr(0), internal.Int64Ptr(100))
	targetConfigKey := segmentTargetIntRangeConfig.GetKey()
	defaultValueToMatch := testutils.CreateConfigValueAndAssertOk(suite.T(), targetConfigKey)

	tests := []struct {
		name                string
		valueToMatch        *prefabProto.ConfigValue
		contextValue        interface{}
		contextValueExists  bool
		targetSegmentConfig *prefabProto.Config
		targetSegmentExists bool
		expected            bool
	}{
		{"returns false in segment", defaultValueToMatch, 10, true, segmentTargetIntRangeConfig, true, false},
		{"returns true not in segment", defaultValueToMatch, 1000, true, segmentTargetIntRangeConfig, true, true},
		{"returns true if target segment returns does not exist", defaultValueToMatch, 1000, true, nil, false, true},
		{"returns true if target segment returns non boolean value", defaultValueToMatch, 1000, true, suite.nonBoolReturnSimpleConfig, true, true},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, testCase.contextValue, testCase.contextValueExists)
			mockContext.On("GetContextValue", "").Return(nil, false)

			defer assertMockCalled()

			assertMockConfigStoreGetterCalled := suite.setupMockConfigStoreGetter(targetConfigKey, testCase.targetSegmentConfig, testCase.targetSegmentExists)
			defer assertMockConfigStoreGetterCalled()

			criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_NOT_IN_SEG, ValueToMatch: testCase.valueToMatch}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(testCase.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) createInIntRangeSegmentTarget(operator prefabProto.Criterion_CriterionOperator, contextPropertyName string, start *int64, end *int64) *prefabProto.Config {
	targetConfigKey := "teamSizeSegment"
	inIntRangeCriterion := &prefabProto.Criterion{
		Operator: operator,
		ValueToMatch: &prefabProto.ConfigValue{
			Type: &prefabProto.ConfigValue_IntRange{
				IntRange: &prefabProto.IntRange{
					Start: start,
					End:   end,
				},
			},
		},
		PropertyName: contextPropertyName,
	}
	targetSegmentConfig := &prefabProto.Config{
		Id:              1,
		ProjectId:       1,
		Key:             targetConfigKey,
		ConfigType:      prefabProto.ConfigType_CONFIG,
		ValueType:       prefabProto.Config_BOOL,
		SendToClientSdk: false,
		Rows: []*prefabProto.ConfigRow{
			{
				ProjectEnvId: internal.Int64Ptr(suite.projectEnvID),
				Values: []*prefabProto.ConditionalValue{
					{
						Criteria: []*prefabProto.Criterion{inIntRangeCriterion},
						Value:    testutils.CreateConfigValueAndAssertOk(suite.T(), true),
					},
					{
						Value: testutils.CreateConfigValueAndAssertOk(suite.T(), false),
					},
				},
			},
		},
	}

	return targetSegmentConfig
}

func (suite *ConfigRuleTestSuite) setupMockContext(contextPropertyName string, contextValue interface{}, contextExistsValue bool) (*MockContextGetter, func()) {
	mockContext := MockContextGetter{}
	if contextExistsValue {
		// When context exists, return the provided contextValue and true.
		mockContext.On("GetContextValue", contextPropertyName).Return(contextValue, contextExistsValue)
	} else {
		// When context doesn't exist, ensure the mock returns nil for the contextValue (or a suitable zero value) and false.
		// This simulates the absence of the value correctly.
		mockContext.On("GetContextValue", contextPropertyName).Return(nil, false)
	}

	cleanupFunc := func() {
		mockContext.AssertExpectations(suite.T())
	}

	return &mockContext, cleanupFunc
}

type ContextMocking struct {
	value               interface{}
	contextPropertyName string
	exists              bool
}

func (suite *ConfigRuleTestSuite) setupMockContextWithMultipleValues(mockings []ContextMocking) (*MockContextGetter, func()) {
	mockContext := &MockContextGetter{}

	for _, mocking := range mockings {
		contextValue := mocking.value
		contextExistsValue := mocking.exists

		if contextExistsValue {
			mockContext.On("GetContextValue", mocking.contextPropertyName).Return(contextValue, true)
		} else {
			mockContext.On("GetContextValue", mocking.contextPropertyName).Return(nil, false)
		}
	}

	cleanupFunc := func() {
		mockContext.AssertExpectations(suite.T())
	}

	return mockContext, cleanupFunc
}

func (suite *ConfigRuleTestSuite) setupMockConfigStoreGetter(configKey string, config *prefabProto.Config, configExists bool) func() {
	suite.mockConfigStoreGetter.On("GetConfig", configKey).Return(config, configExists)

	cleanupFunc := func() {
		suite.mockConfigStoreGetter.AssertExpectations(suite.T())
	}

	return cleanupFunc
}

func (suite *ConfigRuleTestSuite) TestCurrentTimeProperty() {
	// Create a criterion with the "prefab.current-time" property
	criterion := &prefabProto.Criterion{
		PropertyName: "prefab.current-time",
		Operator:     prefabProto.Criterion_PROP_GREATER_THAN,
		ValueToMatch: &prefabProto.ConfigValue{
			Type: &prefabProto.ConfigValue_Int{
				Int: time.Now().UTC().UnixMilli() - 1000, // 1 second ago
			},
		},
	}

	// Create a mock context getter
	mockContextGetter := &MockContextGetter{}

	// Set up the mock to return nil, false for any property name
	// This is because our implementation will handle the "prefab.current-time" property
	// regardless of what the context getter returns
	mockContextGetter.On("GetContextValue", mock.Anything).Return(nil, false)

	// Evaluate the criterion
	result := suite.evaluator.EvaluateCriterion(criterion, mockContextGetter)

	// The result should be true because the current time is greater than 1 second ago
	suite.True(result)

	// Test the opposite case - current time should be less than a future time
	futureCriterion := &prefabProto.Criterion{
		PropertyName: "prefab.current-time",
		Operator:     prefabProto.Criterion_PROP_LESS_THAN,
		ValueToMatch: &prefabProto.ConfigValue{
			Type: &prefabProto.ConfigValue_Int{
				Int: time.Now().UTC().UnixMilli() + 1000, // 1 second in the future
			},
		},
	}

	// Evaluate the criterion
	result = suite.evaluator.EvaluateCriterion(futureCriterion, mockContextGetter)

	// The result should be true because the current time is less than 1 second in the future
	suite.True(result)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestConfigRuleTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigRuleTestSuite))
}
