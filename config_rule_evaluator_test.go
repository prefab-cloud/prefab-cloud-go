package prefab

import (
	"github.com/prefab-cloud/prefab-cloud-go/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MockConfigStoreGetter struct {
	mock.Mock
}

func (m *MockConfigStoreGetter) GetConfig(key string) (config *prefabProto.Config, exists bool) {
	args := m.Called(key)
	return args.Get(0).(*prefabProto.Config), args.Bool(1)
}

type MockContextGetter struct {
	mock.Mock
}

func (m *MockContextGetter) GetValue(propertyName string) (value interface{}, valueExists bool) {
	args := m.Called(propertyName)
	return args.Get(0).(interface{}), args.Bool(1)

}

type ConfigRuleTestSuite struct {
	suite.Suite
	contextSet            *ContextSet
	evaluator             *ConfigRuleEvaluator
	projectEnvId          int64
	mockConfigStoreGetter MockConfigStoreGetter
}

func (suite *ConfigRuleTestSuite) SetupSuite() {
	suite.projectEnvId = 101
	suite.evaluator = NewConfigRuleEvaluator(&suite.mockConfigStoreGetter, suite.projectEnvId)
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
	suite.Run("returns true", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_ALWAYS_TRUE}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().True(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestNotSetCriteria() {
	suite.Run("returns false", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_NOT_SET}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().False(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestPropEndsWithCriteriaEvaluation() {
	suite.Run("returns true for email ending with example.com", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF, ValueToMatch: internal.CreateConfigValue([]string{"example.com", "splat"}), PropertyName: "user.email"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns false for email ending with prefab.cloud", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF, ValueToMatch: internal.CreateConfigValue([]string{"prefab.cloud", "splat"}), PropertyName: "user.email"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns false for property that does not exist", func() {
		criterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_PROP_ENDS_WITH_ONE_OF, ValueToMatch: internal.CreateConfigValue([]string{"prefab.cloud", "splat"}), PropertyName: "not.here"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().False(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestPropNotEndsWithCriteriaEvaluation() {
	operator := prefabProto.Criterion_PROP_DOES_NOT_END_WITH_ONE_OF

	suite.Run("returns false for email ending with example.com", func() {
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: internal.CreateConfigValue([]string{"example.com", "splat"}), PropertyName: "user.email"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().False(isMatch)
	})

	suite.Run("true for email ending with prefab.cloud", func() {
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: internal.CreateConfigValue([]string{"prefab.cloud", "splat"}), PropertyName: "user.email"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns true for property that does not exist", func() {
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: internal.CreateConfigValue([]string{"prefab.cloud", "splat"}), PropertyName: "not.here"}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, suite.contextSet)
		suite.Assert().True(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestPropIsOneOf() {
	operator := prefabProto.Criterion_PROP_IS_ONE_OF
	contextPropertyName := "user.email.domain"
	defaultValueToMatch := internal.CreateConfigValue([]string{"yahoo.com", "example.com"})
	setupMockContext := func(contextValue interface{}, contextExistsValue bool) (mockedContext *MockContextGetter, cleanup func()) {
		mockContext := MockContextGetter{}
		mockContext.On("GetValue", contextPropertyName).Return(contextValue, contextExistsValue)
		cleanupFunc := func() {
			mockContext.AssertExpectations(suite.T())
		}
		return &mockContext, cleanupFunc
	}

	suite.Run("returns true when in set", func() {
		mockContext, assertMockCalled := setupMockContext("yahoo.com", true)
		defer assertMockCalled()
		mockContext.On("GetValue", contextPropertyName).Return("yahoo.com", true)
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns false when not in set", func() {
		mockContext, assertMockCalled := setupMockContext("gmail.com", true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns false when context value is not a string", func() {
		mockContext, assertMockCalled := setupMockContext(123, true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns false when context value does not exist", func() {
		mockContext, assertMockCalled := setupMockContext(123, false)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns false when valueToMatch is not a string slice", func() {
		mockContext, assertMockCalled := setupMockContext("gmail.com", true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: internal.CreateConfigValue("example.com"), PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestPropIsNotOneOf() {
	operator := prefabProto.Criterion_PROP_IS_NOT_ONE_OF
	contextPropertyName := "user.email.domain"
	defaultValueToMatch := internal.CreateConfigValue([]string{"yahoo.com", "example.com"})
	setupMockContext := func(contextValue interface{}, contextExistsValue bool) (mockedContext *MockContextGetter, cleanup func()) {
		mockContext := MockContextGetter{}
		mockContext.On("GetValue", contextPropertyName).Return(contextValue, contextExistsValue)
		cleanupFunc := func() {
			mockContext.AssertExpectations(suite.T())
		}
		return &mockContext, cleanupFunc
	}

	suite.Run("returns false when in set", func() {
		mockContext, assertMockCalled := setupMockContext("yahoo.com", true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns true when not in set", func() {
		mockContext, assertMockCalled := setupMockContext("gmail.com", true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns true when context value is not a string", func() {
		mockContext, assertMockCalled := setupMockContext(123, true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns true when context value does not exist", func() {
		mockContext, assertMockCalled := setupMockContext(123, false)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns false when valueToMatch is not a string slice", func() {
		mockContext, assertMockCalled := setupMockContext("gmail.com", true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: internal.CreateConfigValue("example.com"), PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().True(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestHierarchicalMatch() {
	operator := prefabProto.Criterion_HIERARCHICAL_MATCH
	contextPropertyName := "team.path"
	defaultValueToMatch := internal.CreateConfigValue("a.b.c")
	setupMockContext := func(contextValue interface{}, contextExistsValue bool) (mockedContext *MockContextGetter, cleanup func()) {
		mockContext := MockContextGetter{}
		mockContext.On("GetValue", contextPropertyName).Return(contextValue, contextExistsValue)
		cleanupFunc := func() {
			mockContext.AssertExpectations(suite.T())
		}
		return &mockContext, cleanupFunc
	}

	suite.Run("returns true when matched", func() {
		mockContext, assertMockCalled := setupMockContext("a.b.c.d", true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns false when not matched", func() {
		mockContext, assertMockCalled := setupMockContext("a.b", true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns false when context value is not a string", func() {
		mockContext, assertMockCalled := setupMockContext(123, true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns false when context value does not exist", func() {
		mockContext, assertMockCalled := setupMockContext(1, false)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns false when valueToMatch is not a string", func() {
		mockContext, assertMockCalled := setupMockContext("gmail.com", true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: internal.CreateConfigValue(1234), PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestInIntRangeCriterion() {
	operator := prefabProto.Criterion_IN_INT_RANGE
	contextPropertyName := "team.size"
	defaultValueToMatch := &prefabProto.ConfigValue{Type: &prefabProto.ConfigValue_IntRange{IntRange: &prefabProto.IntRange{Start: Int64Ptr(0), End: Int64Ptr(100)}}}
	setupMockContext := func(contextValue interface{}, contextExistsValue bool) (mockedContext *MockContextGetter, cleanup func()) {
		mockContext := MockContextGetter{}
		mockContext.On("GetValue", contextPropertyName).Return(contextValue, contextExistsValue)
		cleanupFunc := func() {
			mockContext.AssertExpectations(suite.T())
		}
		return &mockContext, cleanupFunc
	}

	suite.Run("returns true when in range", func() {
		mockContext, assertMockCalled := setupMockContext(10, true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns true when equal to range start", func() {
		mockContext, assertMockCalled := setupMockContext(0, true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns false when equal to range end", func() {
		mockContext, assertMockCalled := setupMockContext(100, true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns true when in range (float)", func() {
		mockContext, assertMockCalled := setupMockContext(10.5, true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns true when in range treating empty start as long-min", func() {
		mockContext, assertMockCalled := setupMockContext(-100, true)
		defer assertMockCalled()
		valueToMatch := &prefabProto.ConfigValue{Type: &prefabProto.ConfigValue_IntRange{IntRange: &prefabProto.IntRange{End: Int64Ptr(100)}}}
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: valueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns false when context value is not an int", func() {
		mockContext, assertMockCalled := setupMockContext("123", true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns false when context value does not exist", func() {
		mockContext, assertMockCalled := setupMockContext("123", false)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: defaultValueToMatch, PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns false when valueToMatch is not an int range", func() {
		mockContext, assertMockCalled := setupMockContext("gmail.com", true)
		defer assertMockCalled()
		criterion := &prefabProto.Criterion{Operator: operator, ValueToMatch: internal.CreateConfigValue(1234), PropertyName: contextPropertyName}
		isMatch := suite.evaluator.EvaluateCriterion(criterion, mockContext)
		suite.Assert().False(isMatch)
	})
}

func (suite *ConfigRuleTestSuite) TestSegmentCriteria() {
	contextPropertyName := "team.size"
	segmentTargetIntRangeConfig := suite.createInIntRangeSegmentTarget(contextPropertyName, Int64Ptr(0), Int64Ptr(100))
	targetConfigKey := segmentTargetIntRangeConfig.Key
	defaultValueToMatch := internal.CreateConfigValue(targetConfigKey)

	setupMockContext := func(contextValue interface{}, contextExistsValue bool) (mockedContext *MockContextGetter, cleanup func()) {
		mockContext := MockContextGetter{}
		mockContext.On("GetValue", "").Return("", false)
		mockContext.On("GetValue", contextPropertyName).Return(contextValue, contextExistsValue)
		cleanupFunc := func() {
			mockContext.AssertExpectations(suite.T())
		}
		return &mockContext, cleanupFunc
	}

	setupMockConfigStoreGetter := func(targetSegmentConfig *prefabProto.Config, targetSegmentConfigExists bool) (cleanup func()) {
		suite.mockConfigStoreGetter.On("GetConfig", targetConfigKey).Return(targetSegmentConfig, targetSegmentConfigExists)
		cleanupFunc := func() {
			suite.mockConfigStoreGetter.AssertExpectations(suite.T())
		}
		return cleanupFunc

	}

	suite.Run("returns true when in segment and operator IN_SEGMENT", func() {
		mockContext, assertMockContextCalled := setupMockContext(10, true)
		defer assertMockContextCalled()
		assertMockConfigStoreGetterCalled := setupMockConfigStoreGetter(segmentTargetIntRangeConfig, true)
		defer assertMockConfigStoreGetterCalled()
		segmentCriterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_IN_SEG, ValueToMatch: defaultValueToMatch}

		isMatch := suite.evaluator.EvaluateCriterion(segmentCriterion, mockContext)
		suite.Assert().True(isMatch)
	})

	suite.Run("returns false when not in segment and operator IN_SEGMENT", func() {
		mockContext, assertMockContextCalled := setupMockContext(1000, true)
		defer assertMockContextCalled()
		assertMockConfigStoreGetterCalled := setupMockConfigStoreGetter(segmentTargetIntRangeConfig, true)
		defer assertMockConfigStoreGetterCalled()
		segmentCriterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_IN_SEG, ValueToMatch: defaultValueToMatch}
		isMatch := suite.evaluator.EvaluateCriterion(segmentCriterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns false when in segment and operator NOT_IN_SEGMENT", func() {
		mockContext, assertMockContextCalled := setupMockContext(10, true)
		defer assertMockContextCalled()
		assertMockConfigStoreGetterCalled := setupMockConfigStoreGetter(segmentTargetIntRangeConfig, true)
		defer assertMockConfigStoreGetterCalled()
		segmentCriterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_NOT_IN_SEG, ValueToMatch: defaultValueToMatch}

		isMatch := suite.evaluator.EvaluateCriterion(segmentCriterion, mockContext)
		suite.Assert().False(isMatch)
	})

	suite.Run("returns true when not in segment and operator NOT_IN_SEGMENT", func() {
		mockContext, assertMockContextCalled := setupMockContext(1000, true)
		defer assertMockContextCalled()
		assertMockConfigStoreGetterCalled := setupMockConfigStoreGetter(segmentTargetIntRangeConfig, true)
		defer assertMockConfigStoreGetterCalled()
		segmentCriterion := &prefabProto.Criterion{Operator: prefabProto.Criterion_NOT_IN_SEG, ValueToMatch: defaultValueToMatch}
		isMatch := suite.evaluator.EvaluateCriterion(segmentCriterion, mockContext)
		suite.Assert().True(isMatch)
	})

	// TODO: tests of in/not in segment behavior with non boolean config values, missing config

}
func (suite *ConfigRuleTestSuite) createInIntRangeSegmentTarget(contextPropertyName string, start *int64, end *int64) *prefabProto.Config {
	targetConfigKey := "teamSizeSegment"
	inIntRangeCriterion := &prefabProto.Criterion{
		Operator: prefabProto.Criterion_IN_INT_RANGE,
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestConfigRuleTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigRuleTestSuite))
}
