package prefab

import (
	"errors"
	"os"
	"strconv"
	"time"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/prefab-cloud/prefab-cloud-go/utils"
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
	return ConfigMatch{
		isMatch:                  conditionMatch.isMatch,
		originalMatch:            conditionMatch.match,
		match:                    conditionMatch.match,
		rowIndex:                 conditionMatch.rowIndex,
		conditionalValueIndex:    conditionMatch.rowIndex,
		selectedConditionalValue: conditionMatch.selectedConditionalValue,
	}
}

type RealEnvLookup struct{}

func (RealEnvLookup) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

var (
	ErrConfigDoesNotExist = errors.New("config does not exist")
	ErrEnvVarNotExist     = errors.New("environment variable does not exist")
)

type ConfigResolver struct {
	configStore           ConfigStoreGetter
	ruleEvaluator         ConfigEvaluator
	weightedValueResolver WeightedValueResolverIF
	decrypter             Decrypter
	envLookup             EnvLookup
	contextGetter         ContextValueGetter
}

func NewConfigResolver(configStore ConfigStoreGetter, supplier ProjectEnvIdSupplier, apiContextGetter ContextValueGetter) *ConfigResolver {
	return &ConfigResolver{
		configStore:           configStore,
		ruleEvaluator:         NewConfigRuleEvaluator(configStore, supplier),
		weightedValueResolver: NewWeightedValueResolver(time.Now().UnixNano(), &Hashing{}),
		decrypter:             &Encryption{},
		envLookup:             &RealEnvLookup{},
		contextGetter:         apiContextGetter,
	}
}

func (c ConfigResolver) ResolveValue(key string, contextSet ContextValueGetter) (configMatch ConfigMatch, err error) {
	config, configExists := c.configStore.GetConfig(key)
	if !configExists {
		return ConfigMatch{isMatch: false, originalKey: key}, ErrConfigDoesNotExist
	}
	contextSet = makeMultiContextGetter(contextSet, c.contextGetter)

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
				if coercedValue, coercionWorked := coerceValue(envValue, config.ValueType); coercionWorked {
					newValue, _ := utils.Create(coercedValue)
					configMatch.match = newValue
				}
			} else {
				return configMatch, ErrEnvVarNotExist
			}
		}
	case *prefabProto.ConfigValue_String_:
		if ruleMatchResults.match.GetDecryptWith() != "" {
			decryptedValue, err := c.handleDecryption(ruleMatchResults.match, contextSet)
			if err == nil {
				value, _ := utils.Create(decryptedValue)
				configMatch.match = value
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

func (c ConfigResolver) handleProvided(provided *prefabProto.Provided) (value string, ok bool) {
	switch provided.GetSource() {
	case prefabProto.ProvidedSource_ENV_VAR:
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
	config, configExists := c.configStore.GetConfig(configValue.GetDecryptWith())
	if configExists {
		match := c.ruleEvaluator.EvaluateConfig(config, contextSet)
		if match.isMatch {
			key, keyOk, _ := utils.ExtractValue(match.match)
			if keyOk {
				keyStr, ok := key.(string)
				if !ok {
					return "", errors.New("secret key is not a string")
				}
				decryptedValue, decryptionError := c.decrypter.DecryptValue(keyStr, configValue.GetString_())
				return decryptedValue, decryptionError
			} else {
				return "", errors.New("secret key lookup failed")
			}
		} else {
			return "", errors.New("no match in config value") // TODO
		}
	}

	return "", errors.New("no config value exists") // todo
}

func (c ConfigResolver) handleWeightedValue(configKey string, values *prefabProto.WeightedValues, contextSet ContextValueGetter) (valueResult *prefabProto.ConfigValue, index int) {
	value, index := c.weightedValueResolver.Resolve(values, configKey, contextSet)
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
