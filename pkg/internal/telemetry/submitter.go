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

type QueueItem interface{}

type Submitter struct {
	aggregators                 []Aggregator
	contextAggregators          []Aggregator
	evaluationSummaryAggregator *EvaluationSummaryAggregator
	instanceHash                string
	host                        string
	apiKey                      string
	mutex                       *sync.Mutex
	queue                       chan QueueItem
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
		queue:                       make(chan QueueItem, 10000),
	}
}

func (ts *Submitter) SetupQueueConsumer() {
	go func() {
		for {
			for item := range ts.queue {
				switch item := item.(type) {
				case internal.ConfigMatch:
					ts.internalRecordEvaluation(item)
				case *contexts.ContextSet:
					ts.internalRecordContext(item)
				}
			}
		}
	}()
}

func (ts *Submitter) StartPeriodicSubmission(interval time.Duration) {
	ts.SetupQueueConsumer()

	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			if len(ts.aggregators) > 0 {
				ts.Submit(false)
			}
		}
	}()
}

func (ts *Submitter) enqueue(item QueueItem) {
	select {
	case ts.queue <- item:
		// Successfully enqueued
	default:
		// Queue is full, drop the item
	}
}

func (ts *Submitter) RecordEvaluation(data internal.ConfigMatch) {
	if ts.evaluationSummaryAggregator == nil || !data.IsMatch {
		return
	}

	// We don't track log level evaluations
	if _, ok := data.Match.GetType().(*prefabProto.ConfigValue_LogLevel); ok {
		return
	}

	ts.enqueue(data)
}

func (ts *Submitter) internalRecordEvaluation(data internal.ConfigMatch) {
	ts.evaluationSummaryAggregator.Record(data)
}

func (ts *Submitter) RecordContext(data *contexts.ContextSet) {
	if ts.contextAggregators == nil {
		return
	}

	ts.enqueue(data)
}

func (ts *Submitter) internalRecordContext(data *contexts.ContextSet) {
	for _, aggregator := range ts.contextAggregators {
		aggregator.Record(data)
	}
}

func (ts *Submitter) Submit(waitOnQueueToDrain bool) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if waitOnQueueToDrain {
		for len(ts.queue) > 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if len(ts.aggregators) == 0 {
		return errors.New("no aggregators to submit")
	}

	payload := Payload{
		InstanceHash: ts.instanceHash,
	}

	for _, aggregator := range ts.aggregators {
		aggregator.Lock()
		data := aggregator.GetData()
		if data != nil {
			payload.Events = append(payload.Events, data)
			aggregator.Clear()
		}
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

	return ts.retryRequest(req)
}

// retryRequest attempts an HTTP request with retries and exponential backoff
func (ts *Submitter) retryRequest(req *http.Request) error {
	const maxRetries = 5
	backoff := 1 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := client.Do(req)

		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			return nil
		}

		if resp != nil {
			defer resp.Body.Close()
		}

		if attempt == maxRetries {
			if err != nil {
				return fmt.Errorf("failed to submit telemetry after %d attempts: %v", attempt, err)
			}
			return fmt.Errorf("telemetry submission failed with status %s after %d attempts", resp.Status, attempt)
		}

		time.Sleep(backoff)
		backoff *= 2
	}

	return nil
}
