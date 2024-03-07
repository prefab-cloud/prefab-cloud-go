package mocks

import (
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/stretchr/testify/mock"
)

type MockContextGetter struct {
	mock.Mock
}

func (m *MockContextGetter) GetValue(propertyName string) (value interface{}, valueExists bool) {
	args := m.Called(propertyName)
	if args.Get(0) != nil {
		value = args.Get(0).(interface{})
	} else {
		value = nil
	}
	valueExists = args.Bool(1)
	return value, valueExists
}

func NewMockContext(contextPropertyName string, contextValue interface{}, contextExistsValue bool) (mockedContext *MockContextGetter) {
	mockContext := MockContextGetter{}
	if contextExistsValue {
		// When context exists, return the provided contextValue and true.
		mockContext.On("GetValue", contextPropertyName).Return(contextValue, contextExistsValue)
	} else {
		// When context doesn't exist, ensure the mock returns nil for the contextValue (or a suitable zero value) and false.
		// This simulates the absence of the value correctly.
		mockContext.On("GetValue", contextPropertyName).Return(nil, false)
	}
	return &mockContext
}

type ContextMocking struct {
	ContextPropertyName string
	Value               interface{}
	Exists              bool
}

func NewMockContextWithMultipleValues(mockings []ContextMocking) (mockedContext *MockContextGetter) {
	mockContext := &MockContextGetter{}

	for _, mocking := range mockings {
		contextValue := mocking.Value
		contextExistsValue := mocking.Exists

		if contextExistsValue {
			mockContext.On("GetValue", mocking.ContextPropertyName).Return(contextValue, true)
		} else {
			mockContext.On("GetValue", mocking.ContextPropertyName).Return(nil, false)
		}
	}
	return mockContext
}

type MockConfigStoreGetter struct {
	mock.Mock
}

func (m *MockConfigStoreGetter) GetConfig(key string) (config *prefabProto.Config, configExists bool) {
	args := m.Called(key)
	if args.Get(0) != nil {
		config = args.Get(0).(*prefabProto.Config)
	} else {
		config = nil
	}
	configExists = args.Bool(1)
	return config, configExists
}

type ConfigMockingArgs struct {
	ConfigKey    string
	Config       *prefabProto.Config
	ConfigExists bool
}

func NewMockConfigStoreGetter(args []ConfigMockingArgs) *MockConfigStoreGetter {
	mockConfigStoreGetter := MockConfigStoreGetter{}
	for _, mockingArg := range args {
		mockConfigStoreGetter.On("GetConfig", mockingArg.ConfigKey).Return(mockingArg.Config, mockingArg.ConfigExists)
	}
	return &mockConfigStoreGetter
}

type MockEnvLookup struct {
	mock.Mock
}

func (m *MockEnvLookup) LookupEnv(key string) (string, bool) {
	args := m.Called(key)
	return args.String(0), args.Bool(1)
}

type MockEnvLookupConfig struct {
	Name        string
	Value       string
	ValueExists bool
}

func NewMockEnvLookup(args []MockEnvLookupConfig) *MockEnvLookup {
	mockInstance := MockEnvLookup{}
	for _, mockingArg := range args {
		mockInstance.On("LookupEnv", mockingArg.Name).Return(mockingArg.Value, mockingArg.ValueExists)
	}
	return &mockInstance
}
