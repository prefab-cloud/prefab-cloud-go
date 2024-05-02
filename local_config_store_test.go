package prefab

import (
	"github.com/prefab-cloud/prefab-cloud-go/utils"
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

func (s *LocalConfigStoreSuite) TestNewLocalConfigStore() {
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
				{key: "cool.bool.enabled", expected: utils.CreateConfigValue(true)},
				{key: "cool.bool", exists: false},
				{key: "hot.int", exists: false},
				{key: "sample_to_override", expected: utils.CreateConfigValue("value from override in default")},
				{key: "cool.count", expected: utils.CreateConfigValue(100)},
			},
		},
		{
			name:             "Default and production",
			source_directory: "testdata/local_configs",
			options: &Options{
				EnvironmentNames: []string{"production"},
			},
			expectedConfigs: []configExpectation{
				{key: "cool.bool.enabled", expected: utils.CreateConfigValue(false)},
				{key: "cool.bool", exists: false},
				{key: "hot.int", expected: utils.CreateConfigValue(212)},
				{key: "sample_to_override", expected: utils.CreateConfigValue("value from override in production")},
				{key: "cool.count", expected: utils.CreateConfigValue(100)},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			store := NewLocalConfigStore(tc.source_directory, tc.options)

			// Assert the expected configurations
			for _, expectation := range tc.expectedConfigs {
				config, exists := store.GetConfig(expectation.key)
				if expectation.expected != nil {
					s.Truef(exists, "Expected config with key '%s' to exist", expectation.key)
					s.NotNil(config, "Expected config with key '%s' to not be nil", expectation.key)
					if config != nil {
						value, onlyValue := s.onlyValue(config)
						s.Require().True(onlyValue, "Expected config with key '%s' to only have one value", expectation.key)
						s.Equal(expectation.expected, value, "Expected config with key '%s' to have value %v, got %v", expectation.key, expectation.expected, value)
					}
				} else {
					s.False(exists, "Expected config with key '%s' to not exist", expectation.key)
				}
			}
			s.True(store.initialized, "Expected initialized to be true")
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
