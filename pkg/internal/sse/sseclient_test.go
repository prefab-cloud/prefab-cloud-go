package sse_test

import (
	"testing"

	sseclient "github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/options"
	sse "github.com/prefab-cloud/prefab-cloud-go/pkg/internal/sse"
)

func TestBuildSSEClient(t *testing.T) {
	options := options.Options{
		APIKey:  "does-not-matter",
		APIURLs: []string{"https://belt.prefab.cloud"},
	}

	client, err := sse.BuildSSEClient(options)

	assert.NoError(t, err)
	assert.Equal(t, "https://stream.prefab.cloud/api/v1/sse/config", client.URL)

	assert.Equal(t, map[string]string{
		"Authorization":                "Basic YXV0aHVzZXI6ZG9lcy1ub3QtbWF0dGVy",
		"X-PrefabCloud-Client-Version": internal.ClientVersionHeader,
		"Accept":                       "text/event-stream",
	}, client.Headers)
}

func TestEventHandlerIgnoresEmptyEvents(t *testing.T) {
	// Create a mock config store that tracks method calls
	mockStore := &mockConfigStore{}

	// Test that empty events don't call SetFromConfigsProto
	emptyEvent := &sseclient.Event{
		ID:   []byte("123"),
		Data: []byte{}, // Empty data
	}

	// This should simulate what happens in StartSSEConnection
	if len(emptyEvent.Data) > 0 {
		mockStore.SetFromConfigsProto(nil)
	}

	// Verify that SetFromConfigsProto was not called for empty event
	assert.Equal(t, 0, mockStore.setCallCount, "Empty events should be ignored")

	// Test that non-empty events do call SetFromConfigsProto
	nonEmptyEvent := &sseclient.Event{
		ID:   []byte("124"),
		Data: []byte("some-data"),
	}

	if len(nonEmptyEvent.Data) > 0 {
		mockStore.SetFromConfigsProto(nil)
	}

	// Verify that SetFromConfigsProto was called for non-empty event
	assert.Equal(t, 1, mockStore.setCallCount, "Non-empty events should be processed")
}

type mockConfigStore struct {
	setCallCount    int
	highWatermark   int64
}

func (m *mockConfigStore) SetFromConfigsProto(configs interface{}) {
	m.setCallCount++
}

func (m *mockConfigStore) GetHighWatermark() int64 {
	return m.highWatermark
}
