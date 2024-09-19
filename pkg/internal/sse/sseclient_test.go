package sse_test

import (
	"testing"

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
