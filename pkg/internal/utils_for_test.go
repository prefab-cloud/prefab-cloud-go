package internal_test

import (
	"testing"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/prefab-cloud/prefab-cloud-go/utils"
)

func createConfigValueAndAssertOk(t *testing.T, value any) *prefabProto.ConfigValue {
	t.Helper()

	val, ok := utils.Create(value)
	if !ok {
		t.Fatalf("Unable to create a config value given %v", value)
	}

	return val
}

func boolPtr(val bool) *bool {
	return &val
}

func int64Ptr(val int64) *int64 {
	return &val
}

func intPtr(val int) *int {
	return &val
}

func stringPtr(val string) *string {
	return &val
}
