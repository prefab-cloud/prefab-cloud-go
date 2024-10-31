package telemetry

import (
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/options"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type Aggregator interface {
	Record(data interface{})
	GetData() *prefabProto.TelemetryEvent
	Clear()
	Lock()
	Unlock()
}

func NewContextAggregators(opts options.Options) []Aggregator {
	switch opts.ContextTelemetryMode {
	case options.ContextTelemetryModes.PeriodicExample:
		return []Aggregator{NewExampleContextAggregator(), NewContextShapeAggregator()}
	case options.ContextTelemetryModes.Shapes:
		return []Aggregator{NewContextShapeAggregator()}
	default:
		return nil
	}
}
