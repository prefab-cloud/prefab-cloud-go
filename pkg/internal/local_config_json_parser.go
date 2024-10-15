package internal

import (
	"google.golang.org/protobuf/encoding/protojson"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type LocalConfigJSONParser struct{}

func (p *LocalConfigJSONParser) Parse(jsonData []byte) ([]*prefabProto.Config, int64, error) {
	var configs prefabProto.Configs

	err := protojson.Unmarshal(jsonData, &configs)
	if err != nil {
		return nil, 0, err
	}

	return configs.Configs, configs.ConfigServicePointer.ProjectEnvId, nil
}
