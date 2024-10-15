package internal_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/testutils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

// LocalConfigYamlParserTestSuite structure for running test suite.
type LocalConfigYamlParserTestSuite struct {
	suite.Suite
}

// SetupTest is a setup hook for each test in the suite.
func (s *LocalConfigYamlParserTestSuite) SetupTest() {
}

// createConfig helps to create a Config object for testing.
func (s *LocalConfigYamlParserTestSuite) createConfig(key string, cv *prefabProto.ConfigValue, configType prefabProto.ConfigType, valueType prefabProto.Config_ValueType) *prefabProto.Config {
	row := &prefabProto.ConfigRow{
		Values: []*prefabProto.ConditionalValue{{Value: cv}},
	}

	return &prefabProto.Config{
		Key:        key,
		Rows:       []*prefabProto.ConfigRow{row},
		ValueType:  valueType,
		ConfigType: configType,
	}
}

func (s *LocalConfigYamlParserTestSuite) collectKeys(configs []*prefabProto.Config) []string {
	keys := make([]string, 0, len(configs))
	for _, config := range configs {
		keys = append(keys, config.GetKey())
	}

	return keys
}

// TestLocalConfigYamlParser_parse tests the YAML parser function.
func (s *LocalConfigYamlParserTestSuite) TestLocalConfigYamlParser_parse() {
	tests := []struct {
		wantErr     assert.ErrorAssertionFunc
		name        string
		yamlInput   string
		wantConfigs []*prefabProto.Config
	}{
		{
			name:      "simple int",
			yamlInput: "sample_int: 123",
			wantConfigs: []*prefabProto.Config{
				s.createConfig("sample_int", testutils.CreateConfigValueAndAssertOk(s.T(), 123), prefabProto.ConfigType_CONFIG, prefabProto.Config_INT),
			},
			wantErr: assert.NoError,
		},
		{
			name:      "string feature flag",
			yamlInput: `flag_with_a_value: { "feature_flag": "true", value: "all-features" }`,
			wantConfigs: []*prefabProto.Config{
				s.createConfig("flag_with_a_value", testutils.CreateConfigValueAndAssertOk(s.T(), "all-features"), prefabProto.ConfigType_FEATURE_FLAG, prefabProto.Config_STRING),
			},
			wantErr: assert.NoError,
		},
		{
			name: "nested",
			yamlInput: `
prefab:
  api:
    liveness:
      timeout: 2
      enabled: true`,
			wantConfigs: []*prefabProto.Config{
				s.createConfig("prefab.api.liveness.timeout", testutils.CreateConfigValueAndAssertOk(s.T(), 2), prefabProto.ConfigType_CONFIG, prefabProto.Config_INT),
				s.createConfig("prefab.api.liveness.enabled", testutils.CreateConfigValueAndAssertOk(s.T(), true), prefabProto.ConfigType_CONFIG, prefabProto.Config_BOOL),
			},
			wantErr: assert.NoError,
		},
		{
			name: "underscore and log level",
			yamlInput: `
log-level:
  _: warn
  io:
    micronaut:
      web:
        _: info
        router:
          DefaultRouteBuilder: debug`,
			wantConfigs: []*prefabProto.Config{
				s.createConfig("log-level", testutils.CreateConfigValueAndAssertOk(s.T(), &prefabProto.ConfigValue_LogLevel{LogLevel: prefabProto.LogLevel_WARN}), prefabProto.ConfigType_LOG_LEVEL, prefabProto.Config_LOG_LEVEL),
				s.createConfig("log-level.io.micronaut.web", testutils.CreateConfigValueAndAssertOk(s.T(), &prefabProto.ConfigValue_LogLevel{LogLevel: prefabProto.LogLevel_INFO}), prefabProto.ConfigType_LOG_LEVEL, prefabProto.Config_LOG_LEVEL),
				s.createConfig("log-level.io.micronaut.web.router.DefaultRouteBuilder", testutils.CreateConfigValueAndAssertOk(s.T(), &prefabProto.ConfigValue_LogLevel{LogLevel: prefabProto.LogLevel_DEBUG}), prefabProto.ConfigType_LOG_LEVEL, prefabProto.Config_LOG_LEVEL),
			},
			wantErr: assert.NoError,
		},
	}

	for _, testCase := range tests {
		s.Run(testCase.name, func() {
			p := &internal.LocalConfigYamlParser{}

			gotConfigValues, projectEnvID, err := p.Parse([]byte(testCase.yamlInput))
			if !testCase.wantErr(s.T(), err, fmt.Sprintf("parse(%v)", testCase.yamlInput)) {
				return
			}

			s.ElementsMatch(s.collectKeys(testCase.wantConfigs), s.collectKeys(gotConfigValues))
			s.ElementsMatchf(testCase.wantConfigs, gotConfigValues, "parse(%v)", testCase.yamlInput)
			s.Equal(int64(0), projectEnvID)
		})
	}
}

// TestLocalConfigYamlParserTestSuite runs the test suite.
func TestLocalConfigYamlParserTestSuite(t *testing.T) {
	suite.Run(t, new(LocalConfigYamlParserTestSuite))
}
