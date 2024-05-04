package prefab

import (
	"errors"
	"fmt"
	"sync"
	"time"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/prefab-cloud/prefab-cloud-go/utils"
)

type Client struct {
	options                         *Options
	httpClient                      *HttpClient
	configStore                     ConfigStoreGetter
	apiConfigStore                  *ApiConfigStore // temporary until wrapped in a fetcher/auto updater
	configResolver                  *ConfigResolver
	initializationComplete          chan struct{}
	closeInitializationCompleteOnce sync.Once
}

func NewClient(options Options) (*Client, error) {
	httpClient, err := BuildHttpClient(options)
	if err != nil {
		panic(err)
	}

	apiConfigStore := BuildApiConfigStore()

	configStores := []ConfigStoreGetter{}
	if options.ConfigOverrideDirectory != nil {
		configStores = append(configStores, NewLocalConfigStore(*options.ConfigOverrideDirectory, &options))
	}
	configStores = append(configStores, apiConfigStore)
	if options.ConfigDirectory != nil {
		configStores = append(configStores, NewLocalConfigStore(*options.ConfigDirectory, &options))
	}
	configResolver := NewConfigResolver(apiConfigStore, apiConfigStore)
	configStore := BuildCompositeConfigStore(configStores...)

	client := Client{options: &options, httpClient: httpClient, configStore: configStore, apiConfigStore: apiConfigStore, configResolver: configResolver, initializationComplete: make(chan struct{})}
	go client.fetchFromServer(0)
	return &client, nil
}

// TODO replace this with a fetcher type to manage first fetch and polling (plus SSE eventually)
func (c *Client) fetchFromServer(offset int32) {
	configs, err := c.httpClient.Load(offset)
	if err != nil {
		panic(err)
	}
	c.apiConfigStore.SetFromConfigsProto(configs)
	c.closeInitializationCompleteOnce.Do(func() {
		close(c.initializationComplete)
	})
}
func clientInternalGetValueFunc[T any](key string, contextSet ContextSet, defaultValue *T, parseFunc func(*prefabProto.ConfigValue) (T, error), zeroValue T) func(c *Client) (*T, error) {
	return func(c *Client) (*T, error) {
		var defaultVal *interface{}
		// default value nil means no default value. Not allowed to pass in nil as a default
		if defaultValue != nil {
			val := interface{}(*defaultValue)
			defaultVal = &val
		}
		fetchResult, fetchErr := c.fetchAndProcessValue(key, contextSet, defaultVal, func(cv *prefabProto.ConfigValue) (value interface{}, err error) {
			pVal, pErr := clientParseValueWrapper(cv, parseFunc, defaultVal, zeroValue)
			return pVal, pErr
		})
		if fetchErr != nil {
			return &zeroValue, fetchErr
		}
		if fetchResult == nil {
			return &zeroValue, nil
		}
		typedValue, ok := (*fetchResult).(T)
		if !ok {
			return &zeroValue, fmt.Errorf("unexpected type for %T value: %T", zeroValue, *fetchResult)
		}
		return &typedValue, nil
	}
}

// returnDefaultValue is a generic function that returns the default (zero) value for type T.
func returnZeroValue[T any]() T {
	var value T // Automatically initialized to the zero value of T
	return value
}

func (c *Client) GetRawConfigValueProto(key string, contextSet ContextSet) (*prefabProto.ConfigValue, error) {
	resolutionResult, err := c.internalGetValue(key, contextSet)
	if err != nil {
		return nil, err
	}
	return resolutionResult.configValue, nil
}

//TODO: clients should return the default value if there's an incorrect type too -- these don't do that
// so if there's no value, return default (or nil)
// if there is a value check the type and return default if present. if not present then error
// should check the config value's type first before doing other go operations w/ interface{} and the like
//

func (c *Client) GetIntValueWithDefault(key string, contextSet ContextSet, defaultValue int64) (value *int64, err error) {
	value, err = c.internalGetIntValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetIntValue(key string, contextSet ContextSet) (value *int64, err error) {
	value, err = c.internalGetIntValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetIntValue(key string, contextSet ContextSet, defaultValue *int64) (*int64, error) {
	return clientInternalGetValueFunc(key, contextSet, defaultValue, utils.ParseIntValue, 0)(c)
}

func clientParseValueWrapper[T any](cv *prefabProto.ConfigValue, parseFunc func(*prefabProto.ConfigValue) (T, error), defaultValue *interface{}, zeroValue T) (T, error) {
	if cv != nil {
		pValue, pErr := parseFunc(cv)
		if pErr != nil {
			if defaultValue != nil {
				if defaultVal, ok := (*defaultValue).(T); ok {
					return defaultVal, nil
				}
			}
			return zeroValue, pErr
		}
		return pValue, nil
	} else {
		if defaultValue != nil {
			if defaultVal, ok := (*defaultValue).(T); ok {
				return defaultVal, nil
			}
		}
		return zeroValue, errors.New("config did not produce a result and no default is specified")
	}
}

func (c *Client) GetStringValueWithDefault(key string, contextSet ContextSet, defaultValue string) (value *string, err error) {
	value, err = c.internalGetStringValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetStringValue(key string, contextSet ContextSet) (value *string, err error) {
	value, err = c.internalGetStringValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetStringValue(key string, contextSet ContextSet, defaultValue *string) (*string, error) {
	return clientInternalGetValueFunc(key, contextSet, defaultValue, utils.ParseStringValue, "")(c)
}

func (c *Client) GetBoolValueWithDefault(key string, contextSet ContextSet, defaultValue bool) (value *bool, err error) {
	value, err = c.internalGetBoolValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetBoolValue(key string, contextSet ContextSet) (value *bool, err error) {
	value, err = c.internalGetBoolValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetBoolValue(key string, contextSet ContextSet, defaultValue *bool) (value *bool, err error) {
	return clientInternalGetValueFunc(key, contextSet, defaultValue, utils.ParseBoolValue, false)(c)

}

func (c *Client) GetFloatValueWithDefault(key string, contextSet ContextSet, defaultValue float64) (value *float64, err error) {
	value, err = c.internalGetFloatValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetFloatValue(key string, contextSet ContextSet) (value *float64, err error) {
	value, err = c.internalGetFloatValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetFloatValue(key string, contextSet ContextSet, defaultValue *float64) (value *float64, err error) {
	return clientInternalGetValueFunc(key, contextSet, defaultValue, utils.ParseFloatValue, 0)(c)
}

func (c *Client) GetStringSliceValueWithDefault(key string, contextSet ContextSet, defaultValue []string) (value *[]string, err error) {
	value, err = c.internalGetStringSliceValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetStringSliceValue(key string, contextSet ContextSet) (value *[]string, err error) {
	value, err = c.internalGetStringSliceValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetStringSliceValue(key string, contextSet ContextSet, defaultValue *[]string) (value *[]string, err error) {
	return clientInternalGetValueFunc(key, contextSet, defaultValue, utils.ParseStringListValue, []string{})(c)
}

func (c *Client) GetDurationWithDefault(key string, contextSet ContextSet, defaultValue time.Duration) (value *time.Duration, err error) {
	value, err = c.internalGetDurationValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetDurationValue(key string, contextSet ContextSet) (value *time.Duration, err error) {
	value, err = c.internalGetDurationValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetDurationValue(key string, contextSet ContextSet, defaultValue *time.Duration) (value *time.Duration, err error) {
	return clientInternalGetValueFunc(key, contextSet, defaultValue, utils.ParseDurationValue, time.Duration(0))(c)
}

func (c *Client) fetchAndProcessValue(key string, contextSet ContextSet, defaultValue *interface{}, parser utils.ConfigValueParseFunction) (*interface{}, error) {
	getResult, err := c.internalGetValue(key, contextSet)
	if err != nil {
		return nil, err
	}
	if getResult.configValue == nil {
		if defaultValue != nil {
			return defaultValue, nil
		}
		return nil, errors.New("config did not produce a result and no default is specified")
	}
	parsedValue, err := parser(getResult.configValue)
	if err != nil {
		return nil, err
	}
	return &parsedValue, nil
}

func (c *Client) internalGetValue(key string, contextSet ContextSet) (resolutionResult resolutionResult, err error) {
	if c.awaitInitialization() == TIMEOUT {
		if c.options.OnInitializationFailure == UNLOCK {
			c.closeInitializationCompleteOnce.Do(func() {
				close(c.initializationComplete)
			})
		} else if c.options.OnInitializationFailure == RAISE {
			return resolutionResultError(), errors.New("initialization timeout")
		}
	}
	match, err := c.configResolver.ResolveValue(key, &contextSet)
	if err != nil {
		return resolutionResultError(), err
	}
	return resolutionResultSuccess(match.match), nil
}

type resolutionResultType int

const (
	ResolutionResultTypeSuccess resolutionResultType = iota
	ResolutionResultTypeError
)

type resolutionResult struct {
	resultType  resolutionResultType
	configValue *prefabProto.ConfigValue
}

func resolutionResultError() resolutionResult {
	return resolutionResult{
		resultType: ResolutionResultTypeError,
	}
}

func resolutionResultSuccess(configValue *prefabProto.ConfigValue) resolutionResult {
	return resolutionResult{
		resultType:  ResolutionResultTypeSuccess,
		configValue: configValue,
	}
}

func (c *Client) awaitInitialization() (result awaitInitializationResult) {
	select {
	case <-c.initializationComplete:
		return SUCCESS
	case <-time.After(time.Duration(c.options.InitializationTimeoutSeconds) * time.Second):
		fmt.Println("Timeout expired, proceeding without waiting further")
		return TIMEOUT
	}
}

type awaitInitializationResult int

const (
	SUCCESS awaitInitializationResult = iota
	TIMEOUT
)
