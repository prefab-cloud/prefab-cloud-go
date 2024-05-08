package prefab

import (
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"github.com/prefab-cloud/prefab-cloud-go/utils"
	"testing"
)

func createConfigValueAndAssertOk(value any, t *testing.T) *prefabProto.ConfigValue {
	val, ok := utils.Create(value)
	if !ok {
		t.Fatalf("Unable to create a config value given %v", value)
	}
	return val

}
