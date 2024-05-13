package prefab

import (
	"errors"
	"fmt"
	"github.com/prefab-cloud/prefab-cloud-go/anyhelpers"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

type clientOverrides struct {
	OnNoDefault              *interface{} `yaml:"on_no_default"`
	InitializationTimeOutSec *float32     `yaml:"initialization_timeout_sec"`
	PrefabApiUrl             *string      `yaml:"prefab_api_url"`
	OnInitFailure            *string      `yaml:"on_init_failure"`
}
type input struct {
	Key     *string                            `yaml:"key"`
	Flag    *string                            `yaml:"flag"`
	Default *string                            `yaml:"default"`
	Context *map[string]map[string]interface{} `yaml:"context"`
}

type expected struct {
	Status *string      `yaml:"status"`
	Value  *interface{} `yaml:"value"`
	Error  *string      `yaml:"error"`
	Millis *int64       `yaml:"millis"`
}

type getTest struct {
	Name    *string                            `yaml:"name"`
	Context *map[string]map[string]interface{} `yaml:"context"`
	Cases   []getTestCaseYaml                  `yaml:"cases"`
}

type getTestCaseYaml struct {
	CaseName        string           `yaml:"name"`
	Client          string           `yaml:"client"`
	Function        string           `yaml:"function"`
	Type            *string          `yaml:"type"`
	Input           input            `yaml:"input"`
	Expected        expected         `yaml:"expected"`
	ClientOverrides *clientOverrides `yaml:"client_overrides"`
}

type getTestCase struct {
	getTestCaseYaml
	TestName    *string
	TestContext *map[string]map[string]interface{}
}

type GeneratedTestSuite struct {
	suite.Suite
	BaseDirectory string
	ApiKey        string
}

func TestGeneratedSuite(t *testing.T) {
	suite.Run(t, new(GeneratedTestSuite))
}

func (suite *GeneratedTestSuite) SetupSuite() {
	suite.BaseDirectory = "./testdata/prefab-cloud-integration-test-data/tests/current"
	suite.ApiKey = os.Getenv("PREFAB_INTEGRATION_TEST_API_KEY")
	suite.Require().NotEmpty(suite.ApiKey, "No API key found in environment var PREFAB_INTEGRATION_TEST_API_KEY")
}

func (suite *GeneratedTestSuite) LoadGetTestCasesFromYAML(filename string) []*getTestCase {
	fileContents, err := os.ReadFile(filepath.Join(suite.BaseDirectory, filename))
	suite.NoError(err)

	var data map[string]interface{}

	err = yaml.Unmarshal(fileContents, &data)
	suite.NoError(err)

	testsData, ok := data["tests"].([]interface{})
	suite.True(ok, "Failed to find 'tests' array in YAML file")

	casesBytes, err := yaml.Marshal(testsData)
	suite.NoError(err)

	var tests []*getTest
	err = yaml.Unmarshal(casesBytes, &tests)
	suite.NoError(err)

	var testCases []*getTestCase

	for _, test := range tests {
		for _, testCaseYaml := range test.Cases {
			testCase := &getTestCase{getTestCaseYaml: testCaseYaml, TestName: test.Name, TestContext: test.Context}
			testCases = append(testCases, testCase)
		}
	}

	return testCases
}

var typeMap = map[string]interface{}{
	"INT":         (*Client).GetIntValue,
	"BOOL":        (*Client).GetBoolValue,
	"DOUBLE":      (*Client).GetFloatValue,
	"STRING":      (*Client).GetStringValue,
	"STRING_LIST": (*Client).GetStringSliceValue,
	"DURATION":    (*Client).GetDurationValue,
	// Add more type mappings as needed
}

var typeWithDefaultsMap = map[string]interface{}{
	"INT":         (*Client).GetIntValueWithDefault,
	"BOOL":        (*Client).GetBoolValueWithDefault,
	"DOUBLE":      (*Client).GetFloatValueWithDefault,
	"STRING":      (*Client).GetStringValueWithDefault,
	"STRING_LIST": (*Client).GetStringSliceValueWithDefault,
	// Add more type mappings as needed
}

type configLookupResult struct {
	defaultPresented bool
	defaultValue     any
	value            any
	valueOk          bool
	err              error // only present when default not presented
}

func (suite *GeneratedTestSuite) makeGetCall(client *Client, dataType *string, key string, contextSet *ContextSet, hasDefault bool, defaultValue any) configLookupResult {
	var returnOfGetCall []reflect.Value

	suite.Require().NotNil(dataType, "dataType (from testcase.Type) should not be nil. Fix the data")

	if hasDefault {
		fn, ok := typeWithDefaultsMap[*dataType]
		if !ok {
			suite.Require().Fail("unsupported type for case with default value", "Type was %s", dataType)
		}
		method := reflect.ValueOf(fn)
		returnOfGetCall = method.Call([]reflect.Value{
			reflect.ValueOf(client),
			reflect.ValueOf(key),
			reflect.ValueOf(*contextSet),
			reflect.ValueOf(defaultValue),
		})

	} else {
		fn, ok := typeMap[*dataType]
		if !ok {
			suite.Require().Fail("unsupported type", "Type was %s", dataType)
		}
		method := reflect.ValueOf(fn)
		returnOfGetCall = method.Call([]reflect.Value{
			reflect.ValueOf(client),
			reflect.ValueOf(key),
			reflect.ValueOf(*contextSet),
		})
	}

	ok, okOk := returnOfGetCall[1].Interface().(bool)
	suite.Require().True(okOk, "Expected fetch of ok value to work, check reflection code in test")
	result := configLookupResult{defaultPresented: hasDefault, defaultValue: defaultValue, value: returnOfGetCall[0].Interface(), valueOk: ok}

	if !hasDefault {
		if len(returnOfGetCall) >= 3 {
			returnedValue := returnOfGetCall[2].Interface()
			if returnedValue == nil {
			} else {
				var errOk bool
				result.err, errOk = returnedValue.(error)
				suite.Require().True(errOk, fmt.Sprintf("Expected third return value to be of type error, but got: %T", returnedValue))
			}
		} else {
			suite.Require().Fail(fmt.Sprintf("Expected at least three return values from the function, but got: %d", len(returnOfGetCall)))
		}
	}
	return result
}

func buildClient(apiKey string, testCase *getTestCase) (client *Client, err error) {
	options := NewOptions(func(opts *Options) {
		opts.ApiUrl = "https://api.staging-prefab.cloud"
		opts.ApiKey = apiKey
		// TODO respect anyhelpers client overrides
	})
	client, err = NewClient(options)
	return client, err
}

func (suite *GeneratedTestSuite) TestGet() {
	suite.executeGetTest("get.yaml")
}

func (suite *GeneratedTestSuite) TestGetFeatureFlag() {
	suite.executeGetTest("get_feature_flag.yaml")
}

func (suite *GeneratedTestSuite) TestGetWeightedValues() {
	suite.executeGetTest("get_weighted_values.yaml")
}

func (suite *GeneratedTestSuite) executeGetTest(filename string) {

	testCases := suite.LoadGetTestCasesFromYAML(filename)
	for _, testCase := range testCases {
		suite.Run(buildTestCaseName(testCase, filename), func() {
			client, err := buildClient(suite.ApiKey, testCase)
			suite.Require().NoError(err, "client constructor failed")
			expectedValue, foundExpectedValue := processExpectedValue(testCase)
			suite.Require().True(foundExpectedValue, "no expected value for test case %s", testCase.CaseName)
			defaultValue, defaultValueExists := getDefaultValue(testCase)
			context, contextErr := processContext(testCase)
			suite.Require().NoError(contextErr, "error building context")
			configKey, configKeyErr := getConfigKeyToUse(testCase)
			suite.Require().NoError(configKeyErr)
			result := suite.makeGetCall(client, testCase.Type, configKey, context, defaultValueExists, defaultValue)
			if expectedValue == nil {
				suite.Require().False(result.valueOk, "Expected nil return so the ok return from getter should be false")
				if !result.defaultPresented {
					suite.ErrorContains(result.err, "does not exist", "Error should be present containing `does not exist`")

				}
			} else {
				suite.Require().True(result.valueOk, "GetConfigValue should work")
				suite.Require().NoError(result.err, "error looking up key %s", testCase.Input.Key)
				suite.True(cmp.Equal(result.value, expectedValue))
			}
		})
	}
}

func (suite *GeneratedTestSuite) enabledTest(filename string) {
	testCases := suite.LoadGetTestCasesFromYAML(filename)
	for _, testCase := range testCases {
		suite.Run(buildTestCaseName(testCase, filename), func() {
			client, err := buildClient(suite.ApiKey, testCase)
			suite.Require().NoError(err, "client constructor failed")
			expectedValue, foundExpectedValue := processExpectedValue(testCase)
			suite.Require().True(foundExpectedValue, "no expected value for test case %s", testCase.CaseName)
			context, contextErr := processContext(testCase)
			suite.Require().NoError(contextErr, "error building context")
			featureIsOn, featureIsOnOk := client.FeatureIsOn(*testCase.Input.Flag, *context)
			suite.Require().True(featureIsOnOk, "FeatureIsOn should work")
			suite.Require().Equal(expectedValue, featureIsOn, "FeatureIsOn should be %v", expectedValue)
		})
	}
}

func buildTestCaseName(testCase *getTestCase, filename string) string {
	testName := ""
	if testCase.TestName != nil {
		testName = *testCase.TestName
	}
	return fmt.Sprintf("%s::%s::%s", filename, testName, testCase.CaseName)
}

func (suite *GeneratedTestSuite) TestEnabled() {
	suite.enabledTest("enabled.yaml")
}

func (suite *GeneratedTestSuite) TestEnabledWithContexts() {
	suite.enabledTest("enabled_with_contexts.yaml")
}

func getDefaultValue(testCase *getTestCase) (interface{}, bool) {
	if testCase.Input.Default == nil {
		return nil, false
	}
	return *testCase.Input.Default, true
}

func processContext(testCase *getTestCase) (*ContextSet, error) {
	if testCase.TestContext == nil && testCase.Input.Context == nil {
		return &ContextSet{}, nil
	}
	// handle context stacking for tests here because the client doesn't yet support it
	contextSet := NewContextSet()
	if testCase.TestContext != nil {
		for key, value := range *testCase.TestContext {
			contextSet.SetNamedContext(NewNamedContextWithValues(key, value))
		}
	}
	if testCase.TestContext != nil {
		for key, value := range *testCase.TestContext {
			contextSet.SetNamedContext(NewNamedContextWithValues(key, value))
		}
	}
	if testCase.Input.Context != nil {
		for key, value := range *testCase.Input.Context {
			contextSet.SetNamedContext(NewNamedContextWithValues(key, value))
		}
	}
	return contextSet, nil
}

func getConfigKeyToUse(testCase *getTestCase) (string, error) {
	if testCase.Input.Key != nil {
		return *testCase.Input.Key, nil
	}
	if testCase.Input.Flag != nil {
		return *testCase.Input.Flag, nil
	}
	return "", errors.New("no key or flag in testCase.Input")

}

func processExpectedValue(testCase *getTestCase) (interface{}, bool) {
	if testCase.Expected.Millis != nil {
		return time.Duration(*testCase.Expected.Millis) * time.Millisecond, true
	}
	if testCase.Expected.Value != nil {
		switch val := (*testCase.Expected.Value).(type) {
		case []any:
			if properSlice, isSlice := anyhelpers.CanonicalizeSlice(val); isSlice {
				return properSlice, true
			}
			slog.Warn("slice has mixed type contents, unable to canonicalize")
			return nil, false
		case int:
			return int64(val), true
		default:
			return val, true
		}
	} else {
		return nil, true
	}
}
