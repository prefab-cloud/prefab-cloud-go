package prefab

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type clientOverrides struct {
	OnNoDefault              *interface{} `yaml:"on_no_default"`
	InitializationTimeOutSec *float32     `yaml:"initialization_timeout_sec"`
	PrefabApiUrl             *string      `yaml:"prefab_api_url"`
	OnInitFailure            *string      `yaml:"on_init_failure"`
}
type input struct {
	Key string `yaml:"key"`
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
	// Add more type mappings as needed
}

func (suite *GeneratedTestSuite) TestGet() {

	testCases := suite.LoadGetTestCasesFromYAML("get.yaml")
	fmt.Printf("test cases are %v", testCases)

	for _, testCase := range testCases {
		suite.Run(testCase.Name, func() {
			options := NewOptions(func(opts *Options) {
				opts.PrefabApiUrl = "https://api.staging-prefab.cloud"
				opts.ApiKey = suite.ApiKey
				//TODO respect any client overrides
			})
			client, err := NewClient(options)
			suite.Require().NoError(err, "client did not initialize")
			suite.Require().NotNil(testCase.Type, "testcase.Type should not be nil")
			fmt.Printf("Test case type %s", *testCase.Type)
			fn, ok := typeMap[*testCase.Type]
			if !ok {
				suite.Require().Fail("unsupported type", "Type was %s", *testCase.Type)
			}
			suite.Require().NotNil(fn, "fn should not be nil")

			method := reflect.ValueOf(fn)
			result := method.Call([]reflect.Value{
				reflect.ValueOf(client),
				reflect.ValueOf(testCase.Input.Key),
				reflect.ValueOf(*NewContextSet()),
			})

			value := result[0].Interface()
			err, ok = result[1].Interface().(error)

			if !ok && result[1].Interface() != nil {
				suite.Failf("unexpected error type", "Error type is %T", result[1].Interface())
			}
			suite.Require().NoError(err, "error looking up key %s", testCase.Input.Key)

			if value != nil {
				fmt.Printf("returned value is %v of type %T\n", value, value)
			}
			//TODO test the answer is what we expect. probably use reflect.DeepEquals?

		})

	}

}
