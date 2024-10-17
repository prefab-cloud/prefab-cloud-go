package telemetry

import (
	"sync"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ExampleContextAggregator struct {
	Data  map[string]*prefabProto.ExampleContext
	name  string
	mutex *sync.Mutex
}

func NewExampleContextAggregator() *ExampleContextAggregator {
	return &ExampleContextAggregator{
		name:  "ExampleContextAggregator",
		Data:  make(map[string]*prefabProto.ExampleContext),
		mutex: &sync.Mutex{},
	}
}

func (eca *ExampleContextAggregator) Lock() {
	eca.mutex.Lock()
}

func (eca *ExampleContextAggregator) Unlock() {
	eca.mutex.Unlock()
}

func (eca *ExampleContextAggregator) Record(ctxSet interface{}) {
	eca.Lock()
	defer eca.Unlock()

	contextSet := ctxSet.(*contexts.ContextSet)

	key := contextSet.GroupedKey()

	if key == "" {
		return
	}

	if _, ok := eca.Data[key]; !ok {
		example := &prefabProto.ExampleContext{
			Timestamp:  NowProvider(),
			ContextSet: contextSet.ToProto(),
		}

		eca.Data[key] = example
	}
}

func (eca *ExampleContextAggregator) Clear() {
	eca.Data = make(map[string]*prefabProto.ExampleContext)
}

func (eca *ExampleContextAggregator) GetData() *prefabProto.TelemetryEvent {
	if len(eca.Data) == 0 {
		return nil
	}

	examples := make([]*prefabProto.ExampleContext, 0, len(eca.Data))

	for _, example := range eca.Data {
		examples = append(examples, example)
	}

	event := &prefabProto.TelemetryEvent{
		Payload: &prefabProto.TelemetryEvent_ExampleContexts{
			ExampleContexts: &prefabProto.ExampleContexts{
				Examples: examples,
			},
		},
	}

	return event
}
