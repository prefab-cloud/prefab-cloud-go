package telemetry_test

import (
	"testing"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	integrationtestsupport "github.com/prefab-cloud/prefab-cloud-go/pkg/internal/integration_test_support"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/telemetry"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/testutils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

// add some ConfigMatch	data to the aggregator
// run GetData() and assert

func TestEvaluationSummaryAggregator_Record(t *testing.T) {
	esa := telemetry.NewEvaluationSummaryAggregator()
	letter := testutils.CreateConfigValueAndAssertOk(t, "ABC")
	boolBoyTrue := testutils.CreateConfigValueAndAssertOk(t, true)
	boolBoyFalse := testutils.CreateConfigValueAndAssertOk(t, false)

	integrationtestsupport.MockNowProvider()

	var (
		lettersConfigID1 int64 = 1
		lettersConfigID2 int64 = 2
		boolBoyID1       int64 = 3
		boolBoyID2       int64 = 4
	)

	for i := 1; i <= 3; i++ {
		esa.Record(internal.ConfigMatch{
			ConfigID:              lettersConfigID1,
			ConfigType:            prefabProto.ConfigType_CONFIG,
			Match:                 letter,
			OriginalMatch:         letter,
			IsMatch:               true,
			ConfigKey:             "letters",
			ConditionalValueIndex: internal.IntPtr(1),
			RowIndex:              internal.IntPtr(2),
		})
	}

	esa.Record(internal.ConfigMatch{
		ConfigID:              lettersConfigID2,
		ConfigType:            prefabProto.ConfigType_CONFIG,
		Match:                 letter,
		OriginalMatch:         letter,
		IsMatch:               true,
		ConfigKey:             "letters",
		ConditionalValueIndex: internal.IntPtr(1),
		RowIndex:              internal.IntPtr(3),
	})

	for i := 1; i <= 2; i++ {
		esa.Record(internal.ConfigMatch{
			ConfigID:              boolBoyID1,
			ConfigType:            prefabProto.ConfigType_FEATURE_FLAG,
			Match:                 boolBoyTrue,
			OriginalMatch:         boolBoyTrue,
			IsMatch:               true,
			ConfigKey:             "bool.boy",
			ConditionalValueIndex: internal.IntPtr(1),
			RowIndex:              internal.IntPtr(2),
			WeightedValueIndex:    internal.IntPtr(3),
		})
	}

	esa.Record(internal.ConfigMatch{
		ConfigID:              boolBoyID2,
		ConfigType:            prefabProto.ConfigType_FEATURE_FLAG,
		Match:                 boolBoyTrue,
		OriginalMatch:         boolBoyTrue,
		IsMatch:               true,
		ConfigKey:             "bool.boy",
		ConditionalValueIndex: internal.IntPtr(1),
		RowIndex:              internal.IntPtr(2),
		WeightedValueIndex:    internal.IntPtr(3),
	})

	esa.Record(internal.ConfigMatch{
		ConfigID:              boolBoyID2,
		ConfigType:            prefabProto.ConfigType_FEATURE_FLAG,
		Match:                 boolBoyFalse,
		OriginalMatch:         boolBoyFalse,
		IsMatch:               true,
		ConfigKey:             "bool.boy",
		ConditionalValueIndex: internal.IntPtr(1),
		RowIndex:              internal.IntPtr(4),
		WeightedValueIndex:    internal.IntPtr(3),
	})

	summaries := []*prefabProto.ConfigEvaluationSummary{
		{
			Key:  "letters",
			Type: prefabProto.ConfigType_CONFIG,
			Counters: []*prefabProto.ConfigEvaluationCounter{
				{
					Count:                 3,
					ConfigId:              &lettersConfigID1,
					SelectedValue:         letter,
					ConditionalValueIndex: internal.Uint32Ptr(1),
					ConfigRowIndex:        internal.Uint32Ptr(2),
				},
				{
					Count:                 1,
					ConfigId:              &lettersConfigID2,
					SelectedValue:         letter,
					ConfigRowIndex:        internal.Uint32Ptr(3),
					ConditionalValueIndex: internal.Uint32Ptr(1),
				},
			},
		},
		{
			Key:  "bool.boy",
			Type: prefabProto.ConfigType_FEATURE_FLAG,
			Counters: []*prefabProto.ConfigEvaluationCounter{
				{
					Count:                 2,
					ConfigId:              &boolBoyID1,
					SelectedValue:         boolBoyTrue,
					ConditionalValueIndex: internal.Uint32Ptr(1),
					ConfigRowIndex:        internal.Uint32Ptr(2),
					WeightedValueIndex:    internal.Uint32Ptr(3),
				},
				{
					Count:                 1,
					ConfigId:              &boolBoyID2,
					SelectedValue:         boolBoyTrue,
					ConditionalValueIndex: internal.Uint32Ptr(1),
					ConfigRowIndex:        internal.Uint32Ptr(2),
					WeightedValueIndex:    internal.Uint32Ptr(3),
				},
				{
					Count:                 1,
					ConfigId:              &boolBoyID2,
					SelectedValue:         boolBoyFalse,
					ConditionalValueIndex: internal.Uint32Ptr(1),
					ConfigRowIndex:        internal.Uint32Ptr(4),
					WeightedValueIndex:    internal.Uint32Ptr(3),
				},
			},
		},
	}

	expectedData := &prefabProto.TelemetryEvent{
		Payload: &prefabProto.TelemetryEvent_Summaries{
			Summaries: &prefabProto.ConfigEvaluationSummaries{
				Start:     1,
				End:       2,
				Summaries: summaries,
			},
		},
	}

	testutils.AssertJSONEqual(t, expectedData, esa.GetData())
}
