package internal_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/options"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type LocalConfigStoreSuite struct {
	suite.Suite
}

func TestLocalConfigStoreSuite(t *testing.T) {
	suite.Run(t, new(LocalConfigStoreSuite))
}

type configExpectation struct {
	expected *prefabProto.ConfigValue
	key      string
	// If expected is not specified, exists is assumed to be false
	exists bool
}

func (suite *LocalConfigStoreSuite) TestNewLocalConfigStore() {
	testCases := []struct {
		name            string
		sourceDirectory string
		options         *options.Options
		expectedConfigs []configExpectation
	}{
		{
			name:            "Default only",
			sourceDirectory: "testdata/local_configs",
			options: &options.Options{
				EnvironmentNames: []string{},
			},
			expectedConfigs: []configExpectation{
				{key: "cool.bool.enabled", expected: createConfigValueAndAssertOk(suite.T(), true)},
				{key: "cool.bool", exists: false},
				{key: "hot.int", exists: false},
				{key: "sample_to_override", expected: createConfigValueAndAssertOk(suite.T(), "value from override in default")},
				{key: "cool.count", expected: createConfigValueAndAssertOk(suite.T(), 100)},
			},
		},
		{
			name:            "Default and production",
			sourceDirectory: "testdata/local_configs",
			options: &options.Options{
				EnvironmentNames: []string{"production"},
			},
			expectedConfigs: []configExpectation{
				{key: "cool.bool.enabled", expected: createConfigValueAndAssertOk(suite.T(), false)},
				{key: "cool.bool", exists: false},
				{key: "hot.int", expected: createConfigValueAndAssertOk(suite.T(), 212)},
				{key: "sample_to_override", expected: createConfigValueAndAssertOk(suite.T(), "value from override in production")},
				{key: "cool.count", expected: createConfigValueAndAssertOk(suite.T(), 100)},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			store := internal.NewLocalConfigStore(tc.sourceDirectory, tc.options)

			// Assert the expected configurations
			for _, expectation := range tc.expectedConfigs {
				config, exists := store.GetConfig(expectation.key)
				if expectation.expected != nil {
					suite.Truef(exists, "Expected config with key '%s' to exist", expectation.key)
					suite.NotNil(config, "Expected config with key '%s' to not be nil", expectation.key)

					if config != nil {
						value, onlyValue := suite.onlyValue(config)
						suite.Require().True(onlyValue, "Expected config with key '%s' to only have one value", expectation.key)
						suite.Equal(expectation.expected, value, "Expected config with key '%s' to have value %v, got %v", expectation.key, expectation.expected, value)
					}
				} else {
					suite.False(exists, "Expected config with key '%s' to not exist", expectation.key)
				}
			}

			suite.True(store.Initialized, "Expected initialized to be true")
		})
	}
}

func (suite *LocalConfigStoreSuite) onlyValue(config *prefabProto.Config) (*prefabProto.ConfigValue, bool) {
	if len(config.GetRows()) != 1 {
		return nil, false
	}

	var value *prefabProto.ConfigValue

	for _, row := range config.GetRows() {
		for _, conditional := range row.GetValues() {
			if value != nil {
				return nil, false
			}

			value = conditional.GetValue()
		}
	}

	return value, true
}
