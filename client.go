package prefab

import (
	"errors"
	"fmt"
	"log/slog"
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

	var configStores []ConfigStoreGetter
	if options.ConfigOverrideDirectory != nil {
		configStores = append(configStores, NewLocalConfigStore(*options.ConfigOverrideDirectory, &options))
	}

	configStores = append(configStores, apiConfigStore)
	if options.ConfigDirectory != nil {
		configStores = append(configStores, NewLocalConfigStore(*options.ConfigDirectory, &options))
	}
	configStore := BuildCompositeConfigStore(configStores...)

	configResolver := NewConfigResolver(configStore, apiConfigStore, apiConfigStore)

	client := Client{options: &options, httpClient: httpClient, configStore: configStore, apiConfigStore: apiConfigStore, configResolver: configResolver, initializationComplete: make(chan struct{})}
	go client.fetchFromServer(0)

	return &client, nil
}

// TODO replace this with a fetcher type to manage first fetch and polling (plus SSE eventually)
func (c *Client) fetchFromServer(offset int32) {
	configs, err := c.httpClient.Load(offset)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to get data via http %v", err))
	}

	slog.Info("Loaded configuration data")

	c.apiConfigStore.SetFromConfigsProto(configs)
	c.closeInitializationCompleteOnce.Do(func() {
		close(c.initializationComplete)
	})
	slog.Info("Initialization complete called")

}

func clientInternalGetValueFunc[T any](key string, contextSet ContextSet, parseFunc func(*prefabProto.ConfigValue) (T, bool)) func(c *Client) (T, bool, error) {
	var zeroValue T

	return func(c *Client) (T, bool, error) {

		fetchResult, fetchOk, fetchErr := c.fetchAndProcessValue(key, contextSet, func(cv *prefabProto.ConfigValue) (any, bool) {
			pVal, pOk := clientParseValueWrapper(cv, parseFunc)
			if !pOk {
				return nil, false
			}
			return pVal, pOk
		})

		if fetchErr != nil || !fetchOk {
			return zeroValue, false, fetchErr
		}

		if fetchResult == nil {
			return zeroValue, false, nil
		}

		typedValue, ok := (fetchResult).(T)
		if !ok {
			slog.Warn("unexpected type for %T value: %T", zeroValue, fetchResult)
			return zeroValue, false, nil
		}

		return typedValue, true, nil
	}
}

// returnDefaultValue is a generic function that returns the default (zero) value for type T.
func returnZeroValue[T any]() T {
	var value T // Automatically initialized to the zero value of T
	return value
}

func (c *Client) GetRawConfigValueProto(key string, contextSet ContextSet) (*prefabProto.ConfigValue, bool, error) {
	resolutionResult, err := c.internalGetValue(key, contextSet)
	if err != nil {
		return nil, false, err
	}

	return resolutionResult.configValue, true, nil
}

//TODO: clients should return the default value if there's an incorrect type too -- these don't do that
// so if there's no value, return default (or nil)
// if there is a value check the type and return default if present. if not present then error
// should check the config value's type first before doing other go operations w/ interface{} and the like
//

func (c *Client) GetIntValueWithDefault(key string, contextSet ContextSet, defaultValue int64) (int64, bool) {
	value, ok, err := c.GetIntValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}
	return value, ok
}

func (c *Client) GetIntValue(key string, contextSet ContextSet) (int64, bool, error) {
	val, ok, err := clientInternalGetValueFunc(key, contextSet, utils.ExtractIntValue)(c)
	return val, ok, err
}

func (c *Client) GetStringValueWithDefault(key string, contextSet ContextSet, defaultValue string) (string, bool) {
	value, ok, err := c.GetStringValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}
	return value, ok
}

func (c *Client) GetStringValue(key string, contextSet ContextSet) (string, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, contextSet, utils.ExtractStringValue)(c)
	return value, ok, err
}

func (c *Client) GetBoolValueWithDefault(key string, contextSet ContextSet, defaultValue bool) (value bool, ok bool) {
	value, ok, err := c.GetBoolValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}
	return value, ok
}

func (c *Client) GetBoolValue(key string, contextSet ContextSet) (value bool, ok bool, err error) {
	value, ok, err = clientInternalGetValueFunc(key, contextSet, utils.ExtractBoolValue)(c)
	return value, ok, err
}

func (c *Client) GetFloatValueWithDefault(key string, contextSet ContextSet, defaultValue float64) (value float64, ok bool) {
	value, ok, err := c.GetFloatValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}
	return value, ok
}

func (c *Client) GetFloatValue(key string, contextSet ContextSet) (value float64, ok bool, err error) {
	value, ok, err = clientInternalGetValueFunc(key, contextSet, utils.ExtractFloatValue)(c)
	return value, ok, err
}

func (c *Client) GetStringSliceValueWithDefault(key string, contextSet ContextSet, defaultValue []string) (value []string, ok bool) {
	value, ok, err := c.GetStringSliceValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}
	return value, ok
}

func (c *Client) GetStringSliceValue(key string, contextSet ContextSet) (value []string, ok bool, err error) {
	value, ok, err = clientInternalGetValueFunc(key, contextSet, utils.ExtractStringListValue)(c)
	return value, ok, err
}

func (c *Client) GetDurationWithDefault(key string, contextSet ContextSet, defaultValue time.Duration) (value time.Duration, ok bool) {
	value, ok, err := c.GetDurationValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}
	return value, ok
}

func (c *Client) GetDurationValue(key string, contextSet ContextSet) (value time.Duration, ok bool, err error) {
	value, ok, err = clientInternalGetValueFunc(key, contextSet, utils.ExtractDurationValue)(c)
	return value, ok, err
}

func (c *Client) fetchAndProcessValue(key string, contextSet ContextSet, parser utils.ExtractValueFunction) (any, bool, error) {
	getResult, err := c.internalGetValue(key, contextSet)
	if err != nil {
		return nil, false, err
	}
	if getResult.configValue == nil {
		return nil, false, errors.New("config did not produce a result and no default is specified")
	}
	parsedValue, ok := parser(getResult.configValue)
	if !ok {
		return nil, false, nil
	}

	return parsedValue, true, nil
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

func clientParseValueWrapper[T any](cv *prefabProto.ConfigValue, parseFunc func(*prefabProto.ConfigValue) (T, bool)) (T, bool) {
	if cv != nil {
		pValue, pOk := parseFunc(cv)
		return pValue, pOk
	} else {
		var zeroValue T
		return zeroValue, false
	}
}

func (c *Client) awaitInitialization() (result awaitInitializationResult) {
	select {
	case <-c.initializationComplete:
		return SUCCESS
	case <-time.After(time.Duration(c.options.InitializationTimeoutSeconds) * time.Second):
		slog.Warn("Timeout expired, proceeding without waiting further")
		return TIMEOUT
	}
}

type awaitInitializationResult int

const (
	SUCCESS awaitInitializationResult = iota
	TIMEOUT
)
