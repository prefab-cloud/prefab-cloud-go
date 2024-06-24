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
	"github.com/prefab-cloud/prefab-cloud-go/pkg/options"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/prefab-cloud/prefab-cloud-go/utils"
)

type Options = options.Options

type Option func(*options.Options) error

type OnInitializationFailure = options.OnInitializationFailure

type ContextSet = contexts.ContextSet

const (
	RAISE  OnInitializationFailure = options.RAISE
	UNLOCK OnInitializationFailure = options.UNLOCK
)

func WithConfigDirectory(configDirectory string) Option {
	return func(o *options.Options) error {
		o.ConfigDirectory = &configDirectory

		return nil
	}
}

func WithGlobalContext(globalContext *ContextSet) Option {
	return func(o *options.Options) error {
		o.GlobalContext = globalContext

		return nil
	}
}

func WithConfigOverrideDirectory(configOverrideDirectory string) Option {
	return func(o *options.Options) error {
		o.ConfigOverrideDirectory = &configOverrideDirectory

		return nil
	}
}

func WithEnvironmentNames(environmentNames []string) Option {
	return func(o *options.Options) error {
		o.EnvironmentNames = environmentNames

		return nil
	}
}

func WithAPIKey(apiKey string) Option {
	return func(o *options.Options) error {
		o.APIKey = apiKey

		return nil
	}
}

func WithAPIURL(apiURL string) Option {
	return func(o *options.Options) error {
		o.APIUrl = apiURL

		return nil
	}
}

func WithDatasource(datasource options.Datasource) Option {
	return func(o *options.Options) error {
		o.Datasource = datasource

		return nil
	}
}

func WithInitializationTimeoutSeconds(timeoutSeconds float64) Option {
	return func(o *options.Options) error {
		o.InitializationTimeoutSeconds = timeoutSeconds

		return nil
	}
}

func WithOnInitializationFailure(onInitializationFailure options.OnInitializationFailure) Option {
	return func(o *options.Options) error {
		o.OnInitializationFailure = onInitializationFailure

		return nil
	}
}

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
	httpClient                      *internal.HTTPClient
	sseClient                       *sse.Client
	configStore                     internal.ConfigStoreGetter
	apiConfigStore                  *internal.APIConfigStore // temporary until wrapped in a fetcher/auto updater
	configResolver                  *internal.ConfigResolver
	initializationComplete          chan struct{}
	closeInitializationCompleteOnce sync.Once
}

func NewClient(opts ...Option) (*Client, error) {
	options := options.DefaultOptions

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	slog.Info("Initializing client", "options", options)

	httpClient, err := internal.BuildHTTPClient(options)
	if err != nil {
		panic(err)
	}

	apiConfigStore := internal.BuildAPIConfigStore()

	var configStores []internal.ConfigStoreGetter
	if options.ConfigOverrideDirectory != nil {
		configStores = append(configStores, internal.NewLocalConfigStore(*options.ConfigOverrideDirectory, &options))
	}

	configStores = append(configStores, apiConfigStore)
	if options.ConfigDirectory != nil {
		configStores = append(configStores, internal.NewLocalConfigStore(*options.ConfigDirectory, &options))
	}

	configStore := internal.BuildCompositeConfigStore(configStores...)

	configResolver := internal.NewConfigResolver(configStore, apiConfigStore, apiConfigStore)

	sseClient, err := internal.BuildSSEClient(options)
	if err != nil {
		panic(err)
	}

	client := Client{options: &options, httpClient: httpClient, sseClient: sseClient, configStore: configStore, apiConfigStore: apiConfigStore, configResolver: configResolver, initializationComplete: make(chan struct{})}

	client.boundClient = &boundClient{client: &client, context: options.GlobalContext}

	go client.fetchFromServer(0, 0, func() {
		go internal.StartSSEConnection(sseClient, apiConfigStore)
	})

	return &client, nil
}

const maxRetries = 10

func (c *Client) fetchFromServer(offset int32, retriesAttempted int, then func()) {
	configs, err := c.httpClient.Load(offset)

	if err != nil {
		slog.Warn(fmt.Sprintf("unable to get data via http %v", err))

		if retriesAttempted < maxRetries {
			retryDelay := time.Duration(retriesAttempted) * time.Second

			slog.Info(fmt.Sprintf("retrying in %d seconds, attempt %d/%d", int(retryDelay.Seconds()), retriesAttempted+1, maxRetries))

			time.Sleep(retryDelay)

			c.fetchFromServer(offset, retriesAttempted+1, then)
		} else {
			slog.Error("max retries reached, giving up")
			then()
		}

		return
	} else {
		slog.Info("Loaded configuration data")

		c.apiConfigStore.SetFromConfigsProto(configs)
		c.closeInitializationCompleteOnce.Do(func() {
			close(c.initializationComplete)
		})
		slog.Info("Initialization complete called")

		then()
	}
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
	value, ok, err := clientInternalGetValueFunc(key, c.context, contextSet, utils.ExtractJSONValue)(c)

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

	if getResult.configValue == nil {
		return nil, false, errors.New("config did not produce a result and no default is specified")
	}

	parsedValue, ok := parser(getResult.configValue)
	if !ok {
		return nil, false, nil
	}

	return parsedValue, true, nil
}

func (c *boundClient) WithContext(contextSet *ContextSet) *boundClient {
	mergedContext := contexts.Merge(c.context, contextSet)

	return &boundClient{context: mergedContext, client: c.client}
}

func (c *Client) internalGetValue(key string, contextSet contexts.ContextSet) (resolutionResult, error) {
	if c.awaitInitialization() == TIMEOUT {
		if c.options.OnInitializationFailure == options.UNLOCK {
			c.closeInitializationCompleteOnce.Do(func() {
				close(c.initializationComplete)
			})
		} else if c.options.OnInitializationFailure == options.RAISE {
			return resolutionResultError(), errors.New("initialization timeout")
		}
	}

	match, err := c.configResolver.ResolveValue(key, &contextSet)
	if err != nil {
		return resolutionResultError(), err
	}

	return resolutionResultSuccess(match.Match), nil
}

type resolutionResultType int

const (
	ResolutionResultTypeSuccess resolutionResultType = iota
	ResolutionResultTypeError
)

type resolutionResult struct {
	configValue *prefabProto.ConfigValue
	resultType  resolutionResultType
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
