package client

import (
	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/impressionListener"

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
		cfg:    &conf.SplitSdkConfig{},
		logger: logger,
		sync: &sdkSync{
			countersSync:   countersTask,
			gaugeSync:      gaugesTask,
			impressionSync: impressionsTask,
			latenciesSync:  latenciesTask,
			segmentSync:    segmentsTask,
			splitSync:      splitTask,
		},
	}

	factory := SplitFactory{
		client: &client,
	}

	client.factory = &factory

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

	if !stoppedCounters {
		t.Error("Counters shoud be stopped")
	}

	if !stoppedGauge {
		t.Error("Gauge shoud be stopped")
	}

	if !stoppedImpressions {
		t.Error("Impressions shoud be stopped")
	}

	if !stoppedLatencies {
		t.Error("Latencies shoud be stopped")
	}

	if !stoppedSegments {
		t.Error("Segments shoud be stopped")
	}

	if !stoppedSplit {
		t.Error("Split shoud be stopped")
	}

	treatments := client.Treatments("key", []string{"feature1", "feature2", "feature3"}, nil)
	if len(treatments) != 3 {
		t.Error("Should return 3 treatments.")
	}

	if treatments["feature1"] != evaluator.Control {
		t.Error("Wrong treatment result")
	}

	if treatments["feature2"] != evaluator.Control {
		t.Error("Wrong treatment result")
	}

	if treatments["feature3"] != evaluator.Control {
		t.Error("Wrong treatment result")
	}
}

type ImpressionListenerTest struct {
}

var ilTest = make(map[string]interface{})

func (i *ImpressionListenerTest) LogImpression(data impressionlistener.ILObject) {
	ilTest["Feature"] = data.Impression.Feature
	ilTest["BucketingKey"] = data.Impression.BucketingKey
	ilTest["ChangeNumber"] = data.Impression.ChangeNumber
	ilTest["KeyName"] = data.Impression.KeyName
	ilTest["Label"] = data.Impression.Label
	ilTest["Time"] = data.Impression.Time
	ilTest["Treatment"] = data.Impression.Treatment
	ilTest["Attributes"] = data.Attributes
	ilTest["Version"] = data.SDKLanguageVersion
	ilTest["InstanceName"] = data.InstanceID
}

func TestImpressionListener(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	impTest := &ImpressionListenerTest{}
	impresiionL := impressionlistener.NewImpressionListenerWrapper(impTest)

	client := SplitClient{
		cfg:                cfg,
		evaluator:          &mockEvaluator{},
		impressions:        mutexmap.NewMMImpressionStorage(),
		logger:             logger,
		metrics:            mutexmap.NewMMMetricsStorage(),
		impressionListener: impresiionL,
	}

	attributes := make(map[string]interface{})
	attributes["One"] = "test"

	res := client.Treatment("user1", "feature", attributes)

	if res != "TreatmentA" {
		t.Error("Wrong Treatment result")
	}

	if ilTest["Feature"] != "feature" || ilTest["KeyName"] != "user1" || ilTest["Label"] != "aLabel" || ilTest["Treatment"] != "TreatmentA" || ilTest["ChangeNumber"] != int64(123) || ilTest["BucketingKey"] != "" {
		t.Error("Impression should match")
	}

	if ilTest["InstanceName"] != cfg.InstanceName {
		t.Error("Wrong instance name")
	}

	attr1, _ := ilTest["Attributes"].(map[string]interface{})
	if attr1["One"] != "test" {
		t.Error("Should match attribute")
	}

	expectedVersion := "go-" + splitio.Version
	if ilTest["Version"] != expectedVersion {
		t.Error("Version should be equals")
	}
}
