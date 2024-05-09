package prefab

import (
	"testing"

	"github.com/mohae/deepcopy"

	"github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/stretchr/testify/suite"
)

func stringPtr(val string) *string {
	return &val
}

type ContextTestSuite struct {
	suite.Suite
	contextSet *ContextSet
}

func (suite *ContextTestSuite) SetupSuite() {
	suite.contextSet = NewContextSet()
	suite.contextSet.SetNamedContext(NewNamedContextWithValues("user", map[string]interface{}{
		"key":   "u123",
		"email": "me@example.com",
		"admin": true,
		"age":   int64(42),
	}))
	suite.contextSet.SetNamedContext(NewNamedContextWithValues("team", map[string]interface{}{
		"key":  "t123",
		"name": "dev ops",
	}))
	suite.contextSet.SetNamedContext(NewNamedContextWithValues("", map[string]interface{}{
		"key": "?234",
		"id":  int64(3456),
	}))
}

func (suite *ContextTestSuite) TestContextReads() {
	suite.Run("reading un-prefixed property hits the unnamed context", func() {
		value, valueExists := suite.contextSet.GetContextValue("key")
		suite.Assert().EqualValues(true, valueExists)
		suite.Assert().EqualValues("?234", value)

		value, valueExists = suite.contextSet.GetContextValue("id")
		suite.Assert().EqualValues(true, valueExists)
		suite.Assert().EqualValues(3456, value)
	})

	suite.Run("reading un-prefixed property hits the unnamed context", func() {
		value, valueExists := suite.contextSet.GetContextValue("key")
		suite.Assert().EqualValues(true, valueExists)
		suite.Assert().EqualValues("?234", value)
	})

	suite.Run("reading non existing key in unnamed context works", func() {
		value, valueExists := suite.contextSet.GetContextValue(".foobar")
		suite.Assert().EqualValues(false, valueExists)
		suite.Assert().EqualValues(nil, value)
	})

	suite.Run("reading name-prefixed property hits the appropriate context", func() {
		value, valueExists := suite.contextSet.GetContextValue("user.key")
		suite.Assert().EqualValues(true, valueExists)
		suite.Assert().EqualValues("u123", value)

		value, valueExists = suite.contextSet.GetContextValue("team.name")
		suite.Assert().EqualValues(true, valueExists)
		suite.Assert().EqualValues("dev ops", value)
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
		contextSet := NewContextSetFromProto(pContextSet)

		val, valExists := contextSet.GetContextValue("user.key")
		suite.Assert().Equal(valExists, true)
		suite.Assert().Equal(val, "u123")
		suite.Assert().EqualValues(suite.contextSet, contextSet)
	})
}

func (suite *ContextTestSuite) TestModification() {
	suite.Run("adding a new context replaces existing", func() {
		contextSet := deepcopy.Copy(*suite.contextSet).(ContextSet)
		contextSet.SetNamedContext(NewNamedContextWithValues("team", map[string]interface{}{
			"foo": "bar",
		}))

		value, valueExists := contextSet.GetContextValue("team.foo")
		suite.Assert().True(valueExists)
		suite.Assert().EqualValues(value, "bar")

		value, valueExists = contextSet.GetContextValue("team.name")
		suite.Assert().False(valueExists)
		suite.Assert().Empty(value)
	})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestContextTestSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}
