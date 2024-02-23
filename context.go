package prefab

import (
	"github.com/prefab-cloud/prefab-cloud-go/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"strings"
)

type NamedContext struct {
	Name string
	Data map[string]interface{}
}

func NewNamedContextWithValues(name string, values map[string]interface{}) *NamedContext {
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
	for key, value := range ctx.Values {
		nativeValue, _ := internal.UnpackConfigValue(value)
		values[key] = nativeValue
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
	for _, context := range protoContextSet.Contexts {
		namedContext := NewNamedContextFromProto(context)
		contextSet.Data[context.GetType()] = namedContext
	}
	return contextSet
}

func (c *ContextSet) getValue(propertyName string) (interface{}, bool) {
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

// splitAtFirstDot splits the input string at the first occurrence of "." and returns
// two strings: the part before the dot and the part after the dot.
// If the string starts with ".", the first return value is "" and the second is the rest of the string.
// If there is no ".", it returns "", and the whole string as the second part.
func splitAtFirstDot(input string) (string, string) {
	// Special case for strings starting with "."
	if strings.HasPrefix(input, ".") {
		return "", input[1:]
	}

	index := strings.Index(input, ".")
	if index == -1 {
		// No dot found, return an empty string and the whole string
		return "", input
	}
	// Split the string into before and after the dot
	beforeDot := input[:index]
	afterDot := input[index+1:]
	return beforeDot, afterDot
}
