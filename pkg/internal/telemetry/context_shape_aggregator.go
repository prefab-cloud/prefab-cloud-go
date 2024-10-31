package telemetry

import (
	"sync"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ContextShapeAggregator struct {
	name  string
	data  map[string]map[string]int32
	mutex *sync.Mutex
}

func NewContextShapeAggregator() *ContextShapeAggregator {
	return &ContextShapeAggregator{
		name:  "ContextShapeAggregator",
		data:  make(map[string]map[string]int32),
		mutex: &sync.Mutex{},
	}
}

func (eca *ContextShapeAggregator) Lock() {
	eca.mutex.Lock()
}

func (eca *ContextShapeAggregator) Unlock() {
	eca.mutex.Unlock()
}

func FieldTypeForValue(value interface{}) int32 {
	switch value.(type) {
	case int, int32, int64:
		return 1
	case string:
		return 2
	case float64, float32:
		return 4
	case bool:
		return 5
	case []string, []interface{}:
		return 10
	default:
		// We default to String if the type isn't a primitive we support.
		// This is because we cast in criteria evaluation
		return 2
	}
}

func (eca *ContextShapeAggregator) Record(ctxSet interface{}) {
	eca.Lock()
	defer eca.Unlock()

	contextSet := ctxSet.(*contexts.ContextSet)

	for _, context := range contextSet.Data {
		for key, value := range context.Data {
			if _, ok := eca.data[context.Name]; !ok {
				eca.data[context.Name] = make(map[string]int32)
			}

			if _, ok := eca.data[context.Name][key]; !ok {
				eca.data[context.Name][key] = FieldTypeForValue(value)
			}
		}
	}
}

func (eca *ContextShapeAggregator) Clear() {
	eca.data = make(map[string]map[string]int32)
}

func (eca *ContextShapeAggregator) GetData() *prefabProto.TelemetryEvent {
	shapeMap := map[string]*prefabProto.ContextShape{}

	for key, value := range eca.data {
		if _, ok := shapeMap[key]; !ok {
			shapeMap[key] = &prefabProto.ContextShape{
				FieldTypes: make(map[string]int32),
				Name:       key,
			}
		}

		for k, v := range value {
			shapeMap[key].FieldTypes[k] = v
		}
	}

	shapes := []*prefabProto.ContextShape{}

	for _, shape := range shapeMap {
		shapes = append(shapes, shape)
	}

	event := &prefabProto.TelemetryEvent{
		Payload: &prefabProto.TelemetryEvent_ContextShapes{
			ContextShapes: &prefabProto.ContextShapes{
				Shapes: shapes,
			},
		},
	}

	return event
}
