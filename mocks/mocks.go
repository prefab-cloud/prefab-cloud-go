package mocks

import (
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
