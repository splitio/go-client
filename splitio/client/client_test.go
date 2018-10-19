package client

import (
	"fmt"

	"github.com/splitio/go-client/splitio/service/dtos"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"

	"io/ioutil"
	"os"
	"testing"
	"time"
)

type mockEvaluator struct{}
type mockEvents struct{}
type mockEventsPanic struct{}

func (e *mockEvaluator) Evaluate(
	key string,
	bucketingKey *string,
	feature string,
	attributes map[string]interface{},
) *evaluator.Result {
	switch feature {
	case "feature":
		return &evaluator.Result{
			EvaluationTimeNs:  0,
			Label:             "aLabel",
			SplitChangeNumber: 123,
			Treatment:         "TreatmentA",
		}
	default:
		return &evaluator.Result{
			EvaluationTimeNs:  0,
			Label:             "exception",
			SplitChangeNumber: 123,
			Treatment:         evaluator.Control,
		}
	}
}

func (e *mockEventsPanic) Evaluate(
	key string,
	bucketingKey *string,
	feature string,
	attributes map[string]interface{},
) *evaluator.Result {
	panic("Testing panicking")
}

func (s *mockEvents) Push(event dtos.EventDTO) error { return nil }

func TestClientGetTreatment(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	client := SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexmap.NewMMImpressionStorage(),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
	}

	client.Treatment("key", "feature", nil)

	impressions := client.impressions.(storage.ImpressionStorage)
	impression := impressions.PopAll()[0].KeyImpressions[0]
	if impression.Label != "aLabel" {
		t.Error("Impression should have label when labelsEnabled is true")
	}

	client.cfg.LabelsEnabled = false
	client.Treatment("key", "feature2", nil)

	impression = impressions.PopAll()[0].KeyImpressions[0]
	if impression.Label != "" {
		t.Error("Impression should have label when labelsEnabled is true")
	}
}

func TestTreatments(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	client := SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexmap.NewMMImpressionStorage(),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
	}

	res := client.Treatments("user1", []string{"feature", "notFeature"}, nil)

	featureRes, ok := res["feature"]
	if !ok || featureRes != "TreatmentA" {
		t.Error("Incorrect result for \"feature\"")
	}

	notFeatureRes, ok := res["notFeature"]
	if !ok || notFeatureRes != evaluator.Control {
		t.Error("Incorrect result for \"notFeature\"")
	}
}

func TestTreatmentsEmpty(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	client := SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexmap.NewMMImpressionStorage(),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
	}

	res := client.Treatments("user1", []string{"", ""}, nil)

	if len(res) != 0 {
		t.Error("Should return empty map.")
	}
}

func TestLocalhostMode(t *testing.T) {
	file, err := ioutil.TempFile("", "splitio_tests")
	if err != nil {
		t.Error("Couldn't create temporary file for localhost client tests: ", err)
		return
	}

	file.Write([]byte("feature1 on\n"))
	file.Write([]byte("feature2 off\n"))
	file.Sync()

	sdkConf := conf.Default()
	sdkConf.SplitFile = file.Name()
	factory, _ := NewSplitFactory("localhost", sdkConf)
	client := factory.Client()

	if client.cfg.OperationMode != "localhost" {
		t.Error("Localhost operation mode should be set when received apikey is 'localhost'")
	}

	feature1 := client.Treatment("asd", "feature1", nil)
	if feature1 != "on" {
		t.Error("Feature1 retrieved incorrectly")
	}

	feature2 := client.Treatment("asd", "feature2", nil)
	if feature2 != "off" {
		t.Error("Feature2 retrieved incorrectly")
	}

	file.Close()
	os.Remove(file.Name())
}

func TestClientGetTreatmentConsideringValidationInputs(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	client := SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexmap.NewMMImpressionStorage(),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
		validator:   inputValidation{logger: logger},
	}

	feature1 := client.Treatment(nil, "feature", nil)
	if feature1 != "control" {
		t.Error("Feature1 retrieved incorrectly")
	}

	feature2 := client.Treatment(true, "feature", nil)
	if feature2 != "control" {
		t.Error("Feature2 retrieved incorrectly")
	}

	feature3 := client.Treatment(123, "feature", nil)
	if feature3 != "TreatmentA" {
		t.Error("Feature3 retrieved incorrectly")
	}

	feature4 := client.Treatment("key", "feature", nil)
	if feature4 != "TreatmentA" {
		t.Error("Feature4 retrieved incorrectly")
	}

	var key = &Key{
		MatchingKey:  "key",
		BucketingKey: "bucketing",
	}

	feature5 := client.Treatment(key, "feature", nil)
	if feature5 != "TreatmentA" {
		t.Error("Feature5 retrieved incorrectly")
	}
}

func TestClientTrackValidationInputs(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	client := SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		events:      &mockEvents{},
		impressions: mutexmap.NewMMImpressionStorage(),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
	}

	track1 := client.Track("key", "", "eventType", 123)
	if track1 == nil {
		t.Error("track1 retrieved incorrectly")
	}

	track2 := client.Track("key", "trafficType", "", 123)
	if track2 == nil {
		t.Error("track2 retrieved incorrectly")
	}

	track3 := client.Track("key", "trafficType", "eventType", nil)
	if track3 != nil {
		t.Error("track3 retrieved incorrectly")
	}

	track4 := client.Track("key", "trafficType", "eventType", "invalid")
	if track4 == nil {
		t.Error("track4 retrieved incorrectly")
	}

	track5 := client.Track("key", "trafficType", "eventType", 123)
	if track5 != nil {
		t.Error("track5 retrieved incorrectly")
	}

	track6 := client.Track("key", "trafficType", "eventType", 1.3)
	if track6 != nil {
		t.Error("track6 retrieved incorrectly")
	}
}

func TestClientPanicking(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	client := SplitClient{
		cfg:         cfg,
		evaluator:   &mockEventsPanic{},
		events:      &mockEvents{},
		impressions: mutexmap.NewMMImpressionStorage(),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
	}

	treatment := client.Treatment("key", "some", nil)
	if treatment != "control" {
		t.Error("treatment retrieved incorrectly")
	}
}

func TestClientDestroy(t *testing.T) {
	logger := logging.NewLogger(nil)

	resSplits := 0
	stoppedSplit := false
	resSegments := 0
	stoppedSegments := false
	resImpressions := 0
	stoppedImpressions := false
	resGauge := 0
	stoppedGauge := false
	resCounters := 0
	stoppedCounters := false
	resLatencies := 0
	stoppedLatencies := false

	fmt.Println(stoppedSplit, stoppedSegments, stoppedImpressions, stoppedGauge, stoppedCounters, stoppedLatencies)

	splitSync := func(l logging.LoggerInterface) error { resSplits++; return nil }
	splitStop := func(l logging.LoggerInterface) { stoppedSplit = true }
	segmentSync := func(l logging.LoggerInterface) error { resSegments++; return nil }
	segmentStop := func(l logging.LoggerInterface) { stoppedSegments = true }
	impressionSync := func(l logging.LoggerInterface) error { resImpressions++; return nil }
	impressionStop := func(l logging.LoggerInterface) { stoppedImpressions = true }
	gaugeSync := func(l logging.LoggerInterface) error { resGauge++; return nil }
	gaugeStop := func(l logging.LoggerInterface) { stoppedGauge = true }
	counterSync := func(l logging.LoggerInterface) error { resCounters++; return nil }
	counterStop := func(l logging.LoggerInterface) { stoppedCounters = true }
	latencySync := func(l logging.LoggerInterface) error { resLatencies++; return nil }
	latencyStop := func(l logging.LoggerInterface) { stoppedLatencies = true }

	splitTask := asynctask.NewAsyncTask("splits", splitSync, 100, nil, splitStop, logger)
	segmentsTask := asynctask.NewAsyncTask("segments", segmentSync, 100, nil, segmentStop, logger)
	impressionsTask := asynctask.NewAsyncTask("impressions", impressionSync, 100, nil, impressionStop, logger)
	gaugesTask := asynctask.NewAsyncTask("gauges", gaugeSync, 100, nil, gaugeStop, logger)
	countersTask := asynctask.NewAsyncTask("counters", counterSync, 100, nil, counterStop, logger)
	latenciesTask := asynctask.NewAsyncTask("latencies", latencySync, 100, nil, latencyStop, logger)

	splitTask.Start()
	segmentsTask.Start()
	impressionsTask.Start()
	gaugesTask.Start()
	countersTask.Start()
	latenciesTask.Start()

	client := SplitClient{
		cfg: &conf.SplitSdkConfig{},
		sync: &sdkSync{
			countersSync:   countersTask,
			gaugeSync:      gaugesTask,
			impressionSync: impressionsTask,
			latenciesSync:  latenciesTask,
			segmentSync:    segmentsTask,
			splitSync:      splitTask,
		},
	}

	time.Sleep(1 * time.Second)
	client.Destroy()
	time.Sleep(1 * time.Second)

	if splitTask.IsRunning() {
		t.Error("split task should be stopped")
	}

	if segmentsTask.IsRunning() {
		t.Error("segment task should be stopped")
	}

	if impressionsTask.IsRunning() {
		t.Error("impression task should be stopped")
	}

	if gaugesTask.IsRunning() {
		t.Error("gauges task should be stopped")
	}

	if countersTask.IsRunning() {
		t.Error("counters task should be stopped")
	}

	if latenciesTask.IsRunning() {
		t.Error("latencies task should be stopped")
	}

	// -----

	if resSplits != 1 {
		t.Error("Splits should have run once")
	}

	if resSegments != 1 {
		t.Error("Segments should have run once")
	}

	if resImpressions != 1 {
		t.Error("Impressions should have run once")
	}

	if resGauge != 1 {
		t.Error("Gauge should have run once")
	}

	if resCounters != 1 {
		t.Error("Conters should have run once")
	}

	if resLatencies != 1 {
		t.Error("Latencies should have run once")
	}

	if !client.IsDestroyed() {
		t.Error("Client should be destroyed")
	}

	if client.Treatment("key", "feature", nil) != evaluator.Control {
		t.Error("Single .Treatment() call should return control")
	}

	treatments := client.Treatments("key", []string{"feature1", "feature2", "feature3"}, nil)
	if len(treatments) != 0 {
		t.Error("Should return empty map.")
	}
}
