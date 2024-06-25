package contexts

import (
	"strings"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/utils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type NamedContext struct {
	Data map[string]any
	Name string
}

func NewNamedContextWithValues(name string, values map[string]any) *NamedContext {
	return &NamedContext{
		Data: values,
		Name: name,
	}
}

func NewNamedContext() *NamedContext {
	return &NamedContext{
		Data: make(map[string]interface{}),
		Name: "",
	}
}

func NewNamedContextFromProto(ctx *prefabProto.Context) *NamedContext {
	values := make(map[string]interface{})

	for key, value := range ctx.GetValues() {
		nativeValue, _, _ := utils.ExtractValue(value)
		if nativeValue != nil {
			values[key] = nativeValue
		}
	}

	return NewNamedContextWithValues(ctx.GetType(), values)
}

type ContextSet struct {
	Data map[string]*NamedContext
}

func NewContextSet() *ContextSet {
	return &ContextSet{
		Data: make(map[string]*NamedContext),
	}
}

func NewContextSetFromProto(protoContextSet *prefabProto.ContextSet) *ContextSet {
	contextSet := NewContextSet()

	if protoContextSet != nil {
		for _, context := range protoContextSet.GetContexts() {
			namedContext := NewNamedContextFromProto(context)
			contextSet.Data[context.GetType()] = namedContext
		}
	}

	return contextSet
}

func (c *ContextSet) GetContextValue(propertyName string) (any, bool) {
	contextName, key := splitAtFirstDot(propertyName)
	if namedContext, namedContextExists := c.Data[contextName]; namedContextExists {
		value, valueExists := namedContext.Data[key]

		return value, valueExists
	}

	return nil, false // Return nil and false if the named context doesn't exist.
}

func (c *ContextSet) SetNamedContext(newNamedContext *NamedContext) {
	c.Data[newNamedContext.Name] = newNamedContext
}

func (c *ContextSet) WithNamedContext(newNamedContext *NamedContext) *ContextSet {
	c.Data[newNamedContext.Name] = newNamedContext

	return c
}

func (c *ContextSet) WithNamedContextValues(name string, values map[string]interface{}) *ContextSet {
	c.Data[name] = NewNamedContextWithValues(name, values)

	return c
}

func Merge(contextSets ...*ContextSet) *ContextSet {
	newContextSet := NewContextSet()

	for _, contextSet := range contextSets {
		for name, namedContext := range contextSet.Data {
			newContextSet.Data[name] = namedContext
		}
	}

	return newContextSet
}

// splitAtFirstDot splits the input string at the first occurrence of "." and returns
// two strings: the part before the dot and the part after the dot.
// If the string starts with ".", the first return value is "" and the second is the rest of the string.
// If there is no ".", it returns "", and the whole string as the second part.
func splitAtFirstDot(input string) (string, string) {
	parts := strings.SplitN(input, ".", 2)

	if len(parts) == 1 {
		return "", parts[0]
	}

	return parts[0], parts[1]
}
