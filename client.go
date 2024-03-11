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

func (c *Client) GetIntValueWithDefault(key string, contextSet ContextSet, defaultValue int64) (value int64, err error) {
	value, err = c.internalGetIntValue(key, contextSet, &defaultValue)
	return value, err
}

func (c *Client) GetIntValue(key string, contextSet ContextSet) (value int64, err error) {
	value, err = c.internalGetIntValue(key, contextSet, nil)
	return value, err
}

func (c *Client) internalGetIntValue(key string, contextSet ContextSet, defaultValue *int64) (value int64, err error) {
	defaultValueAvailable := defaultValue != nil
	if c.awaitInitialization() == TIMEOUT {
		if defaultValueAvailable {
			return *defaultValue, nil
		}
		if c.options.OnInitializationFailure == RAISE {
			return 0, errors.New("initialization timeout and no default specified")
		}
	}
	match, err := c.configResolver.ResolveValue(key, &contextSet)
	if err != nil {
		return 0, err
	}
	switch v := match.match.Type.(type) {
	case *prefabProto.ConfigValue_Int:
		return v.Int, nil
	}
	if defaultValueAvailable {
		return *defaultValue, nil
	}
	return 0, errors.New("config did not produce an int and no default is specified")

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
	defaultValueAvailable := defaultValue != nil
	if c.awaitInitialization() == TIMEOUT {
		if defaultValueAvailable {
			return *defaultValue, nil
		}
		if c.options.OnInitializationFailure == RAISE {
			return "", errors.New("initialization timeout and no default specified")
		}
	}
	match, err := c.configResolver.ResolveValue(key, &contextSet)
	if err != nil {
		return "", err
	}
	switch v := match.match.Type.(type) {
	case *prefabProto.ConfigValue_String_:
		return v.String_, nil
	}
	if defaultValueAvailable {
		return *defaultValue, nil
	}
	return "", errors.New("config did not produce an int and no default is specified")
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
