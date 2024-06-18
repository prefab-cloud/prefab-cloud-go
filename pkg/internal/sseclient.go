package internal

import (
	"encoding/base64"
	"log/slog"
	"strconv"

	sse "github.com/r3labs/sse/v2"
	"google.golang.org/protobuf/proto"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/options"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

func BuildSSEClient(options options.Options) (*sse.Client, error) {
	apiURL, err := options.PrefabAPIURLEnvVarOrSetting()
	if err != nil {
		return nil, err
	}

	authString := base64.StdEncoding.EncodeToString([]byte("authuser:" + options.APIKey))

	client := sse.NewClient(apiURL + "/api/v1/sse/config")
	client.Headers = map[string]string{
		"Authorization":                "Basic " + authString,
		"X-PrefabCloud-Client-Version": ClientVersionHeader,
	}

	return client, nil
}

func StartSSEConnection(client *sse.Client, apiConfigStore *APIConfigStore) {
	client.Headers["x-prefab-start-at-id"] = strconv.FormatInt(apiConfigStore.highWatermark, 10)

	err := client.Subscribe("", func(msg *sse.Event) {
		decoded := make([]byte, base64.StdEncoding.DecodedLen(len(msg.Data)))

		numberOfBytesWritten, err := base64.StdEncoding.Decode(decoded, msg.Data)
		if err != nil {
			slog.Error("sse: error decoding base64 data", "err", err.Error())

			return
		}

		// Trim the decoded slice to the actual length of the decoded data
		decoded = decoded[:numberOfBytesWritten]

		var configs prefabProto.Configs

		err = proto.Unmarshal(decoded, &configs)
		if err != nil {
			slog.Error("sse: error unmarshalling proto", "err", err.Error())

			return
		}

		apiConfigStore.SetFromConfigsProto(&configs)
	})
	slog.Error("sse:", "err", err.Error())
}
