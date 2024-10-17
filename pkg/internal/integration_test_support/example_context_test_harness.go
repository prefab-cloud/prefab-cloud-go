package integrationtestsupport

import (
	prefab "github.com/prefab-cloud/prefab-cloud-go/pkg"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ExampleContextTestHarness struct {
	testCase TelemetryTestCase
}

func (c ExampleContextTestHarness) GetOptions() []prefab.Option {
	return []prefab.Option{prefab.WithContextTelemetryMode(prefab.ContextTelemetryMode.PeriodicExample)}
}

func (c ExampleContextTestHarness) GetExpectedEvent() (*prefabProto.TelemetryEvent, error) {
	if c.testCase.Yaml.ExpectedData == nil {
		return nil, nil
	}

	contextSet := ctxDataToContextSet(c.testCase.Yaml.ExpectedData.(map[string]interface{}))

	return &prefabProto.TelemetryEvent{
		Payload: &prefabProto.TelemetryEvent_ExampleContexts{
			ExampleContexts: &prefabProto.ExampleContexts{
				Examples: []*prefabProto.ExampleContext{
					{
						Timestamp:  0,
						ContextSet: contextSet.ToProto(),
					},
				},
			},
		},
	}, nil
}

func (c ExampleContextTestHarness) Exercise(client *prefab.ContextBoundClient) error {
	context := ctxDataToContextSet(c.testCase.Yaml.Data.(map[string]interface{}))

	_, _, err := client.GetIntValue("does.not.exist", *context)

	return err
}

func (c ExampleContextTestHarness) MassagePayload(payload *prefabProto.TelemetryEvents) *prefabProto.TelemetryEvents {
	return payload
}
