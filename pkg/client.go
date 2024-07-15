package prefab

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	sse "github.com/r3labs/sse/v2"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/utils"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/options"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ConfigMatch = internal.ConfigMatch

type OnInitializationFailure = options.OnInitializationFailure

type ContextSet = contexts.ContextSet

type NamedContext = contexts.NamedContext

func NewContextSet() *ContextSet {
	return contexts.NewContextSet()
}

func NewContextSetFromProto(protoContextSet *prefabProto.ContextSet) *ContextSet {
	return contexts.NewContextSetFromProto(protoContextSet)
}

const (
	RAISE  OnInitializationFailure = options.RAISE
	UNLOCK OnInitializationFailure = options.UNLOCK
)

type ClientInterface interface {
	GetIntValue(key string, contextSet ContextSet) (int64, bool, error)
	GetBoolValue(key string, contextSet ContextSet) (bool, bool, error)
	GetStringValue(key string, contextSet ContextSet) (string, bool, error)
	GetFloatValue(key string, contextSet ContextSet) (float64, bool, error)
	GetStringSliceValue(key string, contextSet ContextSet) ([]string, bool, error)
	GetDurationValue(key string, contextSet ContextSet) (time.Duration, bool, error)
	GetIntValueWithDefault(key string, contextSet ContextSet, defaultValue int64) (int64, bool)
	GetBoolValueWithDefault(key string, contextSet ContextSet, defaultValue bool) (bool, bool)
	GetStringValueWithDefault(key string, contextSet ContextSet, defaultValue string) (string, bool)
	GetFloatValueWithDefault(key string, contextSet ContextSet, defaultValue float64) (float64, bool)
	GetStringSliceValueWithDefault(key string, contextSet ContextSet, defaultValue []string) ([]string, bool)
	GetDurationWithDefault(key string, contextSet ContextSet, defaultValue time.Duration) (time.Duration, bool)
	GetLogLevelStringValue(key string, contextSet ContextSet) (string, bool, error)
	GetJSONValue(key string, contextSet ContextSet) (interface{}, bool, error)
	GetJSONValueWithDefault(key string, contextSet ContextSet, defaultValue interface{}) (interface{}, bool)
	GetConfigMatch(key string, contextSet ContextSet) (*ConfigMatch, error)
	GetConfig(key string) (*prefabProto.Config, bool)
	FeatureIsOn(key string, contextSet ContextSet) (bool, bool)
	WithContext(contextSet *ContextSet) *boundClient
}

type boundClient struct {
	context *ContextSet
	client  *Client
}

type Client struct {
	boundClient                     *boundClient
	options                         *options.Options
	sseClient                       *sse.Client
	configStore                     internal.ConfigStoreGetter
	configResolver                  *internal.ConfigResolver
	initializationComplete          chan struct{}
	closeInitializationCompleteOnce sync.Once
}

func NewClient(opts ...Option) (*Client, error) {
	options := options.DefaultOptions

	var client Client

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	slog.Debug("Initializing client", "options", options)

	var configStores []internal.ConfigStoreGetter

	apiSourceFinishedLoading := func() {
		client.closeInitializationCompleteOnce.Do(func() {
			close(client.initializationComplete)
		})
	}

	anyAsync := false

	for _, source := range options.Sources {
		configStore, asyncInit, err := internal.BuildConfigStore(options, source, apiSourceFinishedLoading)

		if err != nil {
			return nil, err
		}

		if asyncInit {
			anyAsync = true
		}

		configStores = append(configStores, configStore)
	}

	configStore := internal.BuildCompositeConfigStore(configStores...)

	configResolver := internal.NewConfigResolver(configStore)

	client = Client{options: &options, configStore: configStore, configResolver: configResolver, initializationComplete: make(chan struct{})}

	if !anyAsync {
		client.closeInitializationCompleteOnce.Do(func() {
			close(client.initializationComplete)
		})
	}

	client.boundClient = &boundClient{client: &client, context: options.GlobalContext}

	return &client, nil
}

func (c *Client) GetIntValue(key string, contextSet ContextSet) (int64, bool, error) {
	return c.boundClient.GetIntValue(key, contextSet)
}

func (c *Client) GetBoolValue(key string, contextSet ContextSet) (bool, bool, error) {
	return c.boundClient.GetBoolValue(key, contextSet)
}

func (c *Client) GetStringValue(key string, contextSet ContextSet) (string, bool, error) {
	return c.boundClient.GetStringValue(key, contextSet)
}

func (c *Client) GetFloatValue(key string, contextSet ContextSet) (float64, bool, error) {
	return c.boundClient.GetFloatValue(key, contextSet)
}

func (c *Client) GetStringSliceValue(key string, contextSet ContextSet) ([]string, bool, error) {
	return c.boundClient.GetStringSliceValue(key, contextSet)
}

func (c *Client) GetDurationValue(key string, contextSet ContextSet) (time.Duration, bool, error) {
	return c.boundClient.GetDurationValue(key, contextSet)
}

func (c *Client) GetIntValueWithDefault(key string, contextSet ContextSet, defaultValue int64) (int64, bool) {
	return c.boundClient.GetIntValueWithDefault(key, contextSet, defaultValue)
}

func (c *Client) GetBoolValueWithDefault(key string, contextSet ContextSet, defaultValue bool) (bool, bool) {
	return c.boundClient.GetBoolValueWithDefault(key, contextSet, defaultValue)
}

func (c *Client) GetStringValueWithDefault(key string, contextSet ContextSet, defaultValue string) (string, bool) {
	return c.boundClient.GetStringValueWithDefault(key, contextSet, defaultValue)
}

func (c *Client) GetFloatValueWithDefault(key string, contextSet ContextSet, defaultValue float64) (float64, bool) {
	return c.boundClient.GetFloatValueWithDefault(key, contextSet, defaultValue)
}

func (c *Client) GetStringSliceValueWithDefault(key string, contextSet ContextSet, defaultValue []string) ([]string, bool) {
	return c.boundClient.GetStringSliceValueWithDefault(key, contextSet, defaultValue)
}

func (c *Client) GetDurationWithDefault(key string, contextSet ContextSet, defaultValue time.Duration) (time.Duration, bool) {
	return c.boundClient.GetDurationWithDefault(key, contextSet, defaultValue)
}

func (c *Client) GetJSONValue(key string, contextSet ContextSet) (interface{}, bool, error) {
	return c.boundClient.GetJSONValue(key, contextSet)
}

func (c *Client) GetJSONValueWithDefault(key string, contextSet ContextSet, defaultValue interface{}) (interface{}, bool) {
	return c.boundClient.GetJSONValueWithDefault(key, contextSet, defaultValue)
}

func (c *Client) FeatureIsOn(key string, contextSet ContextSet) (bool, bool) {
	return c.boundClient.FeatureIsOn(key, contextSet)
}

func (c *Client) GetLogLevelStringValue(key string, contextSet ContextSet) (string, bool, error) {
	return c.boundClient.GetLogLevelStringValue(key, contextSet)
}

func (c *Client) WithContext(contextSet *ContextSet) *boundClient {
	mergedContext := contexts.Merge(c.options.GlobalContext, contextSet)

	return &boundClient{context: mergedContext, client: c}
}

func (c *Client) GetConfig(key string) (*prefabProto.Config, bool) {
	return c.boundClient.GetConfig(key)
}

func (c *Client) GetConfigMatch(key string, contextSet ContextSet) (*ConfigMatch, error) {
	return c.boundClient.GetConfigMatch(key, contextSet)
}

func (c *Client) Keys() ([]string, error) {
	if c.awaitInitialization() == TIMEOUT {
		switch c.options.OnInitializationFailure {
		case options.UNLOCK:
			c.closeInitializationCompleteOnce.Do(func() {
				close(c.initializationComplete)
			})
		case options.RAISE:
			return []string{}, errors.New("initialization timeout")
		}
	}

	return c.configResolver.Keys(), nil
}

func clientInternalGetValueFunc[T any](key string, parentContextSet *contexts.ContextSet, contextSet contexts.ContextSet, parseFunc func(*prefabProto.ConfigValue) (T, bool)) func(c *boundClient) (T, bool, error) {
	var zeroValue T

	mergedContextSet := *contexts.Merge(parentContextSet, &contextSet)

	return func(c *boundClient) (T, bool, error) {
		fetchResult, fetchOk, fetchErr := c.fetchAndProcessValue(key, mergedContextSet, func(cv *prefabProto.ConfigValue) (any, bool) {
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
			slog.Warn(fmt.Sprintf("unexpected type for %T value: %T", zeroValue, fetchResult))

			return zeroValue, false, nil
		}

		return typedValue, true, nil
	}
}

func (c *boundClient) GetIntValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue int64) (int64, bool) {
	value, ok, err := c.GetIntValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *boundClient) GetIntValue(key string, contextSet contexts.ContextSet) (int64, bool, error) {
	val, ok, err := clientInternalGetValueFunc(key, c.context, contextSet, utils.ExtractIntValue)(c)

	return val, ok, err
}

func (c *boundClient) GetStringValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue string) (string, bool) {
	value, ok, err := c.GetStringValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *boundClient) GetStringValue(key string, contextSet contexts.ContextSet) (string, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, c.context, contextSet, utils.ExtractStringValue)(c)

	return value, ok, err
}

func (c *boundClient) GetJSONValue(key string, contextSet contexts.ContextSet) (interface{}, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, c.context, contextSet, utils.ExtractJSONValueWithoutError)(c)

	return value, ok, err
}

func (c *boundClient) GetJSONValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue interface{}) (interface{}, bool) {
	value, ok, err := c.GetJSONValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *boundClient) FeatureIsOn(key string, contextSet contexts.ContextSet) (bool, bool) {
	value, ok := c.GetBoolValueWithDefault(key, contextSet, false)

	return value, ok
}

func (c *boundClient) GetLogLevelStringValue(key string, contextSet contexts.ContextSet) (string, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, c.context, contextSet, utils.ExtractLogLevelValue)(c)

	if err != nil || !ok {
		return "", false, err
	}

	return value.String(), true, nil
}

func (c *boundClient) GetBoolValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue bool) (bool, bool) {
	value, ok, err := c.GetBoolValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *boundClient) GetBoolValue(key string, contextSet contexts.ContextSet) (bool, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, c.context, contextSet, utils.ExtractBoolValue)(c)

	return value, ok, err
}

func (c *boundClient) GetFloatValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue float64) (float64, bool) {
	value, ok, err := c.GetFloatValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *boundClient) GetFloatValue(key string, contextSet contexts.ContextSet) (float64, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, c.context, contextSet, utils.ExtractFloatValue)(c)

	return value, ok, err
}

func (c *boundClient) GetStringSliceValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue []string) ([]string, bool) {
	value, ok, err := c.GetStringSliceValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *boundClient) GetStringSliceValue(key string, contextSet contexts.ContextSet) ([]string, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, c.context, contextSet, utils.ExtractStringListValue)(c)

	return value, ok, err
}

func (c *boundClient) GetDurationWithDefault(key string, contextSet contexts.ContextSet, defaultValue time.Duration) (time.Duration, bool) {
	value, ok, err := c.GetDurationValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *boundClient) GetDurationValue(key string, contextSet contexts.ContextSet) (time.Duration, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, c.context, contextSet, utils.ExtractDurationValue)(c)

	return value, ok, err
}

func (c *boundClient) fetchAndProcessValue(key string, contextSet contexts.ContextSet, parser utils.ExtractValueFunction) (any, bool, error) {
	getResult, err := c.client.internalGetValue(key, contextSet)
	if err != nil {
		return nil, false, err
	}

	if getResult.match.Match == nil {
		return nil, false, errors.New("config did not produce a result and no default is specified")
	}

	parsedValue, ok := parser(getResult.match.Match)
	if !ok {
		return nil, false, nil
	}

	return parsedValue, true, nil
}

func (c *boundClient) WithContext(contextSet *ContextSet) *boundClient {
	mergedContext := contexts.Merge(c.context, contextSet)

	return &boundClient{context: mergedContext, client: c.client}
}

func raw(cv *prefabProto.ConfigValue) (*prefabProto.ConfigValue, bool) {
	return cv, true
}

func (c *boundClient) GetConfig(key string) (*prefabProto.Config, bool) {
	return c.client.configStore.GetConfig(key)
}

func (c *boundClient) GetConfigMatch(key string, contextSet ContextSet) (*ConfigMatch, error) {
	getResult, err := c.client.internalGetValue(key, contextSet)
	if err != nil {
		return nil, err
	}

	return &getResult.match, nil
}

func (c *Client) internalGetValue(key string, contextSet contexts.ContextSet) (resolutionResult, error) {
	if c.awaitInitialization() == TIMEOUT {
		switch c.options.OnInitializationFailure {
		case options.UNLOCK:
			c.closeInitializationCompleteOnce.Do(func() {
				close(c.initializationComplete)
			})
		case options.RAISE:
			return resolutionResultError(), errors.New("initialization timeout")
		}
	}

	match, err := c.configResolver.ResolveValue(key, &contextSet)
	if err != nil {
		return resolutionResultError(), err
	}

	return resolutionResultSuccess(match), nil
}

type resolutionResultType int

const (
	ResolutionResultTypeSuccess resolutionResultType = iota
	ResolutionResultTypeError
)

type resolutionResult struct {
	match      ConfigMatch
	resultType resolutionResultType
}

func resolutionResultError() resolutionResult {
	return resolutionResult{
		resultType: ResolutionResultTypeError,
	}
}

func resolutionResultSuccess(match ConfigMatch) resolutionResult {
	return resolutionResult{
		resultType: ResolutionResultTypeSuccess,
		match:      match,
	}
}

func clientParseValueWrapper[T any](cv *prefabProto.ConfigValue, parseFunc func(*prefabProto.ConfigValue) (T, bool)) (T, bool) {
	if cv != nil {
		pValue, pOk := parseFunc(cv)

		return pValue, pOk
	}

	var zeroValue T

	return zeroValue, false
}

func (c *Client) awaitInitialization() awaitInitializationResult {
	select {
	case <-c.initializationComplete:
		return SUCCESS
	case <-time.After(time.Duration(c.options.InitializationTimeoutSeconds) * time.Second):
		slog.Warn(fmt.Sprintf("%f second timeout expired, proceeding without waiting further. Configure in options `InitializationTimeoutSeconds`", c.options.InitializationTimeoutSeconds))

		return TIMEOUT
	}
}

type awaitInitializationResult int

const (
	SUCCESS awaitInitializationResult = iota
	TIMEOUT
)

func ExtractValue(cv *prefabProto.ConfigValue) (any, bool, error) {
	return utils.ExtractValue(cv)
}
