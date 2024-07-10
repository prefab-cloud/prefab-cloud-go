package mocks

import (
	"github.com/stretchr/testify/mock"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type MockContextGetter struct {
	mock.Mock
}

func (m *MockContextGetter) GetContextValue(propertyName string) (interface{}, bool) {
	var value interface{}

	args := m.Called(propertyName)
	if args.Get(0) != nil {
		value = args.Get(0)
	} else {
		value = nil
	}

	valueExists := args.Bool(1)

	return value, valueExists
}

func NewMockContext(contextPropertyName string, contextValue interface{}, contextExistsValue bool) *MockContextGetter {
	mockContext := MockContextGetter{}
	if contextExistsValue {
		// When context exists, return the provided contextValue and true.
		mockContext.On("GetContextValue", contextPropertyName).Return(contextValue, contextExistsValue)
	} else {
		// When context doesn't exist, ensure the mock returns nil for the contextValue (or a suitable zero value) and false.
		// This simulates the absence of the value correctly.
		mockContext.On("GetContextValue", contextPropertyName).Return(nil, false)
	}

	return &mockContext
}

type ContextMocking struct {
	Value               interface{}
	ContextPropertyName string
	Exists              bool
}

func NewMockContextWithMultipleValues(mockings []ContextMocking) *MockContextGetter {
	mockContext := &MockContextGetter{}

	for _, mocking := range mockings {
		contextValue := mocking.Value
		contextExistsValue := mocking.Exists

		if contextExistsValue {
			mockContext.On("GetContextValue", mocking.ContextPropertyName).Return(contextValue, true)
		} else {
			mockContext.On("GetContextValue", mocking.ContextPropertyName).Return(nil, false)
		}
	}

	return mockContext
}

type MockConfigStoreGetter struct {
	mock.Mock
}

func (m *MockConfigStoreGetter) GetConfig(key string) (*prefabProto.Config, bool) {
	var config *prefabProto.Config

	args := m.Called(key)
	if args.Get(0) != nil {
		config = args.Get(0).(*prefabProto.Config)
	} else {
		config = nil
	}

	configExists := args.Bool(1)

	return config, configExists
}

type ConfigMockingArgs struct {
	Config       *prefabProto.Config
	ConfigKey    string
	ConfigExists bool
}

func NewMockConfigStoreGetter(args []ConfigMockingArgs) *MockConfigStoreGetter {
	mockConfigStoreGetter := MockConfigStoreGetter{}
	for _, mockingArg := range args {
		mockConfigStoreGetter.On("GetConfig", mockingArg.ConfigKey).Return(mockingArg.Config, mockingArg.ConfigExists)
	}

	return &mockConfigStoreGetter
}

type MockProjectEnvIDSupplier struct {
	mock.Mock
}

func (m *MockProjectEnvIDSupplier) GetProjectEnvID() int64 {
	args := m.Called()

	return args.Get(0).(int64)
}

func NewMockProjectEnvIDSupplier(envID int64) *MockProjectEnvIDSupplier {
	mockInstance := MockProjectEnvIDSupplier{}
	mockInstance.On("GetProjectEnvID").Return(envID)

	return &mockInstance
}
