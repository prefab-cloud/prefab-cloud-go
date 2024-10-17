package integrationtestsupport

import (
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
)

type ClientOverridesYaml struct {
	InitializationTimeOutSec *float64 `yaml:"initialization_timeout_sec"`
	PrefabAPIURL             *string  `yaml:"prefab_api_url"`
	OnInitFailure            *string  `yaml:"on_init_failure"`
}

type input struct {
	Key     *string `yaml:"key"`
	Flag    *string `yaml:"flag"`
	Default *string `yaml:"default"`
}

type expected struct {
	Status *string      `yaml:"status"`
	Value  *interface{} `yaml:"value"`
	Error  *string      `yaml:"error"`
	Millis *int64       `yaml:"millis"`
}

type GetTest struct {
	Name  string            `yaml:"name"`
	Cases []GetTestCaseYaml `yaml:"cases"`
}

type GetTestCaseYaml struct {
	Expected        expected               `yaml:"expected"`
	Input           input                  `yaml:"input"`
	Type            *string                `yaml:"type"`
	ClientOverrides *ClientOverridesYaml   `yaml:"client_overrides"`
	RawContexts     map[string]TestContext `yaml:"contexts"`
	CaseName        string                 `yaml:"name"`
	Client          string                 `yaml:"client"`
	Function        string                 `yaml:"function"`
}

type TestContext = map[string]map[string]interface{}

type TestCaseContexts struct {
	Global *contexts.ContextSet
	Local  *contexts.ContextSet
	Block  *contexts.ContextSet
}

type GetTestCase struct {
	Contexts TestCaseContexts
	GetTestCaseYaml
	TestName string
}

func (t GetTestCase) GetGlobalContexts() *contexts.ContextSet {
	return t.Contexts.Global
}

func (t GetTestCase) GetBlockContexts() *contexts.ContextSet {
	return t.Contexts.Block
}

func (t GetTestCase) GetClientOverrides() *ClientOverridesYaml {
	return t.ClientOverrides
}

func ctxDataToContextSet(data map[string]interface{}) *contexts.ContextSet {
	contextSet := contexts.NewContextSet()

	for key, value := range data {
		contextSet.WithNamedContextValues(key, value.(map[string]any))
	}

	return contextSet
}
