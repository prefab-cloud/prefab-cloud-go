package testutils

import (
	"testing"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/utils"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

func CreateConfigValueAndAssertOk(t *testing.T, value any) *prefabProto.ConfigValue {
	t.Helper()

	val, ok := utils.Create(value)
	if !ok {
		t.Fatalf("Unable to create a config value given %v", value)
	}

	return val
}
