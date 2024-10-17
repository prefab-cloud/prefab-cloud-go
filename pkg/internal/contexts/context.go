package contexts

import (
	"sort"
	"strings"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/utils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type NamedContext struct {
	Data map[string]any
	Name string
}

func (nc *NamedContext) ToProto() *prefabProto.Context {
	protoContext := &prefabProto.Context{
		Type:   &nc.Name,
		Values: make(map[string]*prefabProto.ConfigValue),
	}

	for key, value := range nc.Data {
		protoValue, ok := utils.Create(value)

		if ok {
			protoContext.Values[key] = protoValue
		}
	}

	return protoContext
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

func (cs *ContextSet) GroupedKey() string {
	anyKeys := false
	ids := []string{}

	for _, context := range cs.Data {
		if context.Data["key"] != nil {
			anyKeys = true

			ids = append(ids, context.Name+":"+context.Data["key"].(string))
		} else {
			ids = append(ids, context.Name+":")
		}
	}

	if !anyKeys {
		return ""
	}

	sort.Strings(ids)

	return strings.Join(ids, "|")
}

func (cs *ContextSet) ToProto() *prefabProto.ContextSet {
	protoContextSet := &prefabProto.ContextSet{}

	for _, namedContext := range cs.Data {
		protoContextSet.Contexts = append(protoContextSet.Contexts, namedContext.ToProto())
	}

	return protoContextSet
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

func (cs *ContextSet) GetContextValue(propertyName string) (any, bool) {
	contextName, key := splitAtFirstDot(propertyName)
	if namedContext, namedContextExists := cs.Data[contextName]; namedContextExists {
		value, valueExists := namedContext.Data[key]

		return value, valueExists
	}

	return nil, false // Return nil and false if the named context doesn't exist.
}

func (cs *ContextSet) SetNamedContext(newNamedContext *NamedContext) {
	cs.Data[newNamedContext.Name] = newNamedContext
}

func (cs *ContextSet) WithNamedContext(newNamedContext *NamedContext) *ContextSet {
	cs.Data[newNamedContext.Name] = newNamedContext

	return cs
}

func (cs *ContextSet) WithNamedContextValues(name string, values map[string]interface{}) *ContextSet {
	cs.Data[name] = NewNamedContextWithValues(name, values)

	return cs
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
