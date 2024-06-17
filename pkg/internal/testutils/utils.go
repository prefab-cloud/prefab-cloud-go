package testutils

import (
	"testing"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/prefab-cloud/prefab-cloud-go/utils"
)

func CreateConfigValueAndAssertOk(t *testing.T, value any) *prefabProto.ConfigValue {
	t.Helper()

	val, ok := utils.Create(value)
	if !ok {
		t.Fatalf("Unable to create a config value given %v", value)
	}

	return val
}
