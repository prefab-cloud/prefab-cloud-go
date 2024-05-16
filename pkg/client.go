package prefab

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/options"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/prefab-cloud/prefab-cloud-go/utils"
)

type Options = options.Options

type Option func(*options.Options) error

type OnInitializationFailure = options.OnInitializationFailure

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
		// TODO: validate API key
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

type Client struct {
	options                         *options.Options
	httpClient                      *internal.HTTPClient
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

	client := Client{options: &options, httpClient: httpClient, configStore: configStore, apiConfigStore: apiConfigStore, configResolver: configResolver, initializationComplete: make(chan struct{})}
	go client.fetchFromServer(0)

	return &client, nil
}

// TODO: replace this with a fetcher type to manage first fetch and polling (plus SSE eventually)
func (c *Client) fetchFromServer(offset int32) {
	configs, err := c.httpClient.Load(offset) // retry me!
	if err != nil {
		slog.Error(fmt.Sprintf("unable to get data via http %v", err))
	} else {
		slog.Info("Loaded configuration data")

		c.apiConfigStore.SetFromConfigsProto(configs)
		c.closeInitializationCompleteOnce.Do(func() {
			close(c.initializationComplete)
		})
		slog.Info("Initialization complete called")
	}
}

func clientInternalGetValueFunc[T any](key string, contextSet internal.ContextSet, parseFunc func(*prefabProto.ConfigValue) (T, bool)) func(c *Client) (T, bool, error) {
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
			slog.Warn(fmt.Sprintf("unexpected type for %T value: %T", zeroValue, fetchResult))

			return zeroValue, false, nil
		}

		return typedValue, true, nil
	}
}

func (c *Client) GetRawConfigValueProto(key string, contextSet internal.ContextSet) (*prefabProto.ConfigValue, bool, error) {
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

func (c *Client) GetIntValueWithDefault(key string, contextSet internal.ContextSet, defaultValue int64) (int64, bool) {
	value, ok, err := c.GetIntValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *Client) GetIntValue(key string, contextSet internal.ContextSet) (int64, bool, error) {
	val, ok, err := clientInternalGetValueFunc(key, contextSet, utils.ExtractIntValue)(c)

	return val, ok, err
}

func (c *Client) GetStringValueWithDefault(key string, contextSet internal.ContextSet, defaultValue string) (string, bool) {
	value, ok, err := c.GetStringValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *Client) GetStringValue(key string, contextSet internal.ContextSet) (string, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, contextSet, utils.ExtractStringValue)(c)

	return value, ok, err
}

func (c *Client) FeatureIsOn(key string, contextSet internal.ContextSet) (bool, bool) {
	value, ok := c.GetBoolValueWithDefault(key, contextSet, false)

	return value, ok
}

func (c *Client) GetBoolValueWithDefault(key string, contextSet internal.ContextSet, defaultValue bool) (bool, bool) {
	value, ok, err := c.GetBoolValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *Client) GetBoolValue(key string, contextSet internal.ContextSet) (bool, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, contextSet, utils.ExtractBoolValue)(c)

	return value, ok, err
}

func (c *Client) GetFloatValueWithDefault(key string, contextSet internal.ContextSet, defaultValue float64) (float64, bool) {
	value, ok, err := c.GetFloatValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *Client) GetFloatValue(key string, contextSet internal.ContextSet) (float64, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, contextSet, utils.ExtractFloatValue)(c)

	return value, ok, err
}

func (c *Client) GetStringSliceValueWithDefault(key string, contextSet internal.ContextSet, defaultValue []string) ([]string, bool) {
	value, ok, err := c.GetStringSliceValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *Client) GetStringSliceValue(key string, contextSet internal.ContextSet) ([]string, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, contextSet, utils.ExtractStringListValue)(c)

	return value, ok, err
}

func (c *Client) GetDurationWithDefault(key string, contextSet internal.ContextSet, defaultValue time.Duration) (time.Duration, bool) {
	value, ok, err := c.GetDurationValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

func (c *Client) GetDurationValue(key string, contextSet internal.ContextSet) (time.Duration, bool, error) {
	value, ok, err := clientInternalGetValueFunc(key, contextSet, utils.ExtractDurationValue)(c)

	return value, ok, err
}

func (c *Client) fetchAndProcessValue(key string, contextSet internal.ContextSet, parser utils.ExtractValueFunction) (any, bool, error) {
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

func (c *Client) internalGetValue(key string, contextSet internal.ContextSet) (resolutionResult, error) {
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
