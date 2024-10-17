package integrationtestsupport

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/proto"

	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

func StartTestServer() (*[]*http.Request, *httptest.Server) {
	requests := []*http.Request{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Error reading body", "error", err)
		}
		defer r.Body.Close()

		slog.Debug("Received request", "method", r.Method, "url", r.URL, "body", string(body))

		r.Body = io.NopCloser(strings.NewReader(string(body)))
		requests = append(requests, r)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	return &requests, ts
}

func RequestBodyToTelemetryEventsProto(suite *suite.Suite, request *http.Request) *prefabProto.TelemetryEvents {
	body, err := io.ReadAll(request.Body)
	suite.Require().NoError(err, "failed to read request body")

	var payload prefabProto.TelemetryEvents
	err = proto.Unmarshal(body, &payload)
	suite.Require().NoError(err, "failed to unmarshal request body")

	return &payload
}
