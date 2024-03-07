package prefab

import (
	"errors"
	"github.com/prefab-cloud/prefab-cloud-go/internal"
	"github.com/prefab-cloud/prefab-cloud-go/mocks"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockDecrypter struct {
	mock.Mock
}

func (m *mockDecrypter) DecryptValue(secretKey string, value string) (decryptedValue string, err error) {
	args := m.Called(secretKey, value)
	return args.String(0), args.Error(1)
}

type mockDecrypterArgs struct {
	key            string
	encryptedValue string
	decryptedValue string
	err            error
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

func (m *mockWeightedValueResolver) Resolve(weightedValues *prefabProto.WeightedValues, propertyName string, contextGetter ContextGetter) (valueResult *prefabProto.ConfigValue, index int) {
	args := m.Called(weightedValues, propertyName, contextGetter)
	return args.Get(0).(*prefabProto.ConfigValue), args.Int(1)
}

type mockWeightedValueResolverArgs struct {
	weightedValues *prefabProto.WeightedValues
	propertyName   string
	returnValue    *prefabProto.ConfigValue
	index          int
}

func newMockWeightedValueResolver(args []mockWeightedValueResolverArgs) (mockedWeightedValueResolver *mockWeightedValueResolver) {
	mockResolver := &mockWeightedValueResolver{}
	for _, currArg := range args {
		mockResolver.On("Resolve", currArg.weightedValues, currArg.propertyName, mock.Anything).Return(currArg.returnValue, currArg.index)
	}
	return mockResolver
}

type mockConfigEvaluator struct {
	mock.Mock
}

func (m *mockConfigEvaluator) EvaluateConfig(config *prefabProto.Config, contextSet ContextGetter) (match ConditionMatch) {
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
		Value:  internal.CreateConfigValue(1),
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
		}}
	configValueOne := internal.CreateConfigValue("one")

	tests := []struct {
		name                          string
		configKey                     string
		wantConfigMatch               ConfigMatch
		expectError                   bool //TODO make more specific
		mockDecrypterArgs             []mockDecrypterArgs
		mockWeightedValueResolverArgs []mockWeightedValueResolverArgs
		mockConfigEvaluatorArgs       []mockConfigEvaluatorArgs
		mockConfigStoreArgs           []mocks.ConfigMockingArgs
		mockEnvLookup                 *mocks.MockEnvLookup
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
				match:                 internal.CreateConfigValue(providedEnvVarValue),
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
			mockEnvLookup: mocks.NewMockEnvLookup([]mocks.MockEnvLookupConfig{
				{Name: providedEnvVarName, Value: providedEnvVarValue, ValueExists: true},
			}),
		},
		{
			name:        "config has provided but env var does not exist",
			configKey:   theKey,
			expectError: true,
			wantConfigMatch: ConfigMatch{
				match:                 internal.CreateConfigValue(providedEnvVarValue),
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
			mockEnvLookup: mocks.NewMockEnvLookup([]mocks.MockEnvLookupConfig{
				{Name: providedEnvVarName, Value: providedEnvVarValue, ValueExists: false},
			}),
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
						match:                 weightedValuesConfigValue, // points at "decrypt.with.me"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDecrypter := newMockDecrypter(tt.mockDecrypterArgs)
			defer mockDecrypter.AssertExpectations(t)

			mockWeightedValueResolver := newMockWeightedValueResolver(tt.mockWeightedValueResolverArgs)
			defer mockWeightedValueResolver.AssertExpectations(t)

			mockConfigEvaluator := newMockConfigEvaluator(tt.mockConfigEvaluatorArgs)
			defer mockConfigEvaluator.AssertExpectations(t)

			mockConfigStoreGetter := mocks.NewMockConfigStoreGetter(tt.mockConfigStoreArgs)
			defer mockConfigStoreGetter.AssertExpectations(t)

			mockContextGetter := new(mocks.MockContextGetter)
			defer mockContextGetter.AssertExpectations(t)
			if tt.mockEnvLookup != nil {
				defer tt.mockEnvLookup.AssertExpectations(t)
			}

			c := &ConfigResolver{
				configStore:           mockConfigStoreGetter,
				ruleEvaluator:         mockConfigEvaluator,
				weightedValueResolver: mockWeightedValueResolver,
				decrypter:             mockDecrypter,
				envLookup:             tt.mockEnvLookup,
			}
			match, err := c.ResolveValue(tt.configKey, mockContextGetter)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.Equalf(t, tt.wantConfigMatch, match, "ResolveValue(%v, %v)", tt.configKey, mockContextGetter)
			}

		})
	}
}
