package prefab

import (
	"fmt"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"strings"
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
		panic(err)
	}
	// TODO target the cdn first
	uri := fmt.Sprintf("%s/api/v1/configs/%d", c.apiUrl, offset)

	// Perform the HTTP GET request
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		// TODO do not panic
		panic(err)
	}

	req.SetBasicAuth("1", apikey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// TODO do not panic
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// TODO something?
		}
	}(resp.Body)
	if resp.StatusCode >= 300 {
	}

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		// TODO do not panic
		panic(err)
	}

	// Deserialize the data into the protobuf message
	var msg prefabProto.Configs
	err = proto.Unmarshal(bodyBytes, &msg)
	if err != nil {
		panic(err)
	}
	// Use your protobuf message (msg) as needed
	return &msg, nil
}
