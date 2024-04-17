package prefab

import (
	"fmt"
	"testing"

	"github.com/prefab-cloud/prefab-cloud-go/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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

func (s *LocalConfigYamlParserTestSuite) collectKeys(configs []*prefabProto.Config) (keys []string) {
	for _, config := range configs {
		keys = append(keys, config.Key)
	}
	return keys
}

// TestLocalConfigYamlParser_parse tests the YAML parser function.
func (s *LocalConfigYamlParserTestSuite) TestLocalConfigYamlParser_parse() {
	tests := []struct {
		name        string
		yamlInput   string
		wantConfigs []*prefabProto.Config
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:      "simple int",
			yamlInput: "sample_int: 123",
			wantConfigs: []*prefabProto.Config{
				s.createConfig("sample_int", internal.CreateConfigValue(123), prefabProto.ConfigType_CONFIG, prefabProto.Config_INT),
			},
			wantErr: assert.NoError,
		},
		{
			name:      "string feature flag",
			yamlInput: `flag_with_a_value: { "feature_flag": "true", value: "all-features" }`,
			wantConfigs: []*prefabProto.Config{
				s.createConfig("flag_with_a_value", internal.CreateConfigValue("all-features"), prefabProto.ConfigType_FEATURE_FLAG, prefabProto.Config_STRING),
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
				s.createConfig("prefab.api.liveness.timeout", internal.CreateConfigValue(2), prefabProto.ConfigType_CONFIG, prefabProto.Config_INT),
				s.createConfig("prefab.api.liveness.enabled", internal.CreateConfigValue(true), prefabProto.ConfigType_CONFIG, prefabProto.Config_BOOL),
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
				s.createConfig("log-level", internal.CreateConfigValue(&prefabProto.ConfigValue_LogLevel{LogLevel: prefabProto.LogLevel_WARN}), prefabProto.ConfigType_LOG_LEVEL, prefabProto.Config_LOG_LEVEL),
				s.createConfig("log-level.io.micronaut.web", internal.CreateConfigValue(&prefabProto.ConfigValue_LogLevel{LogLevel: prefabProto.LogLevel_INFO}), prefabProto.ConfigType_LOG_LEVEL, prefabProto.Config_LOG_LEVEL),
				s.createConfig("log-level.io.micronaut.web.router.DefaultRouteBuilder", internal.CreateConfigValue(&prefabProto.ConfigValue_LogLevel{LogLevel: prefabProto.LogLevel_DEBUG}), prefabProto.ConfigType_LOG_LEVEL, prefabProto.Config_LOG_LEVEL),
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			p := &LocalConfigYamlParser{}
			gotConfigValues, err := p.parse([]byte(tt.yamlInput))
			if !tt.wantErr(s.T(), err, fmt.Sprintf("parse(%v)", tt.yamlInput)) {
				return
			}
			assert.ElementsMatch(s.T(), s.collectKeys(tt.wantConfigs), s.collectKeys(gotConfigValues))
			assert.ElementsMatchf(s.T(), tt.wantConfigs, gotConfigValues, "parse(%v)", tt.yamlInput)
		})
	}
}

// TestLocalConfigYamlParserTestSuite runs the test suite.
func TestLocalConfigYamlParserTestSuite(t *testing.T) {
	suite.Run(t, new(LocalConfigYamlParserTestSuite))
}
