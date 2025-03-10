package internal

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/utils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ConfigMatch struct {
	ConfigID              int64
	ConfigType            prefabProto.ConfigType
	OriginalMatch         *prefabProto.ConfigValue
	Match                 *prefabProto.ConfigValue
	ConfigKey             string
	WeightedValueIndex    *int
	RowIndex              *int
	ConditionalValueIndex *int
	IsMatch               bool
}

func NewConfigMatchFromConditionMatch(conditionMatch ConditionMatch) ConfigMatch {
	return ConfigMatch{
		IsMatch:               conditionMatch.IsMatch,
		OriginalMatch:         conditionMatch.Match,
		Match:                 conditionMatch.Match,
		RowIndex:              conditionMatch.RowIndex,
		ConditionalValueIndex: conditionMatch.ConditionalValueIndex,
	}
}

type RealEnvLookup struct{}

func (RealEnvLookup) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

var (
	ErrConfigDoesNotExist = errors.New("config does not exist")
	ErrEnvVarNotExist     = errors.New("environment variable does not exist")
	ErrTypeCoercionFailed = errors.New("type coercion failed on value from environment variable")
)

type ConfigResolver struct {
	ConfigStore           ConfigStoreGetter
	RuleEvaluator         ConfigEvaluator
	WeightedValueResolver WeightedValueResolverIF
	Decrypter             Decrypter
	EnvLookup             EnvLookup
	ContextGetter         ContextValueGetter
}

func NewConfigResolver(configStore ConfigStoreGetter) *ConfigResolver {
	return &ConfigResolver{
		ConfigStore:           configStore,
		RuleEvaluator:         NewConfigRuleEvaluator(configStore, configStore),
		WeightedValueResolver: NewWeightedValueResolver(time.Now().UnixNano(), &Hashing{}),
		Decrypter:             &Encryption{},
		EnvLookup:             &RealEnvLookup{},
		ContextGetter:         configStore,
	}
}

func (c ConfigResolver) Keys() []string {
	return c.ConfigStore.Keys()
}

func (c ConfigResolver) ResolveValue(key string, contextSet ContextValueGetter) (ConfigMatch, error) {
	config, configExists := c.ConfigStore.GetConfig(key)
	if !configExists {
		return ConfigMatch{IsMatch: false, ConfigKey: key}, ErrConfigDoesNotExist
	}

	return c.ResolveValueForConfig(config, contextSet, key)
}

func (c ConfigResolver) ResolveValueForConfig(config *prefabProto.Config, contextSet ContextValueGetter, key string) (ConfigMatch, error) {
	contextSet = makeMultiContextGetter(contextSet, c.ContextGetter)

	ruleMatchResults := c.RuleEvaluator.EvaluateConfig(config, contextSet)
	configMatch := NewConfigMatchFromConditionMatch(ruleMatchResults)
	configMatch.ConfigKey = key
	configMatch.ConfigType = config.GetConfigType()
	configMatch.ConfigID = config.GetId()

	switch v := ruleMatchResults.Match.GetType().(type) {
	case *prefabProto.ConfigValue_WeightedValues:
		result, index := c.handleWeightedValue(key, v.WeightedValues, contextSet)
		configMatch.WeightedValueIndex = &index
		configMatch.Match = result
	case *prefabProto.ConfigValue_Provided:
		provided := ruleMatchResults.Match.GetProvided()
		if provided != nil {
			envValue, envValueExists := c.handleProvided(provided)
			if envValueExists {
				if coercedValue, coercionWorked := coerceValue(envValue, config.GetValueType()); coercionWorked {
					newValue, _ := utils.Create(coercedValue)
					configMatch.Match = newValue
				} else {
					return configMatch, ErrTypeCoercionFailed
				}
			} else {
				return configMatch, ErrEnvVarNotExist
			}
		}
	case *prefabProto.ConfigValue_String_:
		if ruleMatchResults.Match.GetDecryptWith() != "" {
			decryptedValue, err := c.handleDecryption(ruleMatchResults.Match, contextSet)
			if err == nil {
				value, _ := utils.Create(decryptedValue)
				configMatch.Match = value
				configMatch.Match.Confidential = BoolPtr(true)
			} else {
				return configMatch, err
			}
		}
	}

	originalMatchIsConfidential := configMatch.OriginalMatch.GetConfidential()
	if configMatch.OriginalMatch != configMatch.Match && ruleMatchResults.Match.GetDecryptWith() == "" && originalMatchIsConfidential {
		configMatch.Match.Confidential = BoolPtr(configMatch.OriginalMatch.GetConfidential())
	}

	return configMatch, nil
}

func (c ConfigResolver) handleProvided(provided *prefabProto.Provided) (string, bool) {
	if provided.GetSource() == prefabProto.ProvidedSource_ENV_VAR {
		if provided.Lookup != nil {
			envValue, envValueExists := os.LookupEnv(provided.GetLookup())

			return envValue, envValueExists
		}
	}

	return "", false
}

func coerceValue(value string, valueType prefabProto.Config_ValueType) (any, bool) {
	switch valueType {
	case prefabProto.Config_STRING:
		return value, true
	case prefabProto.Config_NOT_SET_VALUE_TYPE: // could in this case get info on the requested type (ie which method was called?)
		return value, true
	case prefabProto.Config_INT:
		if number, err := strconv.Atoi(value); err == nil {
			return number, true
		}
	case prefabProto.Config_DOUBLE:
		if number, err := strconv.ParseFloat(value, 64); err == nil {
			return number, true
		}
	case prefabProto.Config_BOOL:
		if bValue, err := strconv.ParseBool(value); err == nil {
			return bValue, true
		}
	}

	return nil, false
}

func (c ConfigResolver) handleDecryption(configValue *prefabProto.ConfigValue, contextSet ContextValueGetter) (string, error) {
	config, configExists := c.ConfigStore.GetConfig(configValue.GetDecryptWith())
	if configExists {
		match := c.RuleEvaluator.EvaluateConfig(config, contextSet)
		if match.IsMatch {
			key, keyOk, _ := utils.ExtractValue(match.Match)
			if keyOk {
				keyStr, ok := key.(string)
				if !ok {
					return "", errors.New("secret key is not a string")
				}

				decryptedValue, decryptionError := c.Decrypter.DecryptValue(keyStr, configValue.GetString_())

				return decryptedValue, decryptionError
			}

			return "", errors.New("secret key lookup failed")
		}

		return "", errors.New("no match in config value")
	}

	return "", errors.New("no config value exists")
}

func (c ConfigResolver) handleWeightedValue(configKey string, values *prefabProto.WeightedValues, contextSet ContextValueGetter) (*prefabProto.ConfigValue, int) {
	value, index := c.WeightedValueResolver.Resolve(values, configKey, contextSet)

	return value, index
}

type multiContextGetter struct {
	contexts []ContextValueGetter
}

func makeMultiContextGetter(passedContext ContextValueGetter, implicitContext ContextValueGetter) multiContextGetter {
	return multiContextGetter{contexts: []ContextValueGetter{passedContext, implicitContext}}
}

func (c multiContextGetter) GetContextValue(propertyName string) (any, bool) {
	for _, context := range c.contexts {
		if value, valueExists := context.GetContextValue(propertyName); valueExists {
			return value, true
		}
	}

	return nil, false
}
