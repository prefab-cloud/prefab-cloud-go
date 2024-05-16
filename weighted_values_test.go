package prefab

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/prefab-cloud/prefab-cloud-go/mocks"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type MockRandomer struct {
	mock.Mock
}

func (m *MockRandomer) Float64() float64 {
	args := m.Called()
	return args.Get(0).(float64)
}

type MockHasher struct {
	mock.Mock
}

func (m *MockHasher) HashZeroToOne(value string) (float64, bool) {
	args := m.Called(value)
	return args.Get(0).(float64), args.Bool(1)
}

type WeightedValueResolverTestSuite struct {
	suite.Suite
	randomer              *MockRandomer
	hasher                *MockHasher
	weightedValueResolver *WeightedValueResolver
}

func (suite *WeightedValueResolverTestSuite) SetupTest() {
	suite.randomer = new(MockRandomer)
	suite.hasher = new(MockHasher)
	suite.weightedValueResolver = &WeightedValueResolver{rand: suite.randomer, hasher: suite.hasher}
}

func (suite *WeightedValueResolverTestSuite) TestValueSelectionInRandomCases() {
	wv1 := &prefabProto.WeightedValue{
		Weight: 100,
		Value:  createConfigValueAndAssertOk(1, suite.T()),
	}
	wv2 := &prefabProto.WeightedValue{
		Weight: 50,
		Value:  createConfigValueAndAssertOk(2, suite.T()),
	}
	wv3 := &prefabProto.WeightedValue{
		Weight: 50,
		Value:  createConfigValueAndAssertOk(3, suite.T()),
	}
	weightedValuesWithoutHashValue := &prefabProto.WeightedValues{
		WeightedValues: []*prefabProto.WeightedValue{
			wv1, wv2, wv3,
		},
	}
	weightedValuesWithHashValue := &prefabProto.WeightedValues{
		HashByPropertyName: stringPtr("some.property"),
		WeightedValues: []*prefabProto.WeightedValue{
			wv1, wv2, wv3,
		},
	}

	tests := []struct {
		weightedValues *prefabProto.WeightedValues
		expectedValue  *prefabProto.ConfigValue
		contextMocking *mocks.ContextMocking
		name           string
		expectedIndex  int
		randomValue    float64
	}{
		{
			name:           "Three weighted values returns first value with low random because no hash by property name is set",
			weightedValues: weightedValuesWithoutHashValue,
			expectedValue:  wv1.Value,
			expectedIndex:  0,
			randomValue:    0.1,
			contextMocking: nil,
		},
		{
			name:           "Three weighted values returns second value with medium random because no hash by property name is set",
			weightedValues: weightedValuesWithoutHashValue,
			expectedValue:  wv2.Value,
			expectedIndex:  1,
			randomValue:    0.51,
			contextMocking: nil,
		},
		{
			name:           "Three weighted values returns third value with high random because no hash by property name is set",
			weightedValues: weightedValuesWithoutHashValue,
			expectedValue:  wv3.Value,
			expectedIndex:  2,
			randomValue:    0.99,
			contextMocking: nil,
		},
		{
			name:           "Three weighted values returns third value with high random because hash by property name doesn't return a context value",
			weightedValues: weightedValuesWithHashValue,
			expectedValue:  wv3.Value,
			expectedIndex:  2,
			randomValue:    0.99,
			contextMocking: &mocks.ContextMocking{ContextPropertyName: "some.property"},
		},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			suite.SetupTest() // call this manually to reset instances at this "sub" test level
			suite.randomer.On("Float64").Return(testCase.randomValue)

			var mockings []mocks.ContextMocking
			if testCase.contextMocking != nil {
				mockings = append(mockings, *testCase.contextMocking)
			}

			result, index := suite.weightedValueResolver.Resolve(testCase.weightedValues, "property name", mocks.NewMockContextWithMultipleValues(mockings))

			suite.Equal(testCase.expectedValue, result)
			suite.Equal(testCase.expectedIndex, index)
			suite.randomer.AssertExpectations(suite.T())
		})
	}
}

func (suite *WeightedValueResolverTestSuite) TestValueSelectionInHashingCases() {
	wv1 := &prefabProto.WeightedValue{
		Weight: 100,
		Value:  createConfigValueAndAssertOk(1, suite.T()),
	}
	wv2 := &prefabProto.WeightedValue{
		Weight: 50,
		Value:  createConfigValueAndAssertOk(2, suite.T()),
	}
	wv3 := &prefabProto.WeightedValue{
		Weight: 50,
		Value:  createConfigValueAndAssertOk(3, suite.T()),
	}
	hashByPropertyName := "some.property"
	weightedValuesWithHashValue := &prefabProto.WeightedValues{
		HashByPropertyName: &hashByPropertyName,
		WeightedValues: []*prefabProto.WeightedValue{
			wv1, wv2, wv3,
		},
	}

	tests := []struct {
		weightedValues       *prefabProto.WeightedValues
		expectedValue        *prefabProto.ConfigValue
		contextMocking       *mocks.ContextMocking
		name                 string
		expectedHashArgument string
		expectedIndex        int
		hashValue            float64
	}{
		{
			name:                 "Three weighted values returns first value with low hash",
			weightedValues:       weightedValuesWithHashValue,
			expectedValue:        wv1.Value,
			expectedIndex:        0,
			hashValue:            0.1,
			contextMocking:       &mocks.ContextMocking{ContextPropertyName: hashByPropertyName, Value: "james", Exists: true},
			expectedHashArgument: "property namejames",
		},
		{
			name:                 "Three weighted values returns second value with low hash",
			weightedValues:       weightedValuesWithHashValue,
			expectedValue:        wv1.Value,
			expectedIndex:        0,
			hashValue:            0.1,
			contextMocking:       &mocks.ContextMocking{ContextPropertyName: hashByPropertyName, Value: 100, Exists: true},
			expectedHashArgument: "property name100",
		},
	}

	for _, testCase := range tests {
		suite.Run(testCase.name, func() {
			suite.SetupTest() // call this manually to reset instances at this "sub" test level
			suite.hasher.On("HashZeroToOne", testCase.expectedHashArgument).Return(testCase.hashValue, true)

			var mockings []mocks.ContextMocking
			if testCase.contextMocking != nil {
				mockings = append(mockings, *testCase.contextMocking)
			}

			result, index := suite.weightedValueResolver.Resolve(testCase.weightedValues, "property name", mocks.NewMockContextWithMultipleValues(mockings))
			suite.Equal(testCase.expectedValue, result)
			suite.Equal(testCase.expectedIndex, index)
			suite.randomer.AssertExpectations(suite.T())
			suite.hasher.AssertExpectations(suite.T())
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestWeightedValueResolverTestSuite(t *testing.T) {
	suite.Run(t, new(WeightedValueResolverTestSuite))
}
