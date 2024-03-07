package prefab

import (
	"github.com/prefab-cloud/prefab-cloud-go/internal"
	"github.com/prefab-cloud/prefab-cloud-go/mocks"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MockContextGetter struct {
	mock.Mock
}

func (m *MockContextGetter) GetValue(propertyName string) (value interface{}, valueExists bool) {
	args := m.Called(propertyName)
	if args.Get(0) != nil {
		value = args.Get(0).(interface{})
	} else {
		value = nil
	}
	valueExists = args.Bool(1)
	return value, valueExists
}

type ConfigRuleTestSuite struct {
	suite.Suite
	evaluator                 *ConfigRuleEvaluator
	projectEnvId              int64
	mockConfigStoreGetter     mocks.MockConfigStoreGetter
	nonBoolReturnSimpleConfig *prefabProto.Config
}

func (suite *ConfigRuleTestSuite) SetupTest() {
	suite.mockConfigStoreGetter = mocks.MockConfigStoreGetter{}
	suite.projectEnvId = 101
	suite.evaluator = NewConfigRuleEvaluator(&suite.mockConfigStoreGetter, suite.projectEnvId)
}

func (suite *ConfigRuleTestSuite) TestFullRuleEvaluation() {
	matchingProjectEnvId := int64(101)
	mismatchingProjectEnvId := int64(102)
	departmentNameEndsWithIngCriterion := &prefabProto.Criterion{
		Operator:     prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF,
		ValueToMatch: internal.CreateConfigValue([]string{"ing"}),
		PropertyName: "department.name",
	}

	departmentNameIsComplianceOrSecurity := &prefabProto.Criterion{
		Operator:     prefabProto.Criterion_PROP_IS_ONE_OF,
		ValueToMatch: internal.CreateConfigValue([]string{"compliance", "security"}),
		PropertyName: "department.name",
	}

	departmentNameIsAliens := &prefabProto.Criterion{
		Operator:     prefabProto.Criterion_PROP_IS_ONE_OF,
		ValueToMatch: internal.CreateConfigValue([]string{"aliens"}),
		PropertyName: "department.name",
	}

	securityClearance := &prefabProto.Criterion{
		Operator:     prefabProto.Criterion_PROP_IS_ONE_OF,
		ValueToMatch: internal.CreateConfigValue([]string{"top secret", "top top secret"}),
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
				ProjectEnvId: Int64Ptr(matchingProjectEnvId),
				Values: []*prefabProto.ConditionalValue{
					{
						Criteria: []*prefabProto.Criterion{departmentNameEndsWithIngCriterion, securityClearance},
						Value:    internal.CreateConfigValue(1),
					},
					{
						Criteria: []*prefabProto.Criterion{departmentNameEndsWithIngCriterion},
						Value:    internal.CreateConfigValue(2),
					},
					{
						Criteria: []*prefabProto.Criterion{departmentNameIsComplianceOrSecurity},
						Value:    internal.CreateConfigValue(3),
					},
				},
			},
			{
				Values: []*prefabProto.ConditionalValue{
					{
						Criteria: []*prefabProto.Criterion{departmentNameIsAliens},
						Value:    internal.CreateConfigValue(10),
					},
					{
						Value: internal.CreateConfigValue(11),
					},
				},
			},
		},
	}

	tests := []struct {
		name                          string
		projectEnvId                  int64
		expectedValue                 *prefabProto.ConfigValue
		expectedRowIndex              int
		expectedConditionalValueIndex int
		contextMockings               []ContextMocking
	}{
		{"returns 1 for high security clearance mining department with matching projectEnv", matchingProjectEnvId, internal.CreateConfigValue(1), 0, 0, []ContextMocking{{contextPropertyName: "department.name", value: "mining", exists: true}, {contextPropertyName: "security.clearance", value: "top secret", exists: true}}},
		{"returns 2 for no security clearance mining department with matching projectEnv", matchingProjectEnvId, internal.CreateConfigValue(2), 0, 1, []ContextMocking{{contextPropertyName: "department.name", value: "mining", exists: true}, {contextPropertyName: "security.clearance", value: nil, exists: false}}},
		{"returns 3 for security department with matching projectEnv", matchingProjectEnvId, internal.CreateConfigValue(3), 0, 2, []ContextMocking{{contextPropertyName: "department.name", value: "security", exists: true}}},
		{"returns 10 for aliens department with matching projectEnv", matchingProjectEnvId, internal.CreateConfigValue(10), 1, 0, []ContextMocking{{contextPropertyName: "department.name", value: "aliens", exists: true}}},
		{"returns 11 for cleanup department with matching projectEnv", matchingProjectEnvId, internal.CreateConfigValue(11), 1, 1, []ContextMocking{{contextPropertyName: "department.name", value: "cleanup", exists: true}}},
		{"returns 10 for aliens department with mismatching projectEnv", mismatchingProjectEnvId, internal.CreateConfigValue(10), 1, 0, []ContextMocking{{contextPropertyName: "department.name", value: "aliens", exists: true}}},
		{"returns 11 for cleanup department with mismatching projectEnv", mismatchingProjectEnvId, internal.CreateConfigValue(11), 1, 1, []ContextMocking{{contextPropertyName: "department.name", value: "cleanup", exists: true}}},
		{"returns 11 for high security clearance mining department with matching projectEnv", mismatchingProjectEnvId, internal.CreateConfigValue(1), 0, 0, []ContextMocking{{contextPropertyName: "department.name", value: "mining", exists: true}, {contextPropertyName: "security.clearance", value: "top secret", exists: true}}},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.evaluator = NewConfigRuleEvaluator(&suite.mockConfigStoreGetter, suite.projectEnvId)
			mockContext, _ := suite.setupMockContextWithMultipleValues(tt.contextMockings)
			conditionMatch := suite.evaluator.EvaluateConfig(config, mockContext)
			suite.Equal(tt.expectedValue, conditionMatch.match, "value should match")
			suite.Equal(tt.expectedRowIndex, conditionMatch.rowIndex, "rowIndex should match")
			suite.Equal(tt.expectedConditionalValueIndex, conditionMatch.conditionalValueIndex, "conditionalValueIndex should match")
		})
	}
}

func (suite *ConfigRuleTestSuite) TestAlwaysTrueCriteria() {
	suite.Run("returns true", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_ALWAYS_TRUE}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, NewContextSet())
		suite.Assert().True(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestNotSetCriteria() {
	suite.Run("returns false", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_NOT_SET}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, NewContextSet())
		suite.Assert().False(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestPropEndsWithCriteriaEvaluation() {
	operator := prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF
	contextPropertyName := "user.email"
	defaultValueToMatch := internal.CreateConfigValue([]string{"example.com", "splat"})

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
		{"returns false when valueToMatch is not a string slice", internal.CreateConfigValue("example.com"), "doesn't matter", true, false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, tt.contextValue, tt.contextValueExists)
			defer assertMockCalled()
			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: tt.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(tt.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestPropNotEndsWithCriteriaEvaluation() {
	operator := prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF
	contextPropertyName := "user.email"
	defaultValueToMatch := internal.CreateConfigValue([]string{"example.com", "splat"})

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
		{"returns true when valueToMatch is not a string slice", internal.CreateConfigValue("example.com"), "doesn't matter", true, true},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, tt.contextValue, tt.contextValueExists)
			defer assertMockCalled()
			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: tt.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(tt.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestPropIsOneOf() {
	operator := prefabProto.Criterion_PROP_IS_ONE_OF
	contextPropertyName := "user.email.domain"
	defaultValueToMatch := internal.CreateConfigValue([]string{"yahoo.com", "example.com"})

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
		{"returns false when valueToMatch is not a string slice", internal.CreateConfigValue("example.com"), "doesn't matter", true, false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, tt.contextValue, tt.contextValueExists)
			defer assertMockCalled()
			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: tt.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(tt.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestPropIsNotOneOf() {
	operator := prefabProto.Criterion_PROP_IS_NOT_ONE_OF
	contextPropertyName := "user.email.domain"
	defaultValueToMatch := internal.CreateConfigValue([]string{"yahoo.com", "example.com"})

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
		{"returns true when valueToMatch is not a string slice", internal.CreateConfigValue("example.com"), "doesn't matter", true, true},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, tt.contextValue, tt.contextValueExists)
			defer assertMockCalled()
			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: tt.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(tt.expected, isMatch)
		})
	}
}

func (suite *ConfigRuleTestSuite) TestHierarchicalMatch() {
	operator := prefabProto.Criterion_HIERARCHICAL_MATCH
	contextPropertyName := "team.path"
	defaultValueToMatch := internal.CreateConfigValue("a.b.c")

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
		{"returns false when valueToMatch is not a string", internal.CreateConfigValue(100), "a.b.c.d", true, false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, tt.contextValue, tt.contextValueExists)
			defer assertMockCalled()
			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: tt.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(tt.expected, isMatch)
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
		{"returns false when value to match isn't an int range", internal.CreateConfigValue("what's up doc"), nil, false, false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, tt.contextValue, tt.contextValueExists)
			defer assertMockCalled()
			criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: tt.valueToMatch, PropertyName: contextPropertyName}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(tt.expected, isMatch)
		})
	}

}

func (suite *ConfigRuleTestSuite) TestInSegmentCriterion() {
	// this test is based on in-int range as the target segment
	contextPropertyName := "team.size"
	segmentTargetIntRangeConfig := suite.createInIntRangeSegmentTarget(prefabProto.Criterion_IN_INT_RANGE, contextPropertyName, Int64Ptr(0), Int64Ptr(100))
	targetConfigKey := segmentTargetIntRangeConfig.Key
	defaultValueToMatch := internal.CreateConfigValue(targetConfigKey)

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

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, tt.contextValue, tt.contextValueExists)
			mockContext.On("GetValue", "").Return(nil, false)
			defer assertMockCalled()
			assertMockConfigStoreGetterCalled := suite.setupMockConfigStoreGetter(targetConfigKey, tt.targetSegmentConfig, tt.targetSegmentExists)
			defer assertMockConfigStoreGetterCalled()
			criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_IN_SEG, ValueToMatch: tt.valueToMatch}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(tt.expected, isMatch)
		})
	}

}

func (suite *ConfigRuleTestSuite) TestNotInSegmentCriterion() {
	// this test is based on in-int range as the target segment
	contextPropertyName := "team.size"
	segmentTargetIntRangeConfig := suite.createInIntRangeSegmentTarget(prefabProto.Criterion_IN_INT_RANGE, contextPropertyName, Int64Ptr(0), Int64Ptr(100))
	targetConfigKey := segmentTargetIntRangeConfig.Key
	defaultValueToMatch := internal.CreateConfigValue(targetConfigKey)

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

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockContext, assertMockCalled := suite.setupMockContext(contextPropertyName, tt.contextValue, tt.contextValueExists)
			mockContext.On("GetValue", "").Return(nil, false)
			defer assertMockCalled()
			assertMockConfigStoreGetterCalled := suite.setupMockConfigStoreGetter(targetConfigKey, tt.targetSegmentConfig, tt.targetSegmentExists)
			defer assertMockConfigStoreGetterCalled()
			criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_NOT_IN_SEG, ValueToMatch: tt.valueToMatch}
			isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
			suite.Equal(tt.expected, isMatch)
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
				ProjectEnvId: Int64Ptr(suite.projectEnvId),
				Values: []*prefabProto.ConditionalValue{
					{
						Criteria: []*prefabProto.Criterion{inIntRangeCriterion},
						Value:    internal.CreateConfigValue(true),
					},
					{
						Value: internal.CreateConfigValue(false),
					},
				},
			},
		},
	}
	return targetSegmentConfig
}

func (suite *ConfigRuleTestSuite) setupMockContext(contextPropertyName string, contextValue interface{}, contextExistsValue bool) (mockedContext *MockContextGetter, cleanup func()) {
	mockContext := MockContextGetter{}
	if contextExistsValue {
		// When context exists, return the provided contextValue and true.
		mockContext.On("GetValue", contextPropertyName).Return(contextValue, contextExistsValue)
	} else {
		// When context doesn't exist, ensure the mock returns nil for the contextValue (or a suitable zero value) and false.
		// This simulates the absence of the value correctly.
		mockContext.On("GetValue", contextPropertyName).Return(nil, false)
	}
	cleanupFunc := func() {
		mockContext.AssertExpectations(suite.T())
	}
	return &mockContext, cleanupFunc
}

type ContextMocking struct {
	contextPropertyName string
	value               interface{}
	exists              bool
}

func (suite *ConfigRuleTestSuite) setupMockContextWithMultipleValues(mockings []ContextMocking) (mockedContext *MockContextGetter, cleanup func()) {
	mockContext := &MockContextGetter{}

	for _, mocking := range mockings {
		contextValue := mocking.value
		contextExistsValue := mocking.exists

		if contextExistsValue {
			mockContext.On("GetValue", mocking.contextPropertyName).Return(contextValue, true)
		} else {
			mockContext.On("GetValue", mocking.contextPropertyName).Return(nil, false)
		}
	}

	cleanupFunc := func() {
		mockContext.AssertExpectations(suite.T())
	}

	return mockContext, cleanupFunc
}

func (suite *ConfigRuleTestSuite) setupMockConfigStoreGetter(configKey string, config *prefabProto.Config, configExists bool) (cleanup func()) {
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
