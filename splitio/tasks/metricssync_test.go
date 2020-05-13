package tasks

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/service/api"
	"github.com/splitio/go-split-commons/dtos"
	"github.com/splitio/go-split-commons/storage/mutexmap"
	"github.com/splitio/go-toolkit/logging"
)

func TestMetricsSyncTask(t *testing.T) {
	countersRequestReceived := atomic.Value{}
	countersRequestReceived.Store(false)
	gaugesRequestReceived := atomic.Value{}
	gaugesRequestReceived.Store(false)
	latenciesRequestReceived := atomic.Value{}
	latenciesRequestReceived.Store(false)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Error("Method should be POST")
		}

		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Error("Error reading body")
			return
		}

		switch r.URL.Path {
		case "/metrics/times":
			latenciesRequestReceived.Store(true)
			var latencies []dtos.LatenciesDTO
			err = json.Unmarshal(body, &latencies)
			if err != nil {
				t.Errorf("Error parsing json: %s", err)
				return
			}
			if latencies[0].MetricName != "metric1" {
				t.Error("Incorrect metric name")
			}
		case "/metrics/counters":
			countersRequestReceived.Store(true)
			var counters []dtos.CounterDTO
			err = json.Unmarshal(body, &counters)
			if err != nil {
				t.Errorf("Error parsing json: %s", err)
				return
			}
			if counters[0].MetricName != "counter1" || counters[0].Count != 1 {
				t.Error("Incorrect count received")
			}
		case "/metrics/gauge":
			gaugesRequestReceived.Store(true)
			var gauge dtos.GaugeDTO
			err = json.Unmarshal(body, &gauge)
			if err != nil {
				t.Errorf("Error parsing json: %s", err)
				return
			}
			if gauge.MetricName != "g1" || gauge.Gauge != 123 {
				t.Error("Incorrect gauge received")
			}
		default:
			t.Errorf("Incorrect url %s", r.URL.Path)
		}

	}))
	defer ts.Close()

	logger := logging.NewLogger(&logging.LoggerOptions{})
	metadata := dtos.Metadata{
		SDKVersion:  "go-0.1",
		MachineIP:   "192.168.0.123",
		MachineName: "machine1",
	}
	metricsRecorder := api.NewHTTPMetricsRecorder(
		"",
		&conf.SplitSdkConfig{
			Advanced: conf.AdvancedConfig{
				EventsURL: ts.URL,
				SdkURL:    ts.URL,
			},
		},
		metadata,
		logger,
	)

	metricsStorage := mutexmap.NewMMMetricsStorage()

	countersTask := NewRecordCountersTask(
		metricsStorage,
		metricsRecorder,
		1,
		logger,
	)
	gaugesTask := NewRecordGaugesTask(
		metricsStorage,
		metricsRecorder,
		1,
		logger,
	)
	latenciesTask := NewRecordLatenciesTask(
		metricsStorage,
		metricsRecorder,
		1,
		logger,
	)

	countersTask.Start()
	gaugesTask.Start()
	latenciesTask.Start()

	if !countersTask.IsRunning() || !gaugesTask.IsRunning() || !latenciesTask.IsRunning() {
		t.Error("Metrics recording tasks should be running")
	}

	metricsStorage.PutGauge("g1", 123)
	metricsStorage.IncLatency("metric1", 5)
	metricsStorage.IncCounter("counter1")
	time.Sleep(time.Second * 2)

	if !countersRequestReceived.Load().(bool) || !gaugesRequestReceived.Load().(bool) || !latenciesRequestReceived.Load().(bool) {
		t.Error("Not all expected requests received")
	}

	countersTask.Stop(true)
	gaugesTask.Stop(true)
	latenciesTask.Stop(true)
	if countersTask.IsRunning() || gaugesTask.IsRunning() || latenciesTask.IsRunning() {
		t.Error("Task should be stopped")
	}
}

type metricsRecorderMock struct {
	gaugeIterations   atomic.Value
	counterIterations atomic.Value
	latencyIterations atomic.Value
}

func (m *metricsRecorderMock) RecordLatencies(latencies []dtos.LatenciesDTO) error {
	m.latencyIterations.Store(m.latencyIterations.Load().(int) + 1)
	return nil
}

func (m *metricsRecorderMock) RecordCounters(counters []dtos.CounterDTO) error {
	m.counterIterations.Store(m.counterIterations.Load().(int) + 1)
	return nil
}

func (m *metricsRecorderMock) RecordGauge(gauge dtos.GaugeDTO) error {
	m.gaugeIterations.Store(m.gaugeIterations.Load().(int) + 1)
	return nil
}

func TestMetricsFlushWhenTaskIsStopped(t *testing.T) {
	logger := logging.NewLogger(nil)
	metricsStorage := mutexmap.NewMMMetricsStorage()
	metricsStorage.IncCounter("c1")
	metricsStorage.IncLatency("l1", 2)
	metricsStorage.PutGauge("g1", 1)
	metricsStorage.IncCounter("c1")
	metricsStorage.IncLatency("l1", 2)
	metricsRecorder := &metricsRecorderMock{}

	metricsRecorder.gaugeIterations.Store(0)
	metricsRecorder.counterIterations.Store(0)
	metricsRecorder.latencyIterations.Store(0)

	gaugeTask := NewRecordGaugesTask(
		metricsStorage,
		metricsRecorder,
		100,
		logger,
	)

	counterTask := NewRecordCountersTask(
		metricsStorage,
		metricsRecorder,
		100,
		logger,
	)

	latencyTask := NewRecordLatenciesTask(
		metricsStorage,
		metricsRecorder,
		100,
		logger,
	)

	gaugeTask.Start()
	latencyTask.Start()
	counterTask.Start()
	time.Sleep(time.Second * 3)

	if metricsRecorder.counterIterations.Load().(int) != 1 ||
		metricsRecorder.gaugeIterations.Load().(int) != 1 ||
		metricsRecorder.latencyIterations.Load().(int) != 1 {
		t.Error("All metrics should already have been flushed once")
	}

	// Add more data so that there's something to flush when tasks are stopped
	metricsStorage.IncCounter("c1")
	metricsStorage.IncLatency("l1", 2)
	metricsStorage.PutGauge("g1", 1)
	metricsStorage.IncCounter("c1")
	metricsStorage.IncLatency("l1", 2)

	gaugeTask.Stop(true)
	latencyTask.Stop(true)
	counterTask.Stop(true)

	time.Sleep(time.Second * 5)

	if metricsRecorder.gaugeIterations.Load().(int) != 2 {
		t.Error("Gauge Task should have ran four time")
	}

	if metricsRecorder.counterIterations.Load().(int) != 2 {
		t.Error("Counter Task should have ran twice")
	}

	if metricsRecorder.latencyIterations.Load().(int) != 2 {
		t.Error("Latency Task should have ran twice")
	}
}
