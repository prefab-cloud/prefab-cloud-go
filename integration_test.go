package prefab

import (
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

type getTestCase struct {
	Name            string           `yaml:"name"`
	Client          string           `yaml:"client"`
	Function        string           `yaml:"function"`
	Type            *string          `yaml:"type"`
	Input           input            `yaml:"input"`
	Expected        expected         `yaml:"expected"`
	ClientOverrides *clientOverrides `yaml:"client_overrides"`
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
	suite.True(ok, "Failed to parse 'tests' array from YAML")
	suite.Len(testsData, 1, "Expected only one child under 'tests'")

	testData, ok := testsData[0].(map[string]interface{})
	suite.True(ok, "Failed to parse test data as map")

	casesData, ok := testData["cases"].([]interface{})
	suite.True(ok, "Failed to parse 'cases' array from test data")

	casesBytes, err := yaml.Marshal(casesData)
	suite.NoError(err)

	var testCases []*getTestCase
	err = yaml.Unmarshal(casesBytes, &testCases)
	suite.NoError(err)

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

func (suite *GeneratedTestSuite) makeCall(client *Client, dataType string, key string, contextSet *ContextSet, hasDefault bool, defaultValue any) configLookupResult {
	var returnOfGetCall []reflect.Value

	if hasDefault {
		fn, ok := typeWithDefaultsMap[dataType]
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
		fn, ok := typeMap[dataType]
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

func (suite *GeneratedTestSuite) TestGet() {
	testCases := suite.LoadGetTestCasesFromYAML("get.yaml")
	for _, testCase := range testCases {
		suite.Run(testCase.Name, func() {
			options := NewOptions(func(opts *Options) {
				opts.ApiUrl = "https://api.staging-prefab.cloud"
				opts.ApiKey = suite.ApiKey
				// TODO respect anyhelpers client overrides
			})
			client, err := NewClient(options)
			suite.Require().NoError(err, "client constructor failed")
			suite.Require().NotNil(testCase.Type, "testcase.Type should not be nil. Fix the data")
			expectedValue, foundExpectedValue := processExpectedValue(testCase)
			suite.Require().True(foundExpectedValue, "no expected value for test case %s", testCase.Name)
			defaultValue, defaultValueExists := getDefaultValue(testCase)
			result := suite.makeCall(client, *testCase.Type, *testCase.Input.Key, NewContextSet(), defaultValueExists, defaultValue)

			if expectedValue == nil {
				suite.Require().False(result.valueOk, "Expected nil return so the ok return from getter should be false")
				if !result.defaultPresented {
					suite.ErrorContains(result.err, "does not exist", "Error should be present containing `does not exist`")

				}
			} else {
				suite.Require().True(result.valueOk, "GetContextValue should work")
				suite.Require().NoError(result.err, "error looking up key %s", testCase.Input.Key)
				suite.True(cmp.Equal(result.value, expectedValue))
			}

		})
	}
}

/*func (suite *GeneratedTestSuite) TestGetFeatureFlag() {
	testCases := suite.LoadGetTestCasesFromYAML("get_feature_flag.yaml")
	for _, testCase := range testCases {
		suite.Run(testCase.Name, func() {
			options := NewOptions(func(opts *Options) {
				opts.PrefabApiUrl = "https://api.staging-prefab.cloud"
				opts.ApiKey = suite.ApiKey
				// TODO respect anyhelpers client overrides
			})
			client, err := NewClient(options)
			suite.Require().NoError(err, "client constructor failed")
			suite.Require().NotNil(testCase.Type, "testcase.Type should not be nil. Fix the data")
			expectedValue, foundExpectedValue := processExpectedValue(testCase)
			suite.Require().True(foundExpectedValue, "no expected value for test case %s", testCase.Name)
			defaultValue, defaultValueExists := getDefaultValue(testCase)
			result := suite.makeCall(client, *testCase.Type, *testCase.Input.Flag, NewContextSet(), defaultValueExists, defaultValue)

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
}*/

func getDefaultValue(testCase *getTestCase) (interface{}, bool) {
	if testCase.Input.Default == nil {
		return nil, false
	}
	return *testCase.Input.Default, true
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
