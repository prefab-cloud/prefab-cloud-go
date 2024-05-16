package internal_test

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/prefab-cloud/prefab-cloud-go/mocks"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"

	"github.com/stretchr/testify/suite"
)

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
		ValueToMatch: createConfigValueAndAssertOk([]string{"ing"}, suite.T()),
		PropertyName: "department.name",
	}

	departmentNameIsComplianceOrSecurity := &prefabProto.Criterion{
		Operator:     prefabProto.Criterion_PROP_IS_ONE_OF,
		ValueToMatch: createConfigValueAndAssertOk([]string{"compliance", "security"}, suite.T()),
		PropertyName: "department.name",
	}

	departmentNameIsAliens := &prefabProto.Criterion{
		Operator:     prefabProto.Criterion_PROP_IS_ONE_OF,
		ValueToMatch: createConfigValueAndAssertOk([]string{"aliens"}, suite.T()),
		PropertyName: "department.name",
	}

	securityClearance := &prefabProto.Criterion{
		Operator:     prefabProto.Criterion_PROP_IS_ONE_OF,
		ValueToMatch: createConfigValueAndAssertOk([]string{"top secret", "top top secret"}, suite.T()),
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
				ProjectEnvId: Int64Ptr(matchingProjectEnvID),
				Values: []*prefabProto.ConditionalValue{
					{
						Criteria: []*prefabProto.Criterion{departmentNameEndsWithIngCriterion, securityClearance},
						Value:    createConfigValueAndAssertOk(1, suite.T()),
					},
					{
						Criteria: []*prefabProto.Criterion{departmentNameEndsWithIngCriterion},
						Value:    createConfigValueAndAssertOk(2, suite.T()),
					},
					{
						Criteria: []*prefabProto.Criterion{departmentNameIsComplianceOrSecurity},
						Value:    createConfigValueAndAssertOk(3, suite.T()),
					},
				},
			},
			{
				Values: []*prefabProto.ConditionalValue{
					{
						Criteria: []*prefabProto.Criterion{departmentNameIsAliens},
						Value:    createConfigValueAndAssertOk(10, suite.T()),
					},
					{
						Value: createConfigValueAndAssertOk(11, suite.T()),
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
		{"returns 1 for high security clearance mining department with matching projectEnv", matchingProjectEnvID, createConfigValueAndAssertOk(1, suite.T()), 0, 0, []ContextMocking{{contextPropertyName: "department.name", value: "mining", exists: true}, {contextPropertyName: "security.clearance", value: "top secret", exists: true}}},
		{"returns 2 for no security clearance mining department with matching projectEnv", matchingProjectEnvID, createConfigValueAndAssertOk(2, suite.T()), 0, 1, []ContextMocking{{contextPropertyName: "department.name", value: "mining", exists: true}, {contextPropertyName: "security.clearance", value: nil, exists: false}}},
		{"returns 3 for security department with matching projectEnv", matchingProjectEnvID, createConfigValueAndAssertOk(3, suite.T()), 0, 2, []ContextMocking{{contextPropertyName: "department.name", value: "security", exists: true}}},
		{"returns 10 for aliens department with matching projectEnv", matchingProjectEnvID, createConfigValueAndAssertOk(10, suite.T()), 1, 0, []ContextMocking{{contextPropertyName: "department.name", value: "aliens", exists: true}}},
		{"returns 11 for cleanup department with matching projectEnv", matchingProjectEnvID, createConfigValueAndAssertOk(11, suite.T()), 1, 1, []ContextMocking{{contextPropertyName: "department.name", value: "cleanup", exists: true}}},
		{"returns 10 for aliens department with mismatching projectEnv", mismatchingProjectEnvID, createConfigValueAndAssertOk(10, suite.T()), 0, 0, []ContextMocking{{contextPropertyName: "department.name", value: "aliens", exists: true}}},
		{"returns 11 for cleanup department with mismatching projectEnv", mismatchingProjectEnvID, createConfigValueAndAssertOk(11, suite.T()), 0, 1, []ContextMocking{{contextPropertyName: "department.name", value: "cleanup", exists: true}}},
		{"returns 11 for high security clearance mining department with mismatching projectEnv", mismatchingProjectEnvID, createConfigValueAndAssertOk(11, suite.T()), 0, 1, []ContextMocking{{contextPropertyName: "department.name", value: "mining", exists: true}, {contextPropertyName: "security.clearance", value: "top secret", exists: true}}},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			suite.evaluator = internal.NewConfigRuleEvaluator(&suite.mockConfigStoreGetter, mocks.NewMockProjectEnvIDSupplier(testCase.projectEnvID))
			mockContext, _ := suite.setupMockContextWithMultipleValues(testCase.contextMockings)
			conditionMatch := suite.evaluator.EvaluateConfig(config, mockContext)
			suite.Equal(testCase.expectedValue, conditionMatch.Match, "value should match")
			suite.Equal(testCase.expectedRowIndex, conditionMatch.RowIndex, "rowIndex should match")
			suite.Equal(testCase.expectedConditionalValueIndex, conditionMatch.ConditionalValueIndex, "conditionalValueIndex should match")
		})
	}
}

func (suite *ConfigRuleTestSuite) TestAlwaysTrueCriteria() {
	suite.Run("returns true", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_ALWAYS_TRUE}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, internal.NewContextSet())
		suite.True(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestNotSetCriteria() {
	suite.Run("returns false", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_NOT_SET}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, internal.NewContextSet())
		suite.False(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestPropEndsWithCriteriaEvaluation() {
	operator := prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF
	contextPropertyName := "user.email"
	defaultValueToMatch := createConfigValueAndAssertOk([]string{"example.com", "splat"}, suite.T())

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
		{"returns false when valueToMatch is not a string slice", createConfigValueAndAssertOk("example.com", suite.T()), "doesn't matter", true, false},
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
	defaultValueToMatch := createConfigValueAndAssertOk([]string{"example.com", "splat"}, suite.T())

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
		{"returns true when valueToMatch is not a string slice", createConfigValueAndAssertOk("example.com", suite.T()), "doesn't matter", true, true},
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

func (suite *ConfigRuleTestSuite) TestPropIsOneOf() {
	operator := prefabProto.Criterion_PROP_IS_ONE_OF
	contextPropertyName := "user.email.domain"
	defaultValueToMatch := createConfigValueAndAssertOk([]string{"yahoo.com", "example.com"}, suite.T())

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
		{"returns false when valueToMatch is not a string slice", createConfigValueAndAssertOk("example.com", suite.T()), "doesn't matter", true, false},
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
	defaultValueToMatch := createConfigValueAndAssertOk([]string{"yahoo.com", "example.com"}, suite.T())

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
		{"returns true when valueToMatch is not a string slice", createConfigValueAndAssertOk("example.com", suite.T()), "doesn't matter", true, true},
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
	defaultValueToMatch := createConfigValueAndAssertOk("a.b.c", suite.T())

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
		{"returns false when valueToMatch is not a string", createConfigValueAndAssertOk(100, suite.T()), "a.b.c.d", true, false},
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
	contextPropertyName := "team.size"
	defaultValueToMatch := &prefabProto.ConfigValue{Type: &prefabProto.ConfigValue_IntRange{IntRange: &prefabProto.IntRange{Start: Int64Ptr(0), End: Int64Ptr(100)}}}

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
		{"returns true when in range treating empty start as long-min", &prefabProto.ConfigValue{Type: &prefabProto.ConfigValue_IntRange{IntRange: &prefabProto.IntRange{End: Int64Ptr(100)}}}, 10, true, true},
		{"returns true when in range treating empty end as long-min", &prefabProto.ConfigValue{Type: &prefabProto.ConfigValue_IntRange{IntRange: &prefabProto.IntRange{Start: Int64Ptr(0)}}}, 10, true, true},
		{"returns true when in range treating empty start and end as all numbers < long-min", &prefabProto.ConfigValue{Type: &prefabProto.ConfigValue_IntRange{IntRange: &prefabProto.IntRange{}}}, 10, true, true},
		{"returns false when context value is a string", defaultValueToMatch, "foo", true, false},
		{"returns false when context value does not exist", defaultValueToMatch, nil, false, false},
		{"returns false when value to match isn't an int range", createConfigValueAndAssertOk("what's up doc", suite.T()), nil, false, false},
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
	contextPropertyName := "team.size"
	segmentTargetIntRangeConfig := suite.createInIntRangeSegmentTarget(prefabProto.Criterion_IN_INT_RANGE, contextPropertyName, Int64Ptr(0), Int64Ptr(100))
	targetConfigKey := segmentTargetIntRangeConfig.GetKey()
	defaultValueToMatch := createConfigValueAndAssertOk(targetConfigKey, suite.T())

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
	contextPropertyName := "team.size"
	segmentTargetIntRangeConfig := suite.createInIntRangeSegmentTarget(prefabProto.Criterion_IN_INT_RANGE, contextPropertyName, Int64Ptr(0), Int64Ptr(100))
	targetConfigKey := segmentTargetIntRangeConfig.GetKey()
	defaultValueToMatch := createConfigValueAndAssertOk(targetConfigKey, suite.T())

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
				ProjectEnvId: Int64Ptr(suite.projectEnvID),
				Values: []*prefabProto.ConditionalValue{
					{
						Criteria: []*prefabProto.Criterion{inIntRangeCriterion},
						Value:    createConfigValueAndAssertOk(true, suite.T()),
					},
					{
						Value: createConfigValueAndAssertOk(false, suite.T()),
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestConfigRuleTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigRuleTestSuite))
}