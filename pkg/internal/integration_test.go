package internal_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	prefab "github.com/prefab-cloud/prefab-cloud-go/pkg"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/anyhelpers"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/options"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

type clientOverridesYaml struct {
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

type getTest struct {
	Name  *string           `yaml:"name"`
	Cases []getTestCaseYaml `yaml:"cases"`
}

type getTestCaseYaml struct {
	Expected        expected               `yaml:"expected"`
	Input           input                  `yaml:"input"`
	Type            *string                `yaml:"type"`
	ClientOverrides *clientOverridesYaml   `yaml:"client_overrides"`
	RawContexts     map[string]TestContext `yaml:"contexts"`
	CaseName        string                 `yaml:"name"`
	Client          string                 `yaml:"client"`
	Function        string                 `yaml:"function"`
}

type TestContext = map[string]map[string]interface{}

type TestCaseContexts struct {
	global *contexts.ContextSet
	local  *contexts.ContextSet
	block  *contexts.ContextSet
}

type getTestCase struct {
	Contexts TestCaseContexts
	TestName *string
	getTestCaseYaml
}

type GeneratedTestSuite struct {
	suite.Suite
	BaseDirectory string
	APIKey        string
}

func TestGeneratedSuite(t *testing.T) {
	suite.Run(t, new(GeneratedTestSuite))
}

func (suite *GeneratedTestSuite) SetupSuite() {
	if os.Getenv("FAST") == "true" {
		suite.T().Skip("Skipping integration tests because FAST=true")
	}

	suite.BaseDirectory = "./testdata/prefab-cloud-integration-test-data/tests/current"
	suite.APIKey = os.Getenv("PREFAB_INTEGRATION_TEST_API_KEY")
	suite.Require().NotEmpty(suite.APIKey, "No API key found in environment var PREFAB_INTEGRATION_TEST_API_KEY")
}

func (suite *GeneratedTestSuite) loadGetTestCasesFromYAML(filename string) []*getTestCase {
	fileContents, err := os.ReadFile(filepath.Join(suite.BaseDirectory, filename))
	suite.Require().NoError(err)

	var data map[string]interface{}

	err = yaml.Unmarshal(fileContents, &data)
	suite.Require().NoError(err)

	testsData, ok := data["tests"].([]interface{})
	suite.True(ok, "Failed to find 'tests' array in YAML file")

	casesBytes, err := yaml.Marshal(testsData)
	suite.Require().NoError(err)

	var tests []*getTest
	err = yaml.Unmarshal(casesBytes, &tests)
	suite.Require().NoError(err)

	var testCases []*getTestCase

	for _, test := range tests {
		for _, testCaseYaml := range test.Cases {
			testCase := &getTestCase{getTestCaseYaml: testCaseYaml, TestName: test.Name}

			testCase.Contexts = TestCaseContexts{
				global: testContextToContextSet(testCase.RawContexts["global"]),
				local:  testContextToContextSet(testCase.RawContexts["local"]),
				block:  testContextToContextSet(testCase.RawContexts["block"]),
			}

			testCases = append(testCases, testCase)
		}
	}

	return testCases
}

func testContextToContextSet(rawContext TestContext) *contexts.ContextSet {
	contextSet := contexts.NewContextSet()

	for key, value := range rawContext {
		contextSet.SetNamedContext(contexts.NewNamedContextWithValues(key, value))
	}

	return contextSet
}

func functionFromTypeString(typeString string) (interface{}, bool) {
	switch typeString {
	case "INT":
		return prefab.ClientInterface.GetIntValue, true
	case "BOOL":
		return prefab.ClientInterface.GetBoolValue, true
	case "DOUBLE":
		return prefab.ClientInterface.GetFloatValue, true
	case "STRING":
		return prefab.ClientInterface.GetStringValue, true
	case "STRING_LIST":
		return prefab.ClientInterface.GetStringSliceValue, true
	case "DURATION":
		return prefab.ClientInterface.GetDurationValue, true
	case "JSON":
		return prefab.ClientInterface.GetJSONValue, true
	}

	slog.Error("unsupported type", "type", typeString)

	return nil, false
}

func functionWithDefaultFromTypeString(typeString string) (interface{}, bool) {
	switch typeString {
	case "INT":
		return prefab.ClientInterface.GetIntValueWithDefault, true
	case "BOOL":
		return prefab.ClientInterface.GetBoolValueWithDefault, true
	case "DOUBLE":
		return prefab.ClientInterface.GetFloatValueWithDefault, true
	case "STRING":
		return prefab.ClientInterface.GetStringValueWithDefault, true
	case "STRING_LIST":
		return prefab.ClientInterface.GetStringSliceValueWithDefault, true
	}

	slog.Error("unsupported type", "type", typeString)

	return nil, false
}

type configLookupResult struct {
	defaultValue     any
	value            any
	err              error
	defaultPresented bool
	valueOk          bool
}

func runSuiteFromFile(suite *GeneratedTestSuite, filename string) func(string) {
	switch filename {
	case "get.yaml", "get_feature_flag.yaml", "get_weighted_values.yaml", "get_or_raise.yaml", "context_precedence.yaml", "enabled.yaml", "enabled_with_contexts.yaml":
		return suite.executeGetOrEnabledTest
	case "get_log_level.yaml":
		return suite.executeLogLevelTest
	case "post.yaml":
		suite.T().Skip("Telemetry integration tests aren't implemented yet")

		return nil
	default:
		suite.T().Fatal("Unsupported test file: ", filename)

		return nil
	}
}

func (suite *GeneratedTestSuite) TestAll() {
	files, err := os.ReadDir(suite.BaseDirectory)
	suite.Require().NoError(err)

	for _, file := range files {
		testFunction := runSuiteFromFile(suite, file.Name())

		if testFunction != nil {
			testFunction(file.Name())
		}
	}
}

func (suite *GeneratedTestSuite) makeGetCall(client prefab.ClientInterface, dataType *string, key string, contextSet *contexts.ContextSet, hasDefault bool, defaultValue any) configLookupResult {
	var returnOfGetCall []reflect.Value

	suite.Require().NotNil(dataType, "dataType (from testcase.Type) should not be nil. Fix the data")

	if hasDefault {
		fn, ok := functionWithDefaultFromTypeString(*dataType)
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
		fn, ok := functionFromTypeString(*dataType)
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

	if hasDefault {
		return result
	}

	if len(returnOfGetCall) >= 3 {
		returnedValue := returnOfGetCall[2].Interface()
		if returnedValue != nil {
			var errOk bool
			result.err, errOk = returnedValue.(error)
			suite.Require().True(errOk, fmt.Sprintf("Expected third return value to be of type error, but got: %T", returnedValue))
		}
	} else {
		suite.Require().Fail(fmt.Sprintf("Expected at least three return values from the function, but got: %d", len(returnOfGetCall)))
	}

	return result
}

func buildClient(suite *GeneratedTestSuite, testCase *getTestCase) (*prefab.ContextBoundClient, error) {
	apiKey := suite.APIKey
	// 50/50 setting the API URLs and using the env var override
	useEnvVarOverride := rand.Intn(2) == 1
	url := "https://api.staging-prefab.cloud"

	// default to using the hardcoded URL override
	options := []prefab.Option{
		prefab.WithAPIURLs([]string{url}),
		prefab.WithAPIKey(apiKey),
		prefab.WithGlobalContext(testCase.Contexts.global),
	}

	// if we flip the coin and get 1, then we'll use the env var override
	if useEnvVarOverride {
		suite.T().Setenv("PREFAB_API_URL_OVERRIDE", url)

		options = []prefab.Option{
			// this URL won't work, but that's great because it is overridden by the env var
			prefab.WithAPIURLs([]string{"https://localhost/this-will-not-work"}),
			prefab.WithAPIKey(apiKey),
			prefab.WithGlobalContext(testCase.Contexts.global),
		}
	}

	if testCase.ClientOverrides != nil {
		options = applyOverrides(testCase, options)
	}

	client, err := prefab.NewClient(options...)
	if err != nil {
		return nil, err
	}

	return client.WithContext(testCase.Contexts.block), nil
}

func applyOverrides(testCase *getTestCase, options []prefab.Option) []prefab.Option {
	if testCase.ClientOverrides.OnInitFailure != nil {
		onInitFailure, onInitFailureMappingErr := mapStringToOnInitializationFailure(*testCase.ClientOverrides.OnInitFailure)
		if onInitFailureMappingErr != nil {
			panic(onInitFailureMappingErr)
		}

		options = append(options, prefab.WithOnInitializationFailure(onInitFailure))
	}

	if testCase.ClientOverrides.InitializationTimeOutSec != nil {
		options = append(options, prefab.WithInitializationTimeoutSeconds(*testCase.ClientOverrides.InitializationTimeOutSec))
	}

	if testCase.ClientOverrides.PrefabAPIURL != nil {
		options = append(options, prefab.WithAPIURLs([]string{*testCase.ClientOverrides.PrefabAPIURL}))
	}

	return options
}

type expectedResult struct {
	value any
	err   *string
}

func enabledTest(suite *GeneratedTestSuite, testCase *getTestCase, client prefab.ClientInterface) {
	expected, ok := processExpectedResult(testCase)
	suite.Require().True(ok, "no expected value for test case %s", testCase.CaseName)

	featureIsOn, featureIsOnOk := client.FeatureIsOn(*testCase.Input.Flag, *testCase.Contexts.local)
	suite.Require().True(featureIsOnOk, "FeatureIsOn should work")
	suite.Require().NotNil(expected.value, "expected result's value field should not be nil")

	suite.Require().Equal(expected.value, featureIsOn, "FeatureIsOn should be %v", expected.value)
}

func (suite *GeneratedTestSuite) executeLogLevelTest(filename string) {
	testCases := suite.loadGetTestCasesFromYAML(filename)
	for _, testCase := range testCases {
		suite.Run(buildTestCaseName(testCase, filename), func() {
			client, err := buildClient(suite, testCase)
			suite.Require().NoError(err, "client constructor failed")

			expectedValue, foundExpectedValue := processExpectedResult(testCase)
			suite.Require().True(foundExpectedValue, "no expected value for test case %s", testCase.CaseName)

			configKey, configKeyErr := getConfigKeyToUse(testCase)
			suite.Require().NoError(configKeyErr)

			actualValue, found, err := client.GetLogLevelStringValue(configKey, *testCase.Contexts.local)

			suite.Require().NoError(err)
			suite.True(found, "expected log level to be found")

			suite.Require().Equal(expectedValue.value, actualValue, "LogLevel should be %v", expectedValue.value)
		})
	}
}

func (suite *GeneratedTestSuite) executeGetOrEnabledTest(filename string) {
	testCases := suite.loadGetTestCasesFromYAML(filename)
	for _, testCase := range testCases {
		suite.Run(buildTestCaseName(testCase, filename), func() {
			client, err := buildClient(suite, testCase)
			suite.Require().NoError(err, "client constructor failed")

			expectedValue, foundExpectedValue := processExpectedResult(testCase)
			suite.Require().True(foundExpectedValue, "no expected value for test case %s", testCase.CaseName)
			defaultValue, defaultValueExists := getDefaultValue(testCase)

			configKey, configKeyErr := getConfigKeyToUse(testCase)
			suite.Require().NoError(configKeyErr)

			if testCase.Function == "enabled" {
				enabledTest(suite, testCase, client)

				return
			}

			result := suite.makeGetCall(client, testCase.Type, configKey, testCase.Contexts.local, defaultValueExists, defaultValue)

			suite.Require().True(foundExpectedValue, "should have found some expected value or error")

			switch {
			case expectedValue.value != nil:
				suite.Require().True(result.valueOk, "GetConfigValue should work")
				suite.Require().NoError(result.err, "error looking up key %s", testCase.Input.Key)

				assertSameValues(suite, expectedValue, result)
			case expectedValue.err != nil:
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
			default:
				// expected nil value
				suite.Require().False(result.valueOk, "Expected nil return so the ok return from getter should be false")

				if !result.defaultPresented {
					suite.ErrorContains(result.err, "does not exist", "Error should be present containing `does not exist`")
				}
			}
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
func mapStringToOnInitializationFailure(str string) (options.OnInitializationFailure, error) {
	switch str {
	case ":raise":
		return prefab.ReturnError, nil
	case ":return":
		return prefab.ReturnNilMatch, nil
	default:
		return prefab.ReturnError, fmt.Errorf("unknown OnInitializationFailure value %s", str)
	}
}

func assertSameValues(suite *GeneratedTestSuite, expectedValue expectedResult, result configLookupResult) {
	switch expectedValue.value.(type) {
	case map[string]interface{}:
		resultJSON, err := json.Marshal(result.value)
		if err != nil {
			suite.Failf("unable to marshal result to JSON", "error was %s", err)
		}

		expectedJSON, err := json.Marshal(expectedValue.value)
		if err != nil {
			suite.Failf("unable to marshal expected value to JSON", "error was %s", err)
		}

		suite.JSONEq(string(expectedJSON), string(resultJSON), "expected value should match")
	default:
		suite.Equal(expectedValue.value, result.value)
	}
}
