package prefab

import (
	"testing"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/stretchr/testify/suite"
)

type LocalConfigStoreSuite struct {
	suite.Suite
}

func TestLocalConfigStoreSuite(t *testing.T) {
	suite.Run(t, new(LocalConfigStoreSuite))
}

type configExpectation struct {
	key      string
	expected *prefabProto.ConfigValue
	// If expected is not specified, exists is assumed to be false
	exists bool
}

func (suite *LocalConfigStoreSuite) TestNewLocalConfigStore() {
	testCases := []struct {
		name             string
		source_directory string
		options          *Options
		expectedConfigs  []configExpectation
	}{
		{
			name:             "Default only",
			source_directory: "testdata/local_configs",
			options: &Options{
				EnvironmentNames: []string{},
			},
			expectedConfigs: []configExpectation{
				{key: "cool.bool.enabled", expected: createConfigValueAndAssertOk(true, suite.T())},
				{key: "cool.bool", exists: false},
				{key: "hot.int", exists: false},
				{key: "sample_to_override", expected: createConfigValueAndAssertOk("value from override in default", suite.T())},
				{key: "cool.count", expected: createConfigValueAndAssertOk(100, suite.T())},
			},
		},
		{
			name:             "Default and production",
			source_directory: "testdata/local_configs",
			options: &Options{
				EnvironmentNames: []string{"production"},
			},
			expectedConfigs: []configExpectation{
				{key: "cool.bool.enabled", expected: createConfigValueAndAssertOk(false, suite.T())},
				{key: "cool.bool", exists: false},
				{key: "hot.int", expected: createConfigValueAndAssertOk(212, suite.T())},
				{key: "sample_to_override", expected: createConfigValueAndAssertOk("value from override in production", suite.T())},
				{key: "cool.count", expected: createConfigValueAndAssertOk(100, suite.T())},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			store := NewLocalConfigStore(tc.source_directory, tc.options)

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

			suite.True(store.initialized, "Expected initialized to be true")
		})
	}
}

func (s *LocalConfigStoreSuite) onlyValue(config *prefabProto.Config) (value *prefabProto.ConfigValue, onlyValue bool) {
	if len(config.Rows) != 1 {
		return nil, false
	}

	value = nil

	for _, row := range config.Rows {
		for _, conditional := range row.Values {
			if value != nil {
				return nil, false
			}

			value = conditional.Value
		}
	}

	return value, true
}
