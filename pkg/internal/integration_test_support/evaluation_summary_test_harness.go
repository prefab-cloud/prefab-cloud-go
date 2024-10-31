package integrationtestsupport

import (
	prefab "github.com/prefab-cloud/prefab-cloud-go/pkg"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ExpectedEvaluationSummaryData struct {
	Key        string `yaml:"key"`
	Value      any    `yaml:"value"`
	ConfigType string `yaml:"type"`
	ValueType  string `yaml:"value_type"`
	Count      int    `yaml:"count"`
	Summary    struct {
		ConfigRowIndex        *uint32 `yaml:"config_row_index"`
		ConditionalValueIndex *uint32 `yaml:"conditional_value_index"`
		WeightedValueIndex    *uint32 `yaml:"weighted_value_index"`
	} `yaml:"summary"`
}

type EvaluationSummaryTestHarness struct {
	testCase TelemetryTestCase
}

func (c EvaluationSummaryTestHarness) ExpectedData() ([]ExpectedEvaluationSummaryData, error) {
	var expectedData []ExpectedEvaluationSummaryData
	err := unmarshalExampleData(c.testCase.Yaml.ExpectedData, &expectedData)

	return expectedData, err
}

func (c EvaluationSummaryTestHarness) GetOptions() []prefab.Option {
	return []prefab.Option{prefab.WithCollectEvaluationSummaries(true)}
}

func (c EvaluationSummaryTestHarness) GetExpectedEvents() ([]*prefabProto.TelemetryEvent, error) {
	summaries := []*prefabProto.ConfigEvaluationSummary{}

	expectedData, err := c.ExpectedData()
	if err != nil {
		return nil, err
	}

	for _, item := range expectedData {
		counters := []*prefabProto.ConfigEvaluationCounter{
			{
				Count:                 int64(item.Count),
				ConfigRowIndex:        item.Summary.ConfigRowIndex,
				ConditionalValueIndex: item.Summary.ConditionalValueIndex,
				WeightedValueIndex:    item.Summary.WeightedValueIndex,
			},
		}

		summary := &prefabProto.ConfigEvaluationSummary{
			Key:      item.Key,
			Type:     lookupConfigType(item.ConfigType),
			Counters: counters,
		}

		summaries = append(summaries, summary)
	}

	return []*prefabProto.TelemetryEvent{
		{
			Payload: &prefabProto.TelemetryEvent_Summaries{
				Summaries: &prefabProto.ConfigEvaluationSummaries{
					Summaries: summaries,
				},
			},
		},
	}, nil
}

func (c EvaluationSummaryTestHarness) Exercise(client *prefab.ContextBoundClient) error {
	expectedData, err := c.ExpectedData()
	if err != nil {
		return err
	}

	ctx := contexts.NewContextSet()

	for _, item := range expectedData {
		for i := 0; i < item.Count; i++ {
			var err error

			switch item.ValueType {
			case "string":
				_, _, err = client.GetStringValue(item.Key, *ctx)
			case "string_list":
				_, _, err = client.GetStringSliceValue(item.Key, *ctx)
			case "int":
				_, _, err = client.GetIntValue(item.Key, *ctx)
			default:
				panic("Unknown value type: " + item.ValueType)
			}

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func lookupConfigType(configType string) prefabProto.ConfigType {
	switch configType {
	case "CONFIG":
		return prefabProto.ConfigType_CONFIG
	case "FEATURE_FLAG":
		return prefabProto.ConfigType_FEATURE_FLAG
	default:
		panic("Unknown config type: " + configType)
	}
}

func (c EvaluationSummaryTestHarness) MassagePayload(events *prefabProto.TelemetryEvents) *prefabProto.TelemetryEvents {
	// The test data doesn't include config Id or selected value, so we can't compare those
	for _, event := range events.Events {
		payload := event.Payload.(*prefabProto.TelemetryEvent_Summaries)

		for _, summary := range payload.Summaries.Summaries {
			for _, counter := range summary.Counters {
				counter.ConfigId = nil
				counter.SelectedValue = nil
			}
		}
	}

	return events
}
