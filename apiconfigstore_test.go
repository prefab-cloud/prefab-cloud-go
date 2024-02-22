package prefab

import (
	"github.com/prefab-cloud/prefab-cloud-go/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"testing"
)

func Int64Ptr(val int64) *int64 {
	return &val
}

func RunTests(t *testing.T) {

	configFoo := &prefabProto.Config{
		Key:        "foo",
		Id:         10,
		ProjectId:  100,
		ConfigType: prefabProto.ConfigType_CONFIG,
		ValueType:  prefabProto.Config_STRING,
		Rows: []*prefabProto.ConfigRow{
			{
				ProjectEnvId: Int64Ptr(101),
				Values: []*prefabProto.ConditionalValue{
					{
						Value: internal.CreateConfigValue("foo-value"),
					},
				},
			},
		},
	}

	configFooTombstone := &prefabProto.Config{
		Key:        "foo",
		Id:         11,
		ProjectId:  100,
		ConfigType: prefabProto.ConfigType_CONFIG,
		ValueType:  prefabProto.Config_STRING,
	}

	configFooWithDifferentValue := &prefabProto.Config{
		Key:        "foo",
		Id:         11,
		ProjectId:  100,
		ConfigType: prefabProto.ConfigType_CONFIG,
		ValueType:  prefabProto.Config_STRING,
		Rows: []*prefabProto.ConfigRow{
			{
				ProjectEnvId: Int64Ptr(101),
				Values: []*prefabProto.ConditionalValue{
					{
						Value: internal.CreateConfigValue("foo-value-two"),
					},
				},
			},
		},
	}

	configBar := &prefabProto.Config{
		Key:        "bar",
		Id:         10,
		ProjectId:  100,
		ConfigType: prefabProto.ConfigType_CONFIG,
		ValueType:  prefabProto.Config_STRING,
		Rows: []*prefabProto.ConfigRow{
			{
				ProjectEnvId: Int64Ptr(101),
				Values: []*prefabProto.ConditionalValue{
					{
						Value: internal.CreateConfigValue(1234),
					},
				},
			},
		},
	}

	configs := &prefabProto.Configs{}
	configs.Configs = append(configs.Configs, configFoo, configBar)
	emptyConfigs := &prefabProto.Configs{}

	t.Run("store initialized after set called and has two values", func(t *testing.T) {
		store := BuildApiConfigStore()
		store.SetFromConfigsProto(configs)
		assert.Equal(t, 2, store.Len())
		assert.True(t, store.initialized)

		foo, fooExists := store.GetConfig("foo")

		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFoo, foo)

		bar, barExists := store.GetConfig("bar")
		assert.True(t, barExists)
		assert.NotNil(t, bar)
		assert.Equal(t, configBar, bar)

	})

	t.Run("store initialized with empty configs still marked initialized", func(t *testing.T) {
		store := BuildApiConfigStore()
		store.SetFromConfigsProto(emptyConfigs)
		assert.Equal(t, 0, store.Len())
		assert.True(t, store.initialized)

		foo, fooExists := store.GetConfig("foo")

		assert.False(t, fooExists)
		assert.Nil(t, foo)
	})

	t.Run("updating with tombstoned config foo deletes", func(t *testing.T) {
		store := BuildApiConfigStore()
		store.SetFromConfigsProto(configs)
		assert.Equal(t, 2, store.Len())
		assert.True(t, store.initialized)
		foo, fooExists := store.GetConfig("foo")

		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFoo, foo)

		store.SetFromConfigsProto(&prefabProto.Configs{Configs: []*prefabProto.Config{configFooTombstone}})
		assert.Equal(t, 1, store.Len())

		foo, fooExists = store.GetConfig("foo")

		assert.False(t, fooExists)
		assert.Nil(t, foo)
	})

	t.Run("updating with tombstoned config foo does nothing with smaller id", func(t *testing.T) {
		store := BuildApiConfigStore()
		store.SetFromConfigsProto(configs)
		assert.Equal(t, 2, store.Len())
		assert.True(t, store.initialized)

		foo, fooExists := store.GetConfig("foo")
		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFoo, foo)

		configFooTombstoneWithSmallerId := proto.Clone(configFooTombstone).(*prefabProto.Config)
		configFooTombstoneWithSmallerId.Id = 1
		store.SetFromConfigsProto(&prefabProto.Configs{Configs: []*prefabProto.Config{configFooTombstoneWithSmallerId}})
		assert.Equal(t, 2, store.Len())

		foo, fooExists = store.GetConfig("foo")
		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFoo, foo)
	})

	t.Run("updating with changed config foo does nothing with smaller id", func(t *testing.T) {
		store := BuildApiConfigStore()
		store.SetFromConfigsProto(configs)
		assert.Equal(t, 2, store.Len())
		assert.True(t, store.initialized)

		foo, fooExists := store.GetConfig("foo")
		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFoo, foo)
		configFooWithDifferentValuePlusSmallerId := proto.Clone(configFooWithDifferentValue).(*prefabProto.Config)
		configFooWithDifferentValuePlusSmallerId.Id = 1
		store.SetFromConfigsProto(&prefabProto.Configs{Configs: []*prefabProto.Config{configFooWithDifferentValuePlusSmallerId}})
		assert.Equal(t, 2, store.Len())

		foo, fooExists = store.GetConfig("foo")
		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFoo, foo)
	})

	t.Run("updating with changed config foo updates when id is larger", func(t *testing.T) {
		store := BuildApiConfigStore()
		store.SetFromConfigsProto(configs)
		assert.Equal(t, 2, store.Len())
		assert.True(t, store.initialized)

		foo, fooExists := store.GetConfig("foo")
		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFoo, foo)
		store.SetFromConfigsProto(&prefabProto.Configs{Configs: []*prefabProto.Config{configFooWithDifferentValue}})
		assert.Equal(t, 2, store.Len())

		foo, fooExists = store.GetConfig("foo")
		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFooWithDifferentValue, foo)
	})
}

func TestApiConfigStore(t *testing.T) {
	RunTests(t)
}
