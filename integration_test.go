package prefab

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/prefab-cloud/prefab-cloud-go/anyhelpers"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

type clientOverridesYaml struct {
	InitializationTimeOutSec *float64 `yaml:"initialization_timeout_sec"`
	PrefabApiUrl             *string  `yaml:"prefab_api_url"`
	OnInitFailure            *string  `yaml:"on_init_failure"`
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
	CaseName        string               `yaml:"name"`
	Client          string               `yaml:"client"`
	Function        string               `yaml:"function"`
	Type            *string              `yaml:"type"`
	Input           input                `yaml:"input"`
	Expected        expected             `yaml:"expected"`
	ClientOverrides *clientOverridesYaml `yaml:"client_overrides"`
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

func (suite *GeneratedTestSuite) TestGet() {
	suite.executeGetTest("get.yaml")
}

func (suite *GeneratedTestSuite) TestGetFeatureFlag() {
	suite.executeGetTest("get_feature_flag.yaml")
}

func (suite *GeneratedTestSuite) TestGetWeightedValues() {
	suite.executeGetTest("get_weighted_values.yaml")
}

func (suite *GeneratedTestSuite) TestGetOrRaise() {
	suite.executeGetTest("get_or_raise.yaml")
}

func (suite *GeneratedTestSuite) TestEnabled() {
	suite.enabledTest("enabled.yaml")
}

func (suite *GeneratedTestSuite) TestEnabledWithContexts() {
	suite.enabledTest("enabled_with_contexts.yaml")
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

		if testCase.ClientOverrides != nil {
			if testCase.ClientOverrides.OnInitFailure != nil {
				onInitFailure, onInitFailureMappingErr := mapStringToOnInitializationFailure(*testCase.ClientOverrides.OnInitFailure)
				if onInitFailureMappingErr != nil {
					panic(onInitFailureMappingErr)
				}

				opts.OnInitializationFailure = onInitFailure
			}

			if testCase.ClientOverrides.InitializationTimeOutSec != nil {
				opts.InitializationTimeoutSeconds = *testCase.ClientOverrides.InitializationTimeOutSec
			}

			if testCase.ClientOverrides.PrefabApiUrl != nil {
				opts.ApiUrl = *testCase.ClientOverrides.PrefabApiUrl
			}
		}
	})
	client, err = NewClient(options)

	return client, err
}

type expectedResult struct {
	value any
	err   *string
}

func (suite *GeneratedTestSuite) executeGetTest(filename string) {
	testCases := suite.LoadGetTestCasesFromYAML(filename)
	for _, testCase := range testCases {
		suite.Run(buildTestCaseName(testCase, filename), func() {
			client, err := buildClient(suite.ApiKey, testCase)
			suite.Require().NoError(err, "client constructor failed")

			expectedValue, foundExpectedValue := processExpectedResult(testCase)
			suite.Require().True(foundExpectedValue, "no expected value for test case %s", testCase.CaseName)
			defaultValue, defaultValueExists := getDefaultValue(testCase)
			context, contextErr := processContext(testCase)
			suite.Require().NoError(contextErr, "error building context")

			configKey, configKeyErr := getConfigKeyToUse(testCase)
			suite.Require().NoError(configKeyErr)
			result := suite.makeGetCall(client, testCase.Type, configKey, context, defaultValueExists, defaultValue)
			suite.Require().True(foundExpectedValue, "should have found some expected value or error")

			if expectedValue.value != nil {
				suite.Require().True(result.valueOk, "GetConfigValue should work")
				suite.Require().NoError(result.err, "error looking up key %s", testCase.Input.Key)
				suite.True(cmp.Equal(result.value, expectedValue.value))
			} else if expectedValue.err != nil {
				suite.Require().Error(result.err, "there should be some kind of error")

				switch *expectedValue.err {
				case "unable_to_coerce_env_var":
					suite.Require().ErrorContains(result.err, "type coercion failed")
				case "unable_to_decrypt":
					suite.Require().ErrorContains(result.err, "message authentication failed")
				case "missing_env_var":
					suite.Require().ErrorContains(result.err, "environment variable does not exist")
				case "initialization_timeout":
					suite.Require().ErrorContains(result.err, "initialization timeout")
				case "missing_default":
					suite.Require().ErrorContains(result.err, "config does not exist")
				default:
					suite.Failf("unsupported expected error type", "type was %s", *expectedValue.err)
				}
			} else {
				// expected nil value
				suite.Require().False(result.valueOk, "Expected nil return so the ok return from getter should be false")

				if !result.defaultPresented {
					suite.ErrorContains(result.err, "does not exist", "Error should be present containing `does not exist`")
				}
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

			expectedResult, foundExpectedResult := processExpectedResult(testCase)
			suite.Require().True(foundExpectedResult, "no expected value for test case %s", testCase.CaseName)
			context, contextErr := processContext(testCase)
			suite.Require().NoError(contextErr, "error building context")

			featureIsOn, featureIsOnOk := client.FeatureIsOn(*testCase.Input.Flag, *context)
			suite.Require().True(featureIsOnOk, "FeatureIsOn should work")
			suite.Require().NotNil(expectedResult.value, "expected result's value field should not be nil")

			suite.Require().Equal(expectedResult.value, featureIsOn, "FeatureIsOn should be %v", expectedResult.value)
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

func processExpectedResult(testCase *getTestCase) (expectedResult, bool) {
	if testCase.Expected.Millis != nil {
		return expectedResult{value: time.Duration(*testCase.Expected.Millis) * time.Millisecond}, true
	}

	if testCase.Expected.Value != nil {
		switch val := (*testCase.Expected.Value).(type) {
		case []any:
			if properSlice, isSlice := anyhelpers.CanonicalizeSlice(val); isSlice {
				return expectedResult{value: properSlice}, true
			}

			slog.Warn("slice has mixed type contents, unable to canonicalize")

			return expectedResult{}, false
		case int:
			return expectedResult{value: int64(val)}, true
		default:
			return expectedResult{value: val}, true
		}
	}

	if testCase.Expected.Error != nil {
		return expectedResult{err: testCase.Expected.Error}, true
	}

	return expectedResult{}, true // this being true handles the         expected.value: ~
}

// Mapping function from string to OnInitializationFailure
func mapStringToOnInitializationFailure(s string) (OnInitializationFailure, error) {
	switch s {
	case ":raise":
		return RAISE, nil
	case ":return":
		return UNLOCK, nil
	default:
		return RAISE, errors.New(fmt.Sprintf("unknown OnInitializationFailure value %s", s))
	}
}
