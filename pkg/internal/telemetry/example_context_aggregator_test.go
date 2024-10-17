package telemetry_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	prefab "github.com/prefab-cloud/prefab-cloud-go/pkg"
	integrationtestsupport "github.com/prefab-cloud/prefab-cloud-go/pkg/internal/integration_test_support"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/telemetry"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

func comparableExampleContext(ctx *prefabProto.ContextSet) []string {
	data := []string{}

	for _, set := range ctx.Contexts {
		data = append(data, strings.ReplaceAll(set.String(), "  ", " "))
	}

	sort.Strings(data)

	return data
}

func TestExampleContextAggregator_Record(t *testing.T) {
	assert := assert.New(t)

	integrationtestsupport.MockNowProvider()

	eca := telemetry.NewExampleContextAggregator()

	// Recording the initial context set adds a new entry
	contextSet := prefab.NewContextSet().
		WithNamedContextValues("host", map[string]interface{}{
			"key":    "h123",
			"name":   "HOSTNAME",
			"region": "REGION",
			"cpu":    12,
		}).WithNamedContextValues("user", map[string]interface{}{
		"key":   "u123",
		"email": "test@example.com",
	})

	eca.Record(contextSet)

	assert.Equal(1, len(eca.Data))
	assert.NotNil(eca.Data["host:h123|user:u123"])

	assert.Equal(int64(1), eca.Data["host:h123|user:u123"].Timestamp)

	data := comparableExampleContext(eca.Data["host:h123|user:u123"].ContextSet)

	assert.Equal(`type:"host" values:{key:"cpu" value:{int:12}} values:{key:"key" value:{string:"h123"}} values:{key:"name" value:{string:"HOSTNAME"}} values:{key:"region" value:{string:"REGION"}}`, data[0])
	assert.Equal(`type:"user" values:{key:"email" value:{string:"test@example.com"}} values:{key:"key" value:{string:"u123"}}`, data[1])

	// Recording another context set adds a new entry
	anotherContextSet := prefab.NewContextSet().
		WithNamedContextValues("host", map[string]interface{}{
			"key":    "h987",
			"name":   "HOSTNAME2",
			"region": "REGION2",
			"cpu":    8,
		}).WithNamedContextValues("user", map[string]interface{}{
		"key":   "u987",
		"email": "bort@example.com",
	})

	eca.Record(anotherContextSet)

	assert.Equal(2, len(eca.Data))
	assert.NotNil(eca.Data["host:h987|user:u987"])

	assert.Equal(int64(2), eca.Data["host:h987|user:u987"].Timestamp)

	data = comparableExampleContext(eca.Data["host:h987|user:u987"].ContextSet)

	assert.Equal(`type:"host" values:{key:"cpu" value:{int:8}} values:{key:"key" value:{string:"h987"}} values:{key:"name" value:{string:"HOSTNAME2"}} values:{key:"region" value:{string:"REGION2"}}`, data[0])
	assert.Equal(`type:"user" values:{key:"email" value:{string:"bort@example.com"}} values:{key:"key" value:{string:"u987"}}`, data[1])

	// Recording the same context keys again does not add a new entry
	modifiedContextSet := prefab.NewContextSet().
		WithNamedContextValues("host", map[string]interface{}{
			// Key is the same
			"key": "h123",
			// Name is different
			"name": "XYZ",
		}).WithNamedContextValues("user", map[string]interface{}{
		"key": "u123",
	})

	eca.Record(modifiedContextSet)
	assert.Equal(2, len(eca.Data))

	assert.Equal(int64(1), eca.Data["host:h123|user:u123"].Timestamp)

	data = comparableExampleContext(eca.Data["host:h123|user:u123"].ContextSet)

	assert.Equal(`type:"host" values:{key:"cpu" value:{int:12}} values:{key:"key" value:{string:"h123"}} values:{key:"name" value:{string:"HOSTNAME"}} values:{key:"region" value:{string:"REGION"}}`, data[0])
	assert.Equal(`type:"user" values:{key:"email" value:{string:"test@example.com"}} values:{key:"key" value:{string:"u123"}}`, data[1])

	// Recording an empty context does nothing
	emptyContextSet := prefab.NewContextSet()

	eca.Record(emptyContextSet)
	assert.Equal(2, len(eca.Data))
}
