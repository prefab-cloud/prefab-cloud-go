package internal_test

import (
	"testing"

	"github.com/mohae/deepcopy"

	"github.com/stretchr/testify/suite"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/proto"
)

func stringPtr(val string) *string {
	return &val
}

type ContextTestSuite struct {
	suite.Suite
	contextSet *internal.ContextSet
}

func (suite *ContextTestSuite) SetupSuite() {
	suite.contextSet = internal.NewContextSet()
	suite.contextSet.SetNamedContext(internal.NewNamedContextWithValues("user", map[string]interface{}{
		"key":   "u123",
		"email": "me@example.com",
		"admin": true,
		"age":   int64(42),
	}))
	suite.contextSet.SetNamedContext(internal.NewNamedContextWithValues("team", map[string]interface{}{
		"key":  "t123",
		"name": "dev ops",
	}))
	suite.contextSet.SetNamedContext(internal.NewNamedContextWithValues("", map[string]interface{}{
		"key": "?234",
		"id":  int64(3456),
	}))
}

func (suite *ContextTestSuite) TestContextReads() {
	suite.Run("reading un-prefixed property hits the unnamed context", func() {
		value, valueExists := suite.contextSet.GetContextValue("key")
		suite.True(valueExists)
		suite.EqualValues("?234", value)

		value, valueExists = suite.contextSet.GetContextValue("id")
		suite.True(valueExists)
		suite.EqualValues(3456, value)
	})

	suite.Run("reading un-prefixed property hits the unnamed context", func() {
		value, valueExists := suite.contextSet.GetContextValue("key")
		suite.True(valueExists)
		suite.EqualValues("?234", value)
	})

	suite.Run("reading non existing key in unnamed context works", func() {
		value, valueExists := suite.contextSet.GetContextValue(".foobar")
		suite.False(valueExists)
		suite.Nil(value)
	})

	suite.Run("reading name-prefixed property hits the appropriate context", func() {
		value, valueExists := suite.contextSet.GetContextValue("user.key")
		suite.True(valueExists)
		suite.EqualValues("u123", value)

		value, valueExists = suite.contextSet.GetContextValue("team.name")
		suite.True(valueExists)
		suite.EqualValues("dev ops", value)
	})
}

func (suite *ContextTestSuite) TestContextConversionFromProto() {
	suite.Run("initializing from a proto works", func() {
		pContextSet := &proto.ContextSet{
			Contexts: []*proto.Context{
				{
					Type: stringPtr("user"),
					Values: map[string]*proto.ConfigValue{
						"key":   createConfigValueAndAssertOk("u123", suite.T()),
						"email": createConfigValueAndAssertOk("me@example.com", suite.T()),
						"admin": createConfigValueAndAssertOk(true, suite.T()),
						"age":   createConfigValueAndAssertOk(42, suite.T()),
					},
				},
				{
					Type: stringPtr("team"),
					Values: map[string]*proto.ConfigValue{
						"key":  createConfigValueAndAssertOk("t123", suite.T()),
						"name": createConfigValueAndAssertOk("dev ops", suite.T()),
					},
				},
				{
					Type: stringPtr(""),
					Values: map[string]*proto.ConfigValue{
						"key": createConfigValueAndAssertOk("?234", suite.T()),
						"id":  createConfigValueAndAssertOk(3456, suite.T()),
					},
				},
			},
		}
		contextSet := internal.NewContextSetFromProto(pContextSet)

		val, valExists := contextSet.GetContextValue("user.key")
		suite.True(valExists)
		suite.Equal("u123", val)
		suite.EqualValues(suite.contextSet, contextSet)
	})
}

func (suite *ContextTestSuite) TestModification() {
	suite.Run("adding a new context replaces existing", func() {
		contextSet := deepcopy.Copy(*suite.contextSet).(internal.ContextSet)
		contextSet.SetNamedContext(internal.NewNamedContextWithValues("team", map[string]interface{}{
			"foo": "bar",
		}))

		value, valueExists := contextSet.GetContextValue("team.foo")
		suite.True(valueExists)
		suite.EqualValues("bar", value)

		value, valueExists = contextSet.GetContextValue("team.name")
		suite.False(valueExists)
		suite.Empty(value)
	})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestContextTestSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}
