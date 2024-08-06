package sse

import (
	"encoding/base64"
	"errors"
	"log/slog"
	"strconv"
	"time"

	sse "github.com/r3labs/sse/v2"
	"google.golang.org/protobuf/proto"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/options"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

func BuildSSEClient(options options.Options) (*sse.Client, error) {
	apiURLs, err := options.PrefabAPIURLEnvVarOrSetting()
	if err != nil {
		return nil, err
	}

	if len(apiURLs) == 0 {
		return nil, errors.New("no api urls provided")
	}

	authString := base64.StdEncoding.EncodeToString([]byte("authuser:" + options.APIKey))

	// TODO: handle multiple api urls
	client := sse.NewClient(apiURLs[0] + "/api/v1/sse/config")
	client.Headers = map[string]string{
		"Authorization":                "Basic " + authString,
		"X-PrefabCloud-Client-Version": internal.ClientVersionHeader,
		"Accept":                       "text/event-stream",
	}

	return client, nil
}

type ConfigStore interface {
	SetFromConfigsProto(configs *prefabProto.Configs)
	GetHighWatermark() int64
}

func StartSSEConnection(client *sse.Client, apiConfigStore ConfigStore) {
	for {
		client.Headers["x-prefab-start-at-id"] = strconv.FormatInt(apiConfigStore.GetHighWatermark(), 10)

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
		if err != nil {
			slog.Error("sse:", "err", err.Error())
		}

		// If we get here, the connection was closed. We should try to reconnect.
		// We sleep for a second to avoid hammering the server.
		time.Sleep(1 * time.Second)
	}
}
