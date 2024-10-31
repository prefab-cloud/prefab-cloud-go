package telemetry

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/contexts"
	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal/options"
	prefabProto "github.com/prefab-cloud/prefab-cloud-go/proto"
)

var NowProvider = time.Now().UnixMilli
var client = &http.Client{}

type Submitter struct {
	aggregators                 []Aggregator
	contextAggregators          []Aggregator
	evaluationSummaryAggregator *EvaluationSummaryAggregator
	instanceHash                string
	host                        string
	apiKey                      string
	mutex                       *sync.Mutex
}

type Payload = prefabProto.TelemetryEvents

func NewTelemetrySubmitter(options options.Options) *Submitter {
	aggregators := []Aggregator{}

	contextAggregators := NewContextAggregators(options)

	if contextAggregators != nil {
		aggregators = append(aggregators, contextAggregators...)
	}

	var evaluationSummaryAggregator *EvaluationSummaryAggregator
	if options.CollectEvaluationSummaries {
		evaluationSummaryAggregator = NewEvaluationSummaryAggregator()
		aggregators = append(aggregators, evaluationSummaryAggregator)
	}

	return &Submitter{
		aggregators:                 aggregators,
		host:                        options.TelemetryHost,
		apiKey:                      options.APIKey,
		contextAggregators:          contextAggregators,
		evaluationSummaryAggregator: evaluationSummaryAggregator,
		mutex:                       &sync.Mutex{},
		instanceHash:                options.InstanceHash,
	}
}

func (ts *Submitter) StartPeriodicSubmission(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			if len(ts.aggregators) > 0 {
				ts.Submit()
			}
		}
	}()
}

func (ts *Submitter) RecordEvaluation(data internal.ConfigMatch) {
	if ts.evaluationSummaryAggregator == nil || !data.IsMatch {
		return
	}

	// We don't track log level evaluations
	if _, ok := data.Match.GetType().(*prefabProto.ConfigValue_LogLevel); ok {
		return
	}

	ts.evaluationSummaryAggregator.Record(data)
}

func (ts *Submitter) RecordContext(data *contexts.ContextSet) {
	if ts.contextAggregators == nil {
		return
	}

	for _, aggregator := range ts.contextAggregators {
		aggregator.Record(data)
	}
}

func (ts *Submitter) Submit() error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if len(ts.aggregators) == 0 {
		return errors.New("no aggregators to submit")
	}

	payload := Payload{
		InstanceHash: ts.instanceHash,
	}

	for _, aggregator := range ts.aggregators {
		aggregator.Lock()

		data := aggregator.GetData()
		if data == nil {
			aggregator.Unlock()
			continue
		}

		payload.Events = append(payload.Events, data)

		aggregator.Clear()
		aggregator.Unlock()
	}

	if len(payload.Events) == 0 {
		return nil
	}

	url := fmt.Sprintf("%s/api/v1/telemetry", ts.host)

	payloadData, err := proto.Marshal(&payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Accept", "application/x-protobuf")
	req.Header.Set("X-PrefabCloud-Client-Version", internal.ClientVersionHeader)

	encodedAuth := base64.StdEncoding.EncodeToString([]byte("authuser:" + ts.apiKey))
	req.Header.Set("Authorization", "Basic "+encodedAuth)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit telemetry: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telemetry submission failed with status: %s", resp.Status)
	}

	return nil
}
