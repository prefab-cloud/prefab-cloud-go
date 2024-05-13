package prefab

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"google.golang.org/protobuf/proto"
)

type HttpClient struct {
	Options             *Options
	apiUrl              string
	cdnUrl              string
	prefabVersionHeader string // TODO calculate this from version
}

func BuildHttpClient(options Options) (*HttpClient, error) {
	apiUrl, err := options.PrefabApiUrlEnvVarOrSetting()
	if err != nil {
		return nil, err
	}

	cdnUrl := strings.Replace(apiUrl, "api", "cdn", 1)
	client := HttpClient{Options: &options, apiUrl: apiUrl, cdnUrl: cdnUrl}

	return &client, nil
}

func (c *HttpClient) Load(offset int32) (*prefabProto.Configs, error) {
	apikey, err := c.Options.apiKeySettingOrEnvVar()
	if err != nil {
		return nil, err
	}
	// TODO target the cdn first
	uri := fmt.Sprintf("%s/api/v1/configs/%d", c.apiUrl, offset)

	slog.Info(fmt.Sprintf("Getting data from %s", uri))

	// Perform the HTTP GET request
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("1", apikey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// TODO something?
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error loading configs. Response code %s", resp.Status)
	}

	slog.Info(fmt.Sprintf("Received data from %s. Loading", uri))

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Deserialize the data into the protobuf message
	var msg prefabProto.Configs

	err = proto.Unmarshal(bodyBytes, &msg)
	if err != nil {
		return nil, err
	}
	// Use your protobuf message (msg) as needed
	return &msg, nil
}
