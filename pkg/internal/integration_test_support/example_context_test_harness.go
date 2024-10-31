package integrationtestsupport

import (
	prefab "github.com/prefab-cloud/prefab-cloud-go/pkg"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/telemetry"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type ExampleContextTestHarness struct {
	testCase TelemetryTestCase
}

func (c ExampleContextTestHarness) GetOptions() []prefab.Option {
	return []prefab.Option{prefab.WithContextTelemetryMode(prefab.ContextTelemetryMode.PeriodicExample)}
}

func (c ExampleContextTestHarness) GetExpectedEvents() ([]*prefabProto.TelemetryEvent, error) {
	if c.testCase.Yaml.ExpectedData == nil {
		return nil, nil
	}

	contextSet := ctxDataToContextSet(c.testCase.Yaml.ExpectedData.(map[string]interface{}))

	return []*prefabProto.TelemetryEvent{
		// This is the primary payload, but example contexts also sends context shapes
		{
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
		},
		// This is the context shape payload
		{
			Payload: &prefabProto.TelemetryEvent_ContextShapes{
				ContextShapes: &prefabProto.ContextShapes{
					Shapes: contextShapesForContextSet(contextSet),
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

func contextShapesForContextSet(contextSet *contexts.ContextSet) []*prefabProto.ContextShape {
	shapes := []*prefabProto.ContextShape{}

	for _, context := range contextSet.Data {
		shape := &prefabProto.ContextShape{
			Name:       context.Name,
			FieldTypes: make(map[string]int32),
		}

		for key, value := range context.Data {
			shape.FieldTypes[key] = telemetry.FieldTypeForValue(value)
		}

		shapes = append(shapes, shape)
	}

	return shapes
}
