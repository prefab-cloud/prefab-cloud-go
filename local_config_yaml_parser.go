package prefab

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/prefab-cloud/prefab-cloud-go/utils"
	"gopkg.in/yaml.v3"
)

type LocalConfigYamlParser struct{}

func (p *LocalConfigYamlParser) parse(yamlData []byte) (configValues []*prefabProto.Config, err error) {
	var data map[string]interface{}
	// Unmarshal the YAML into the map
	err = yaml.Unmarshal(yamlData, &data)
	if err != nil {
		return nil, err
	}

	var outputValues []*prefabProto.Config

	for mapKey, mapValue := range data {
		configValues, err := p.handleMapKeyValue([]string{}, mapKey, mapValue)
		if err != nil {
			return nil, err
		}

		outputValues = append(outputValues, configValues...)
	}

	return outputValues, nil
}

func (p *LocalConfigYamlParser) handleMapKeyValue(keyPath []string, mapKey string, mapValue interface{}) (configValues []*prefabProto.Config, err error) {
	configType := prefabProto.ConfigType_CONFIG

	switch value := mapValue.(type) {
	case map[string]interface{}:
		{
			featureFlagValue, featureFlagKeyExists := value["feature_flag"]
			if featureFlagKeyExists {
				isFeatureFlag, parsingWorked := p.coerceToBool(featureFlagValue)
				if isFeatureFlag && parsingWorked {
					configType = prefabProto.ConfigType_FEATURE_FLAG
				}

				configValueValue, configValueExists := value["value"]
				if !configValueExists {
					return nil, errors.New("yaml must contain 'value' key")
				}

				actualConfigValue, ok := utils.Create(configValueValue)
				if !ok {
					return nil, fmt.Errorf("unable to create configValue with %v", configValueValue)
				}

				newConfig, ok := p.createConfig(mapKey, actualConfigValue, configType)
				if !ok {
					return nil, fmt.Errorf("unable to create config with %v", actualConfigValue)
				}

				return []*prefabProto.Config{newConfig}, nil
			}

			var accumulatedConfigValues []*prefabProto.Config

			if underscoreValue, underscoreExists := value["_"]; underscoreExists {
				fullKeyName := strings.Join(append(keyPath, mapKey), ".")
				configValue, ok := p.createConfig(fullKeyName, underscoreValue, configType)

				if !ok {
					return nil, fmt.Errorf("unable to create config for key %s", fullKeyName)
				}

				accumulatedConfigValues = append(accumulatedConfigValues, configValue)
			}

			for nextMapKey, nextMapValue := range value {
				if nextMapKey == "_" {
					continue // already handled special case above
				}

				newConfigValues, err := p.handleMapKeyValue(append(keyPath, mapKey), nextMapKey, nextMapValue)
				if err != nil {
					return nil, err
				}

				accumulatedConfigValues = append(accumulatedConfigValues, newConfigValues...)
			}

			return accumulatedConfigValues, nil
		}

	default:
		newKey := strings.Join(append(keyPath, mapKey), ".")

		newConfig, ok := p.createConfig(newKey, mapValue, configType)
		if !ok {
			return nil, fmt.Errorf("unable to create config with %v", newKey)
		}

		return []*prefabProto.Config{newConfig}, nil
	}
}

func (p *LocalConfigYamlParser) coerceToBool(maybeBool interface{}) (value bool, ok bool) {
	switch val := maybeBool.(type) {
	case bool:
		return val, true
	case string:
		return strings.ToLower(val) == "true", true
	default:
		return false, false
	}
}

func (p *LocalConfigYamlParser) createConfig(key string, value any, configType prefabProto.ConfigType) (*prefabProto.Config, bool) {
	var configValue *prefabProto.ConfigValue

	if key == "log-level" || strings.HasPrefix(key, "log-level") {
		configType = prefabProto.ConfigType_LOG_LEVEL

		switch v := value.(type) {
		case string:
			logLevel, ok := prefabProto.LogLevel_value[strings.ToUpper(v)]
			if !ok {
				slog.Info("key %s has invalid log level: %s", key, v)
				return nil, false
			}

			configValue, ok = utils.Create(&prefabProto.ConfigValue_LogLevel{LogLevel: prefabProto.LogLevel(logLevel)})
		default:
			slog.Info("key %s should have a string value type but it was %T", key, value)
			return nil, false
		}
	} else {
		var ok bool

		configValue, ok = utils.Create(value)
		if !ok {
			slog.Info(fmt.Sprintf("create value failed for key %s", key))
			return nil, false
		}
	}

	row := &prefabProto.ConfigRow{
		Values: []*prefabProto.ConditionalValue{{Value: configValue}},
	}

	valueType := utils.GetValueType(configValue)

	return &prefabProto.Config{Key: key, Rows: []*prefabProto.ConfigRow{row}, ValueType: valueType, ConfigType: configType}, true
}

func (p *LocalConfigYamlParser) isSingleNestedMap(m map[string]interface{}) bool {
	return len(m) == 1 && reflect.TypeOf(m[p.getFirstKey(m)]).Kind() == reflect.Map
}

func (p *LocalConfigYamlParser) getFirstKey(m map[string]interface{}) string {
	for k := range m {
		return k
	}

	return ""
}
