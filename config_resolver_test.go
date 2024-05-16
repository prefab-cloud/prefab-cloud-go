package prefab

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/prefab-cloud/prefab-cloud-go/mocks"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type mockDecrypter struct {
	mock.Mock
}

func (m *mockDecrypter) DecryptValue(secretKey string, value string) (string, error) {
	args := m.Called(secretKey, value)
	return args.String(0), args.Error(1)
}

type mockDecrypterArgs struct {
	err            error
	key            string
	encryptedValue string
	decryptedValue string
}

func newMockDecrypter(args []mockDecrypterArgs) *mockDecrypter {
	mockedDecrypter := &mockDecrypter{}
	for _, currArg := range args {
		mockedDecrypter.On("DecryptValue", currArg.key, currArg.encryptedValue).Return(currArg.decryptedValue, currArg.err)
	}

	return mockedDecrypter
}

type mockWeightedValueResolver struct {
	mock.Mock
}

func (m *mockWeightedValueResolver) Resolve(weightedValues *prefabProto.WeightedValues, propertyName string, contextGetter ContextValueGetter) (*prefabProto.ConfigValue, int) {
	args := m.Called(weightedValues, propertyName, contextGetter)
	return args.Get(0).(*prefabProto.ConfigValue), args.Int(1)
}

type mockWeightedValueResolverArgs struct {
	weightedValues *prefabProto.WeightedValues
	returnValue    *prefabProto.ConfigValue
	propertyName   string
	index          int
}

func newMockWeightedValueResolver(args []mockWeightedValueResolverArgs) *mockWeightedValueResolver {
	mockResolver := &mockWeightedValueResolver{}
	for _, currArg := range args {
		mockResolver.On("Resolve", currArg.weightedValues, currArg.propertyName, mock.Anything).Return(currArg.returnValue, currArg.index)
	}

	return mockResolver
}

type mockConfigEvaluator struct {
	mock.Mock
}

func (m *mockConfigEvaluator) EvaluateConfig(config *prefabProto.Config, contextSet ContextValueGetter) ConditionMatch {
	args := m.Called(config, contextSet)
	return args.Get(0).(ConditionMatch)
}

type mockConfigEvaluatorArgs struct {
	config *prefabProto.Config
	match  ConditionMatch
}

func newMockConfigEvaluator(args []mockConfigEvaluatorArgs) *mockConfigEvaluator {
	mockInstance := &mockConfigEvaluator{}
	for _, currArg := range args {
		mockInstance.On("EvaluateConfig", currArg.config, mock.Anything).Return(currArg.match)
	}

	return mockInstance
}

func TestConfigResolver_ResolveValue(t *testing.T) {
	theKey := "the.key"
	emptyConfigInstance := &prefabProto.Config{Key: theKey}
	emptyConfigInstance2 := &prefabProto.Config{}
	decryptWithConfigKey := "decrypt.with.me"
	decryptWithSecretKey := "the-secret-key"
	decryptedValue := "the-decrypted-value"
	encryptedValue := "the-encrypted-value"
	providedEnvVarName := "SOME_ENV"
	providedEnvVarValue := "THE_VALUE"
	envVarSource := prefabProto.ProvidedSource_ENV_VAR
	providedConfigValue := &prefabProto.ConfigValue{
		Type: &prefabProto.ConfigValue_Provided{Provided: &prefabProto.Provided{Lookup: &providedEnvVarName, Source: &envVarSource}},
	}
	decryptWithConfigValue := &prefabProto.ConfigValue{
		Type:         &prefabProto.ConfigValue_String_{String_: encryptedValue},
		DecryptWith:  stringPtr(decryptWithConfigKey),
		Confidential: boolPtr(true),
	}
	decryptionKeyConfigValue := &prefabProto.ConfigValue{
		Type: &prefabProto.ConfigValue_String_{String_: decryptWithSecretKey},
	}
	decryptedConfigValue := &prefabProto.ConfigValue{
		Type:         &prefabProto.ConfigValue_String_{String_: decryptedValue},
		Confidential: boolPtr(true),
	}

	weightedValueOne := &prefabProto.WeightedValue{
		Weight: 100,
		Value:  createConfigValueAndAssertOk(1, t),
	}
	weightedValues := &prefabProto.WeightedValues{
		HashByPropertyName: stringPtr("some.property"),
		WeightedValues: []*prefabProto.WeightedValue{
			weightedValueOne,
		},
	}

	weightedValuesConfigValue := &prefabProto.ConfigValue{
		Type: &prefabProto.ConfigValue_WeightedValues{
			weightedValues,
		},
	}
	configValueOne := createConfigValueAndAssertOk("one", t)

	type keyValuePair struct {
		name  string
		value string
	}

	tests := []struct {
		name                          string
		configKey                     string
		mockDecrypterArgs             []mockDecrypterArgs
		mockWeightedValueResolverArgs []mockWeightedValueResolverArgs
		mockConfigEvaluatorArgs       []mockConfigEvaluatorArgs
		mockConfigStoreArgs           []mocks.ConfigMockingArgs
		envVarsToSet                  []keyValuePair
		wantConfigMatch               ConfigMatch
		expectError                   bool
	}{
		{
			name:      "standard pass through",
			configKey: theKey,
			wantConfigMatch: ConfigMatch{
				match:                 configValueOne,
				isMatch:               true,
				originalKey:           theKey,
				originalMatch:         configValueOne,
				conditionalValueIndex: 1,
				rowIndex:              1,
			},
			mockConfigStoreArgs: []mocks.ConfigMockingArgs{
				{
					ConfigKey:    theKey,
					Config:       emptyConfigInstance,
					ConfigExists: true,
				},
			},
			mockConfigEvaluatorArgs: []mockConfigEvaluatorArgs{
				{
					config: emptyConfigInstance,
					match: ConditionMatch{
						isMatch:               true,
						match:                 configValueOne,
						rowIndex:              1,
						conditionalValueIndex: 1,
					},
				},
			},
		},
		{
			name:        "config does not exist",
			expectError: true,
			configKey:   theKey,
			wantConfigMatch: ConfigMatch{
				match:       nil,
				isMatch:     false,
				originalKey: theKey,
			},
			mockConfigStoreArgs: []mocks.ConfigMockingArgs{
				{
					ConfigKey:    theKey,
					Config:       nil,
					ConfigExists: false,
				},
			},
		},
		{
			name:      "config has provided set",
			configKey: theKey,
			wantConfigMatch: ConfigMatch{
				match:                 createConfigValueAndAssertOk(providedEnvVarValue, t),
				isMatch:               true,
				originalKey:           theKey,
				originalMatch:         providedConfigValue,
				conditionalValueIndex: 1,
				rowIndex:              1,
			},
			mockConfigStoreArgs: []mocks.ConfigMockingArgs{
				{
					ConfigKey:    theKey,
					Config:       emptyConfigInstance,
					ConfigExists: true,
				},
			},
			mockConfigEvaluatorArgs: []mockConfigEvaluatorArgs{
				{
					config: emptyConfigInstance,
					match: ConditionMatch{
						isMatch:               true,
						match:                 providedConfigValue,
						rowIndex:              1,
						conditionalValueIndex: 1,
					},
				},
			},
			envVarsToSet: []keyValuePair{{providedEnvVarName, providedEnvVarValue}},
		},
		{
			name:        "config has provided but env var does not exist",
			configKey:   theKey,
			expectError: true,
			wantConfigMatch: ConfigMatch{
				match:                 createConfigValueAndAssertOk(providedEnvVarValue, t),
				isMatch:               true,
				originalKey:           theKey,
				originalMatch:         providedConfigValue,
				conditionalValueIndex: 1,
				rowIndex:              1,
			},
			mockConfigStoreArgs: []mocks.ConfigMockingArgs{
				{
					ConfigKey:    theKey,
					Config:       emptyConfigInstance,
					ConfigExists: true,
				},
			},
			mockConfigEvaluatorArgs: []mockConfigEvaluatorArgs{
				{
					config: emptyConfigInstance,
					match: ConditionMatch{
						isMatch:               true,
						match:                 providedConfigValue,
						rowIndex:              1,
						conditionalValueIndex: 1,
					},
				},
			},
		},
		{
			name:      "config has decrypt with and it works", // need to resolve two configs, the main one and the one with the key
			configKey: theKey,
			wantConfigMatch: ConfigMatch{
				match:                 decryptedConfigValue,
				isMatch:               true,
				originalKey:           theKey,
				originalMatch:         decryptWithConfigValue,
				conditionalValueIndex: 1,
				rowIndex:              1,
			},
			mockConfigStoreArgs: []mocks.ConfigMockingArgs{
				{
					ConfigKey:    theKey,
					Config:       emptyConfigInstance,
					ConfigExists: true,
				},
				{
					ConfigKey:    decryptWithConfigKey,
					Config:       emptyConfigInstance2,
					ConfigExists: true,
				},
			},
			mockConfigEvaluatorArgs: []mockConfigEvaluatorArgs{
				{
					config: emptyConfigInstance,
					match: ConditionMatch{
						isMatch:               true,
						match:                 decryptWithConfigValue, // points at "decrypt.with.me"
						rowIndex:              1,
						conditionalValueIndex: 1,
					},
				},
				{
					config: emptyConfigInstance2,
					match: ConditionMatch{
						isMatch:               true,
						match:                 decryptionKeyConfigValue,
						rowIndex:              1,
						conditionalValueIndex: 0,
					},
				},
			},
			mockDecrypterArgs: []mockDecrypterArgs{{encryptedValue: encryptedValue, decryptedValue: decryptedValue, key: decryptWithSecretKey}},
		},
		{
			name:        "config has decrypt with and it fails", // need to resolve two configs, the main one and the one with the key
			configKey:   theKey,
			expectError: true,
			wantConfigMatch: ConfigMatch{
				match:                 decryptedConfigValue,
				isMatch:               true,
				originalKey:           theKey,
				originalMatch:         decryptWithConfigValue,
				conditionalValueIndex: 1,
				rowIndex:              1,
			},
			mockConfigStoreArgs: []mocks.ConfigMockingArgs{
				{
					ConfigKey:    theKey,
					Config:       emptyConfigInstance,
					ConfigExists: true,
				},
				{
					ConfigKey:    decryptWithConfigKey,
					Config:       emptyConfigInstance2,
					ConfigExists: true,
				},
			},
			mockConfigEvaluatorArgs: []mockConfigEvaluatorArgs{
				{
					config: emptyConfigInstance,
					match: ConditionMatch{
						isMatch:               true,
						match:                 decryptWithConfigValue, // points at "decrypt.with.me"
						rowIndex:              1,
						conditionalValueIndex: 1,
					},
				},
				{
					config: emptyConfigInstance2,
					match: ConditionMatch{
						isMatch:               true,
						match:                 decryptionKeyConfigValue,
						rowIndex:              1,
						conditionalValueIndex: 0,
					},
				},
			},
			mockDecrypterArgs: []mockDecrypterArgs{{encryptedValue: encryptedValue, decryptedValue: decryptedValue, key: decryptWithSecretKey, err: errors.New("decryption went poorly")}},
		},
		{
			name:        "config has decrypt with but config containing key does not exist", // need to resolve two configs, the main one and the one with the key
			configKey:   theKey,
			expectError: true,
			wantConfigMatch: ConfigMatch{
				match:                 decryptedConfigValue,
				isMatch:               true,
				originalKey:           theKey,
				originalMatch:         decryptWithConfigValue,
				conditionalValueIndex: 1,
				rowIndex:              1,
			},
			mockConfigStoreArgs: []mocks.ConfigMockingArgs{
				{
					ConfigKey:    theKey,
					Config:       emptyConfigInstance,
					ConfigExists: true,
				},
				{
					ConfigKey:    decryptWithConfigKey,
					Config:       emptyConfigInstance2,
					ConfigExists: false,
				},
			},
			mockConfigEvaluatorArgs: []mockConfigEvaluatorArgs{
				{
					config: emptyConfigInstance,
					match: ConditionMatch{
						isMatch:               true,
						match:                 decryptWithConfigValue, // points at "decrypt.with.me"
						rowIndex:              1,
						conditionalValueIndex: 1,
					},
				},
			},
			mockDecrypterArgs: []mockDecrypterArgs{},
		},
		{
			name:      "config has weighted values, succeeds", // need to resolve two configs, the main one and the one with the key
			configKey: theKey,
			wantConfigMatch: ConfigMatch{
				match:                 weightedValueOne.Value,
				isMatch:               true,
				originalKey:           theKey,
				originalMatch:         weightedValuesConfigValue,
				conditionalValueIndex: 1,
				rowIndex:              1,
				weightedValueIndex:    intPtr(2),
			},
			mockConfigStoreArgs: []mocks.ConfigMockingArgs{
				{
					ConfigKey:    theKey,
					Config:       emptyConfigInstance,
					ConfigExists: true,
				},
			},
			mockConfigEvaluatorArgs: []mockConfigEvaluatorArgs{
				{
					config: emptyConfigInstance,
					match: ConditionMatch{
						isMatch:               true,
						match:                 weightedValuesConfigValue, // points at "decrypt.with.me"
						rowIndex:              1,
						conditionalValueIndex: 1,
					},
				},
			},
			mockWeightedValueResolverArgs: []mockWeightedValueResolverArgs{{returnValue: weightedValueOne.GetValue(), weightedValues: weightedValues, propertyName: theKey, index: 2}},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockDecrypter := newMockDecrypter(testCase.mockDecrypterArgs)
			defer mockDecrypter.AssertExpectations(t)

			mockWeightedValueResolver := newMockWeightedValueResolver(testCase.mockWeightedValueResolverArgs)
			defer mockWeightedValueResolver.AssertExpectations(t)

			mockConfigEvaluator := newMockConfigEvaluator(testCase.mockConfigEvaluatorArgs)
			defer mockConfigEvaluator.AssertExpectations(t)

			mockConfigStoreGetter := mocks.NewMockConfigStoreGetter(testCase.mockConfigStoreArgs)
			defer mockConfigStoreGetter.AssertExpectations(t)

			mockContextGetter := new(mocks.MockContextGetter)
			defer mockContextGetter.AssertExpectations(t)

			for _, pair := range testCase.envVarsToSet {
				t.Setenv(pair.name, pair.value)
			}

			resolver := &ConfigResolver{
				configStore:           mockConfigStoreGetter,
				ruleEvaluator:         mockConfigEvaluator,
				weightedValueResolver: mockWeightedValueResolver,
				decrypter:             mockDecrypter,
			}

			match, err := resolver.ResolveValue(testCase.configKey, mockContextGetter)
			if testCase.expectError {
				assert.Error(t, err)
			} else {
				assert.Equalf(t, testCase.wantConfigMatch, match, "ResolveValue(%v, %v)", testCase.configKey, mockContextGetter)
			}
		})
	}
}
