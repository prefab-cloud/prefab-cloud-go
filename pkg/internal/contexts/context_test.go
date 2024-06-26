package contexts_test

import (
	"testing"

	"github.com/mohae/deepcopy"

	"github.com/stretchr/testify/suite"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/testutils"
	"github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ContextTestSuite struct {
	suite.Suite
	contextSet *contexts.ContextSet
}

func (suite *ContextTestSuite) SetupSuite() {
	suite.contextSet = contexts.NewContextSet()
	suite.contextSet.SetNamedContext(contexts.NewNamedContextWithValues("user", map[string]interface{}{
		"key":   "u123",
		"email": "me@example.com",
		"admin": true,
		"age":   int64(42),
	}))
	suite.contextSet.SetNamedContext(contexts.NewNamedContextWithValues("team", map[string]interface{}{
		"key":  "t123",
		"name": "dev ops",
	}))
	suite.contextSet.SetNamedContext(contexts.NewNamedContextWithValues("", map[string]interface{}{
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
					Type: internal.StringPtr("user"),
					Values: map[string]*proto.ConfigValue{
						"key":   testutils.CreateConfigValueAndAssertOk(suite.T(), "u123"),
						"email": testutils.CreateConfigValueAndAssertOk(suite.T(), "me@example.com"),
						"admin": testutils.CreateConfigValueAndAssertOk(suite.T(), true),
						"age":   testutils.CreateConfigValueAndAssertOk(suite.T(), 42),
					},
				},
				{
					Type: internal.StringPtr("team"),
					Values: map[string]*proto.ConfigValue{
						"key":  testutils.CreateConfigValueAndAssertOk(suite.T(), "t123"),
						"name": testutils.CreateConfigValueAndAssertOk(suite.T(), "dev ops"),
					},
				},
				{
					Type: internal.StringPtr(""),
					Values: map[string]*proto.ConfigValue{
						"key": testutils.CreateConfigValueAndAssertOk(suite.T(), "?234"),
						"id":  testutils.CreateConfigValueAndAssertOk(suite.T(), 3456),
					},
				},
			},
		}
		contextSet := contexts.NewContextSetFromProto(pContextSet)

		val, valExists := contextSet.GetContextValue("user.key")
		suite.True(valExists)
		suite.Equal("u123", val)
		suite.EqualValues(suite.contextSet, contextSet)
	})
}

func (suite *ContextTestSuite) TestModification() {
	suite.Run("adding a new context replaces existing", func() {
		contextSet := deepcopy.Copy(*suite.contextSet).(contexts.ContextSet)
		contextSet.SetNamedContext(contexts.NewNamedContextWithValues("team", map[string]interface{}{
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
