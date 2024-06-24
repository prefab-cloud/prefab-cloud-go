package internal

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/options"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

type HTTPClient struct {
	Options *options.Options
	apiURL  string
	cdnURL  string
}

func BuildHTTPClient(options options.Options) (*HTTPClient, error) {
	apiURL, err := options.PrefabAPIURLEnvVarOrSetting()
	if err != nil {
		return nil, err
	}

	cdnURL := strings.ReplaceAll(apiURL, ".", "-") + ".global.ssl.fastly.net"

	client := HTTPClient{Options: &options, apiURL: apiURL, cdnURL: cdnURL}

	return &client, nil
}

func (c *HTTPClient) Load(offset int32) (*prefabProto.Configs, error) {
	apiKey, err := c.Options.APIKeySettingOrEnvVar()
	if err != nil {
		return nil, err
	}

	for _, url := range []string{c.cdnURL, c.apiURL} {
		uri := fmt.Sprintf("%s/api/v1/configs/%d", url, offset)

		configs, err := c.LoadFromURI(uri, apiKey, offset)
		if err != nil {
			slog.Error("Error loading from URI", "err", err)

			continue
		}

		return configs, nil
	}

	return nil, errors.New("error loading configs from all URIs")
}

func (c *HTTPClient) LoadFromURI(uri string, apiKey string, offset int32) (*prefabProto.Configs, error) {
	slog.Info("Getting data from " + uri)

	// Perform the HTTP GET request
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth("1", apiKey)
	req.Header.Add("X-PrefabCloud-Client-Version", ClientVersionHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			slog.Error("Error closing response body", "err", err)
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
