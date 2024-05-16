package internal_test

import (
	"testing"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/prefab-cloud/prefab-cloud-go/utils"
)

func createConfigValueAndAssertOk(value any, t *testing.T) *prefabProto.ConfigValue {
	val, ok := utils.Create(value)
	if !ok {
		t.Fatalf("Unable to create a config value given %v", value)
	}

	return val
}

func boolPtr(val bool) *bool {
	return &val
}
