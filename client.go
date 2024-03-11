package prefab

import (
	"errors"
	"fmt"
	"github.com/prefab-cloud/prefab-cloud-go/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"sync"
	"time"
)

type Client struct {
	options                         *Options
	httpClient                      *HttpClient
	configStore                     *ApiConfigStore
	configResolver                  *ConfigResolver
	initializationComplete          chan struct{}
	closeInitializationCompleteOnce sync.Once
}

func NewClient(options Options) (*Client, error) {
	httpClient, err := BuildHttpClient(options)
	if err != nil {
		panic(err)
	}

	configStore := BuildApiConfigStore()
	configResolver := NewConfigResolver(configStore, configStore)

	client := Client{options: &options, httpClient: httpClient, configStore: configStore, configResolver: configResolver, initializationComplete: make(chan struct{})}
	go client.fetchFromServer(0)
	return &client, nil
}

func (c *Client) fetchFromServer(offset int32) {
	configs, err := c.httpClient.Load(offset)
	if err != nil {
		panic(err)
	}
	c.configStore.SetFromConfigsProto(configs)

	c.closeInitializationCompleteOnce.Do(func() {
		close(c.initializationComplete)
	})
}

func (c *Client) ManualLoad(configs *prefabProto.Configs) {
	c.configStore.SetFromConfigsProto(configs)
}

// returnDefaultValue is a generic function that returns the default (zero) value for type T.
func returnZeroValue[T any]() T {
	var value T // Automatically initialized to the zero value of T
	return value
}

func (c *Client) GetIntValueWithDefault(key string, contextSet ContextSet, defaultValue int64) (value int64, err error) {
	value, err = c.internalGetIntValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetIntValue(key string, contextSet ContextSet) (value int64, err error) {
	value, err = c.internalGetIntValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetIntValue(key string, contextSet ContextSet, defaultValue *int64) (value int64, err error) {
	var defaultVal *interface{}
	if defaultValue != nil {
		val := interface{}(*defaultValue) // Convert int to interface{}
		defaultVal = &val
	}
	fetchResult, fetchErr := c.fetchAndProcessValue(key, contextSet, defaultVal, func(cv *prefabProto.ConfigValue) (value interface{}, err error) {
		pVal, pErr := parseValueWrapper(cv, internal.ParseIntValue, defaultVal, 0)
		return pVal, pErr
	})
	if fetchErr != nil {
		return 0, fetchErr
	}
	return fetchResult.(int64), nil
}

func (c *Client) GetStringValueWithDefault(key string, contextSet ContextSet, defaultValue string) (value string, err error) {
	value, err = c.internalGetStringValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetStringValue(key string, contextSet ContextSet) (value string, err error) {
	value, err = c.internalGetStringValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetStringValue(key string, contextSet ContextSet, defaultValue *string) (value string, err error) {
	var defaultVal *interface{}
	if defaultValue != nil {
		val := interface{}(*defaultValue) // Convert string to interface{}
		defaultVal = &val
	}
	fetchResult, fetchErr := c.fetchAndProcessValue(key, contextSet, defaultVal, func(cv *prefabProto.ConfigValue) (value interface{}, err error) {
		pVal, pErr := parseValueWrapper(cv, internal.ParseStringValue, defaultVal, "")
		return pVal, pErr
	})
	if fetchErr != nil {
		return "", fetchErr
	}
	return fetchResult.(string), nil
}

func (c *Client) GetBoolValueWithDefault(key string, contextSet ContextSet, defaultValue bool) (value bool, err error) {
	value, err = c.internalGetBoolValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetBoolValue(key string, contextSet ContextSet) (value bool, err error) {
	value, err = c.internalGetBoolValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetBoolValue(key string, contextSet ContextSet, defaultValue *bool) (value bool, err error) {
	var defaultVal *interface{}
	if defaultValue != nil {
		val := interface{}(*defaultValue) // Convert bool to interface{}
		defaultVal = &val
	}
	fetchResult, fetchErr := c.fetchAndProcessValue(key, contextSet, defaultVal, func(cv *prefabProto.ConfigValue) (value interface{}, err error) {
		pVal, pErr := parseValueWrapper(cv, internal.ParseBoolValue, defaultVal, false)
		return pVal, pErr
	})
	if fetchErr != nil {
		return false, fetchErr
	}
	return fetchResult.(bool), nil
}

func (c *Client) GetFloatValueWithDefault(key string, contextSet ContextSet, defaultValue float64) (value float64, err error) {
	value, err = c.internalGetFloatValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetFloatValue(key string, contextSet ContextSet) (value float64, err error) {
	value, err = c.internalGetFloatValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetFloatValue(key string, contextSet ContextSet, defaultValue *float64) (value float64, err error) {
	var defaultVal *interface{}
	if defaultValue != nil {
		val := interface{}(*defaultValue) // Convert bool to interface{}
		defaultVal = &val
	}
	fetchResult, fetchErr := c.fetchAndProcessValue(key, contextSet, defaultVal, func(cv *prefabProto.ConfigValue) (value interface{}, err error) {
		pVal, pErr := parseValueWrapper(cv, internal.ParseFloatValue, defaultVal, 0)
		return pVal, pErr
	})
	if fetchErr != nil {
		return 0, fetchErr
	}
	return fetchResult.(float64), nil
}

func (c *Client) GetStringSliceValueWithDefault(key string, contextSet ContextSet, defaultValue []string) (value []string, err error) {
	value, err = c.internalGetStringSliceValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetStringSliceValue(key string, contextSet ContextSet) (value []string, err error) {
	value, err = c.internalGetStringSliceValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetStringSliceValue(key string, contextSet ContextSet, defaultValue *[]string) (value []string, err error) {
	var defaultVal *interface{}
	if defaultValue != nil {
		val := interface{}(*defaultValue) // Convert bool to interface{}
		defaultVal = &val
	}
	zeroValue := returnZeroValue[[]string]()
	fetchResult, fetchErr := c.fetchAndProcessValue(key, contextSet, defaultVal, func(cv *prefabProto.ConfigValue) (value interface{}, err error) {
		pVal, pErr := parseValueWrapper(cv, internal.ParseFloatValue, defaultVal, zeroValue)
		return pVal, pErr
	})
	if fetchErr != nil {
		return zeroValue, fetchErr
	}
	return fetchResult.([]string), nil
}

func parseValueWrapper(cv *prefabProto.ConfigValue, parseFunc internal.ConfigValueParseFunction, defaultValue *interface{}, zeroValue interface{}) (interface{}, error) {
	if cv != nil {
		pValue, pErr := parseFunc(cv)
		if pErr != nil {
			if defaultValue != nil {
				return *defaultValue, nil
			}
			return zeroValue, pErr
		}
		return pValue, nil
	} else {
		if defaultValue != nil {
			return *defaultValue, nil
		}
		return zeroValue, errors.New("config did not produce a result and no default is specified")
	}
}

func (c *Client) fetchAndProcessValue(key string, contextSet ContextSet, defaultValue *interface{}, parser internal.ConfigValueParseFunction) (interface{}, error) {
	getResult, err := c.internalGetValue(key, contextSet)
	if err != nil {
		return nil, err
	}
	if getResult.configValue == nil {
		if defaultValue != nil {
			return *defaultValue, nil
		}
		return nil, errors.New("config did not produce a result and no default is specified")
	}
	return parser(getResult.configValue)
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
