package integrationtestsupport

import (
	prefab "github.com/prefab-cloud/prefab-cloud-go/pkg"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type expectedContextShapeData struct {
	Name       string           `yaml:"name"`
	FieldTypes map[string]int32 `yaml:"field_types"`
}

type ContextShapeTestHarness struct {
	testCase TelemetryTestCase
}

func (c ContextShapeTestHarness) GetOptions() []prefab.Option {
	return []prefab.Option{prefab.WithContextTelemetryMode(prefab.ContextTelemetryMode.Shapes)}
}

func (c ContextShapeTestHarness) GetExpectedEvent() (*prefabProto.TelemetryEvent, error) {
	var expectedData []expectedContextShapeData

	err := unmarshalExampleData(c.testCase.Yaml.ExpectedData, &expectedData)
	if err != nil {
		return nil, err
	}

	shapes := []*prefabProto.ContextShape{}

	for _, item := range expectedData {
		shape := &prefabProto.ContextShape{
			Name:       item.Name,
			FieldTypes: item.FieldTypes,
		}

		shapes = append(shapes, shape)
	}

	return &prefabProto.TelemetryEvent{
		Payload: &prefabProto.TelemetryEvent_ContextShapes{
			ContextShapes: &prefabProto.ContextShapes{
				Shapes: shapes,
			},
		},
	}, nil
}

func (c ContextShapeTestHarness) Exercise(client *prefab.ContextBoundClient) error {
	context := ctxDataToContextSet(c.testCase.Yaml.Data.(map[string]interface{}))

	_, _, err := client.GetIntValue("does.not.exist", *context)

	return err
}

func (c ContextShapeTestHarness) MassagePayload(payload *prefabProto.TelemetryEvents) *prefabProto.TelemetryEvents {
	return payload
}
