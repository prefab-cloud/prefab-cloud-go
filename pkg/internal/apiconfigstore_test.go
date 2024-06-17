package internal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/testutils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

func TestApiConfigStore(t *testing.T) {
	configFoo := &prefabProto.Config{
		Key:        "foo",
		Id:         10,
		ProjectId:  100,
		ConfigType: prefabProto.ConfigType_CONFIG,
		ValueType:  prefabProto.Config_STRING,
		Rows: []*prefabProto.ConfigRow{
			{
				ProjectEnvId: internal.Int64Ptr(101),
				Values: []*prefabProto.ConditionalValue{
					{
						Value: testutils.CreateConfigValueAndAssertOk(t, "foo-value"),
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
				ProjectEnvId: internal.Int64Ptr(101),
				Values: []*prefabProto.ConditionalValue{
					{
						Value: testutils.CreateConfigValueAndAssertOk(t, "foo-value-two"),
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
				ProjectEnvId: internal.Int64Ptr(101),
				Values: []*prefabProto.ConditionalValue{
					{
						Value: testutils.CreateConfigValueAndAssertOk(t, 1234),
					},
				},
			},
		},
	}

	configs := &prefabProto.Configs{}
	configs.Configs = append(configs.Configs, configFoo, configBar)
	emptyConfigs := &prefabProto.Configs{}

	t.Run("store initialized after set called and has two values", func(t *testing.T) {
		store := internal.BuildAPIConfigStore()
		store.SetFromConfigsProto(configs)
		assert.Equal(t, 2, store.Len())
		assert.True(t, store.Initialized)

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
		store := internal.BuildAPIConfigStore()
		store.SetFromConfigsProto(emptyConfigs)
		assert.Equal(t, 0, store.Len())
		assert.True(t, store.Initialized)

		foo, fooExists := store.GetConfig("foo")

		assert.False(t, fooExists)
		assert.Nil(t, foo)
	})

	t.Run("updating with tombstoned config foo deletes", func(t *testing.T) {
		store := internal.BuildAPIConfigStore()
		store.SetFromConfigsProto(configs)
		assert.Equal(t, 2, store.Len())
		assert.True(t, store.Initialized)
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
		store := internal.BuildAPIConfigStore()
		store.SetFromConfigsProto(configs)
		assert.Equal(t, 2, store.Len())
		assert.True(t, store.Initialized)

		foo, fooExists := store.GetConfig("foo")
		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFoo, foo)

		configFooTombstoneWithSmallerID := proto.Clone(configFooTombstone).(*prefabProto.Config)
		configFooTombstoneWithSmallerID.Id = 1
		store.SetFromConfigsProto(&prefabProto.Configs{Configs: []*prefabProto.Config{configFooTombstoneWithSmallerID}})
		assert.Equal(t, 2, store.Len())

		foo, fooExists = store.GetConfig("foo")
		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFoo, foo)
	})

	t.Run("updating with changed config foo does nothing with smaller id", func(t *testing.T) {
		store := internal.BuildAPIConfigStore()
		store.SetFromConfigsProto(configs)
		assert.Equal(t, 2, store.Len())
		assert.True(t, store.Initialized)

		foo, fooExists := store.GetConfig("foo")
		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFoo, foo)

		configFooWithDifferentValuePlusSmallerID := proto.Clone(configFooWithDifferentValue).(*prefabProto.Config)
		configFooWithDifferentValuePlusSmallerID.Id = 1
		store.SetFromConfigsProto(&prefabProto.Configs{Configs: []*prefabProto.Config{configFooWithDifferentValuePlusSmallerID}})
		assert.Equal(t, 2, store.Len())

		foo, fooExists = store.GetConfig("foo")
		assert.True(t, fooExists)
		assert.NotNil(t, foo)
		assert.Equal(t, configFoo, foo)
	})

	t.Run("updating with changed config foo updates when id is larger", func(t *testing.T) {
		store := internal.BuildAPIConfigStore()
		store.SetFromConfigsProto(configs)
		assert.Equal(t, 2, store.Len())
		assert.True(t, store.Initialized)

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
