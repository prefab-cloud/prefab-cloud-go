package prefab

import (
	"errors"
	"github.com/prefab-cloud/prefab-cloud-go/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"os"
)

type ConfigMatch struct {
	isMatch                  bool
	originalMatch            *prefabProto.ConfigValue
	originalKey              string
	match                    *prefabProto.ConfigValue
	rowIndex                 int
	conditionalValueIndex    int
	weightedValueIndex       *int // pointer so we can check nil
	selectedConditionalValue *prefabProto.ConditionalValue
}

func NewConfigMatchFromConditionMatch(conditionMatch ConditionMatch) ConfigMatch {
	return ConfigMatch{isMatch: conditionMatch.isMatch,
		originalMatch:            conditionMatch.match,
		match:                    conditionMatch.match,
		rowIndex:                 conditionMatch.rowIndex,
		conditionalValueIndex:    conditionMatch.rowIndex,
		selectedConditionalValue: conditionMatch.selectedConditionalValue}
}

type RealEnvLookup struct{}

func (RealEnvLookup) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

var ErrConfigDoesNotExist = errors.New("config does not exist")
var ErrEnvVarNotExist = errors.New("environment variable does not exist")

type ConfigResolver struct {
	configStore           ConfigStoreGetter
	ruleEvaluator         ConfigEvaluator
	weightedValueResolver WeightedValueResolverIF
	decrypter             Decrypter
	envLookup             EnvLookup
}

func (c *ConfigResolver) ResolveValue(key string, contextSet ContextGetter) (configMatch ConfigMatch, err error) {

	config, configExists := c.configStore.GetConfig(key)
	if !configExists {
		return ConfigMatch{isMatch: false, originalKey: key}, ErrConfigDoesNotExist
	}
	ruleMatchResults := c.ruleEvaluator.EvaluateConfig(config, contextSet)
	configMatch = NewConfigMatchFromConditionMatch(ruleMatchResults)
	configMatch.originalKey = key
	switch v := ruleMatchResults.match.Type.(type) {
	case *prefabProto.ConfigValue_WeightedValues:
		result, index := c.handleWeightedValue(key, v.WeightedValues, contextSet)
		configMatch.weightedValueIndex = &index
		configMatch.match = result
	case *prefabProto.ConfigValue_Provided:
		provided := ruleMatchResults.match.GetProvided()
		if provided != nil {
			envValue, envValueExists := c.handleProvided(provided)
			if envValueExists {
				configMatch.match = internal.CreateConfigValue(envValue)
			} else {
				return configMatch, ErrEnvVarNotExist
			}
		}
	case *prefabProto.ConfigValue_String_:
		if ruleMatchResults.match.GetDecryptWith() != "" {
			decryptedValue, err := c.handleDecryption(ruleMatchResults.match, contextSet)
			if err == nil {
				configMatch.match = internal.CreateConfigValue(decryptedValue)
				configMatch.match.Confidential = boolPtr(true)
			} else {
				return configMatch, err
			}
		}
	}
	originalMatchIsConfidential := configMatch.originalMatch.GetConfidential()
	if configMatch.originalMatch != configMatch.match && ruleMatchResults.match.GetDecryptWith() == "" && originalMatchIsConfidential {
		configMatch.match.Confidential = configMatch.originalMatch.Confidential
	}

	return configMatch, nil
}

func boolPtr(val bool) *bool {
	return &val
}

func (c *ConfigResolver) handleProvided(provided *prefabProto.Provided) (value string, ok bool) {
	switch provided.GetSource() {
	case prefabProto.ProvidedSource_ENV_VAR:
		if provided.Lookup != nil {
			envValue, envValueExists := c.envLookup.LookupEnv(provided.GetLookup())
			return envValue, envValueExists
		}
	}
	return "", false
}

func (c *ConfigResolver) handleDecryption(configValue *prefabProto.ConfigValue, contextSet ContextGetter) (decryptedValue string, err error) {
	config, configExists := c.configStore.GetConfig(configValue.GetDecryptWith())
	if configExists {
		match := c.ruleEvaluator.EvaluateConfig(config, contextSet)
		if match.isMatch {
			value, err := c.decrypter.DecryptValue(match.match.GetString_(), configValue.GetString_())
			if err != nil {
				return "", err
			}
			return value, nil
		} else {
			return "", errors.New("no match in config value") //TODO
		}

	}
	return "", errors.New("no config value exists") // todo

}

func (c *ConfigResolver) handleWeightedValue(configKey string, values *prefabProto.WeightedValues, contextSet ContextGetter) (valueResult *prefabProto.ConfigValue, index int) {
	value, index := c.weightedValueResolver.Resolve(values, configKey, contextSet)
	return value, index
}
