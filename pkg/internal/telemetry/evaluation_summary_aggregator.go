package telemetry

import (
	"fmt"
	"sync"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type EvaluationSummaryAggregator struct {
	name      string
	data      map[string]internal.ConfigMatch
	counts    map[string]int64
	dataStart int64
	mutex     *sync.Mutex
}

func NewEvaluationSummaryAggregator() *EvaluationSummaryAggregator {
	return &EvaluationSummaryAggregator{
		name:   "EvaluationSummaryAggregator",
		data:   make(map[string]internal.ConfigMatch),
		counts: make(map[string]int64),
		mutex:  &sync.Mutex{},
	}
}

func (esa *EvaluationSummaryAggregator) Lock() {
	esa.mutex.Lock()
}

func (esa *EvaluationSummaryAggregator) Unlock() {
	esa.mutex.Unlock()
}

func (esa *EvaluationSummaryAggregator) Record(data interface{}) {
	esa.Lock()
	defer esa.Unlock()

	if esa.dataStart == 0 {
		esa.dataStart = NowProvider()
	}

	evaluation := data.(internal.ConfigMatch)

	key := fmt.Sprintf("%d-%d-%d-%d-%s", evaluation.ConfigID, intUnlessNil(evaluation.RowIndex), intUnlessNil(evaluation.ConditionalValueIndex), intUnlessNil(evaluation.WeightedValueIndex), evaluation.Match.String())

	if _, ok := esa.data[key]; !ok {
		esa.data[key] = evaluation
		esa.counts[key] = 1
	} else {
		esa.counts[key]++
	}
}

func (esa *EvaluationSummaryAggregator) GetData() *prefabProto.TelemetryEvent {
	summaries := []*prefabProto.ConfigEvaluationSummary{}

	counterLookup := make(map[string]map[prefabProto.ConfigType][]*prefabProto.ConfigEvaluationCounter)

	for groupingKey, evaluation := range esa.data {
		key := evaluation.ConfigKey

		// Each of these is a ConfigEvaluationCounter
		if _, ok := counterLookup[key]; !ok {
			counterLookup[key] = make(map[prefabProto.ConfigType][]*prefabProto.ConfigEvaluationCounter)
		}

		if _, ok := counterLookup[key][evaluation.ConfigType]; !ok {
			counterLookup[key][evaluation.ConfigType] = []*prefabProto.ConfigEvaluationCounter{}
		}

		counter := &prefabProto.ConfigEvaluationCounter{
			ConfigId:              internal.Int64Ptr(evaluation.ConfigID),
			ConfigRowIndex:        coerceToUint32PtrUnlessNil(evaluation.RowIndex),
			ConditionalValueIndex: coerceToUint32PtrUnlessNil(evaluation.ConditionalValueIndex),
			WeightedValueIndex:    coerceToUint32PtrUnlessNil(evaluation.WeightedValueIndex),
			SelectedValue:         evaluation.Match,
			Count:                 esa.counts[groupingKey],
		}

		counterLookup[key][evaluation.ConfigType] = append(counterLookup[key][evaluation.ConfigType], counter)
	}

	for key, countersByType := range counterLookup {
		for configType, counters := range countersByType {
			summaries = append(summaries,
				&prefabProto.ConfigEvaluationSummary{
					Key:      key,
					Type:     configType,
					Counters: counters,
				})
		}
	}

	return &prefabProto.TelemetryEvent{
		Payload: &prefabProto.TelemetryEvent_Summaries{
			Summaries: &prefabProto.ConfigEvaluationSummaries{
				Start:     esa.dataStart,
				End:       NowProvider(),
				Summaries: summaries,
			},
		},
	}
}

func (esa *EvaluationSummaryAggregator) Clear() {
	esa.data = make(map[string]internal.ConfigMatch)
	esa.counts = make(map[string]int64)
	esa.dataStart = 0
}

func coerceToUint32PtrUnlessNil(value *int) *uint32 {
	if value == nil {
		return nil
	}

	return internal.Uint32Ptr(uint32(*value))
}

func intUnlessNil(value *int) int {
	if value == nil {
		return 0
	}

	return *value
}
