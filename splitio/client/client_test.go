package client

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/engine/evaluator/impressionlabels"
	impressionlistener "github.com/splitio/go-client/splitio/impressionListener"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-client/splitio/storage/mutexqueue"
	"github.com/splitio/go-client/splitio/storage/redisdb"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
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
	case "feature2":
		return &evaluator.Result{
			EvaluationTimeNs:  0,
			Label:             "bLabel",
			SplitChangeNumber: 123,
			Treatment:         "TreatmentB",
		}
	case "some_feature":
		return &evaluator.Result{
			EvaluationTimeNs:  0,
			Label:             "bLabel",
			SplitChangeNumber: 123,
			Treatment:         evaluator.Control,
		}
	default:
		return &evaluator.Result{
			EvaluationTimeNs:  0,
			Label:             impressionlabels.SplitNotFound,
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

func (s *mockEvents) Push(event dtos.EventDTO, size int) error { return nil }

func TestClientGetTreatment(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	factory := SplitFactory{
		cfg: cfg,
		storages: sdkStorages{
			impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
			telemetry:   mutexmap.NewMMMetricsStorage(),
		},
		logger: logger,
	}

	client := factory.Client()
	client.evaluator = &mockEvaluator{}
	factory.status.Store(sdkStatusReady)
	client.Treatment("key", "feature", nil)

	impressionsQueue := client.impressions.(storage.ImpressionStorage)
	impressions, _ := impressionsQueue.PopN(cfg.Advanced.ImpressionsBulkSize)
	impression := impressions[0]
	if impression.Label != "aLabel" {
		t.Error("Impression should have label when labelsEnabled is true")
	}

	client.factory.cfg.LabelsEnabled = false
	client.Treatment("key", "feature2", nil)

	impressions, _ = impressionsQueue.PopN(cfg.Advanced.ImpressionsBulkSize)
	impression = impressions[0]
	if impression.Label != "" {
		t.Error("Impression should have label when labelsEnabled is true")
	}
}

func TestTreatments(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	factory := SplitFactory{
		cfg: cfg,
		storages: sdkStorages{
			impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
			telemetry:   mutexmap.NewMMMetricsStorage(),
		},
		logger: logger,
	}

	client := factory.Client()
	client.evaluator = &mockEvaluator{}
	factory.status.Store(sdkStatusReady)

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
	client.BlockUntilReady(1)

	if factory.cfg.OperationMode != "localhost" {
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

	if client.Track("somekey", "somett", "somee", nil, nil) != nil {
		t.Error("It should be ok")
	}

	client.Destroy()
	file.Close()
	os.Remove(file.Name())
}

func TestClientGetTreatmentConsideringValidationInputs(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	factory := SplitFactory{
		cfg: cfg,
		storages: sdkStorages{
			impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
			telemetry:   mutexmap.NewMMMetricsStorage(),
		},
		logger: logger,
	}

	client := factory.Client()
	client.evaluator = &mockEvaluator{}
	factory.status.Store(sdkStatusReady)

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

	factory := SplitFactory{
		cfg: cfg,
		storages: sdkStorages{
			impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
			telemetry:   mutexmap.NewMMMetricsStorage(),
			events:      &mockEvents{},
		},
		logger: logger,
	}

	client := factory.Client()
	client.evaluator = &mockEventsPanic{}
	factory.status.Store(sdkStatusReady)

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

	factory := &SplitFactory{
		tasks: sdkSync{
			counters:    countersTask,
			gauges:      gaugesTask,
			impressions: impressionsTask,
			latencies:   latenciesTask,
			segments:    segmentsTask,
			splits:      splitTask,
		},
		cfg: conf.Default(),
	}
	client := SplitClient{
		logger:  logger,
		factory: factory,
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

	if !client.isDestroyed() {
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

var ilResult = make(map[string]interface{})

func (i *ImpressionListenerTest) LogImpression(data impressionlistener.ILObject) {
	ilTest := make(map[string]interface{})
	ilTest["FeatureName"] = data.Impression.FeatureName
	ilTest["BucketingKey"] = data.Impression.BucketingKey
	ilTest["ChangeNumber"] = data.Impression.ChangeNumber
	ilTest["KeyName"] = data.Impression.KeyName
	ilTest["Label"] = data.Impression.Label
	ilTest["Time"] = data.Impression.Time
	ilTest["Treatment"] = data.Impression.Treatment
	ilTest["Attributes"] = data.Attributes
	ilTest["Version"] = data.SDKLanguageVersion
	ilTest["InstanceName"] = data.InstanceID

	ilResult[data.Impression.FeatureName] = ilTest
}

func compareListener(ilTest map[string]interface{}, f string, k string, l string, t string, c int64, b string, a string, i string, v string) bool {
	if ilTest["FeatureName"] != f || ilTest["KeyName"] != k || ilTest["Label"] != l || ilTest["Treatment"] != t || ilTest["ChangeNumber"] != c || ilTest["BucketingKey"] != b {
		return false
	}
	if ilTest["Version"] != v {
		return false
	}
	attr1, _ := ilTest["Attributes"].(map[string]interface{})
	if attr1["One"] != a {
		return false
	}
	return true
}

func TestImpressionListener(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	impTest := &ImpressionListenerTest{}
	impresionL := impressionlistener.NewImpressionListenerWrapper(impTest, &splitio.SdkMetadata{
		SDKVersion:  "go-" + splitio.Version,
		MachineIP:   "123.123.123.123",
		MachineName: "ip-123-123-123-123",
	})

	factory := &SplitFactory{
		cfg: cfg,
		metadata: splitio.SdkMetadata{
			SDKVersion: "go-" + splitio.Version,
		},
	}
	client := SplitClient{
		evaluator:          &mockEvaluator{},
		impressions:        mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
		logger:             logger,
		metrics:            mutexmap.NewMMMetricsStorage(),
		impressionListener: impresionL,
		factory:            factory,
	}

	factory.status.Store(sdkStatusReady)

	attributes := make(map[string]interface{})
	attributes["One"] = "test"

	res := client.Treatment("user1", "feature", attributes)

	if res != "TreatmentA" {
		t.Error("Wrong Treatment result")
	}

	expectedVersion := "go-" + splitio.Version

	if !compareListener(ilResult["feature"].(map[string]interface{}), "feature", "user1", "aLabel", "TreatmentA", int64(123), "", "test", cfg.InstanceName, expectedVersion) {
		t.Error("Impression should match")
	}
	ilResult = make(map[string]interface{})

	delete(ilResult, "feature")
}

func TestImpressionListenerForTreatments(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)
	impTest := &ImpressionListenerTest{}
	impresionL := impressionlistener.NewImpressionListenerWrapper(impTest, &splitio.SdkMetadata{
		SDKVersion:  "go-" + splitio.Version,
		MachineIP:   "123.123.123.123",
		MachineName: "ip-123-123-123-123",
	})

	factory := &SplitFactory{
		cfg: cfg,
		metadata: splitio.SdkMetadata{
			SDKVersion: splitio.Version,
		},
	}
	client := SplitClient{
		evaluator:          &mockEvaluator{},
		impressions:        mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
		logger:             logger,
		metrics:            mutexmap.NewMMMetricsStorage(),
		impressionListener: impresionL,
		factory:            factory,
	}

	factory.status.Store(sdkStatusReady)

	attributes := make(map[string]interface{})
	attributes["One"] = "test"

	res := client.Treatments("user1", []string{"feature", "feature2"}, attributes)

	if res["feature"] != "TreatmentA" || res["feature2"] != "TreatmentB" {
		t.Error("Wrong Treatment result")
	}

	if len(ilResult) != 2 {
		t.Error("Error on ImpressionListener")
	}

	expectedVersion := "go-" + splitio.Version

	if !compareListener(ilResult["feature"].(map[string]interface{}), "feature", "user1", "aLabel", "TreatmentA", int64(123), "", "test", cfg.InstanceName, expectedVersion) {
		t.Error("Impression should match")
	}

	if !compareListener(ilResult["feature2"].(map[string]interface{}), "feature2", "user1", "bLabel", "TreatmentB", int64(123), "", "test", cfg.InstanceName, expectedVersion) {
		t.Error("Impression should match")
	}
	ilResult = make(map[string]interface{})

	delete(ilResult, "feature")
	delete(ilResult, "feature2")
}

func TestBlockUntilReadyWrongTimerPassed(t *testing.T) {
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
	err = client.BlockUntilReady(-1)
	expected := "SDK Initialization: timer must be positive number"
	if err != nil && err.Error() != expected {
		t.Error("Error was expected")
	}

	manager := factory.Manager()
	err = manager.BlockUntilReady(-1)
	if err != nil && err.Error() != expected {
		t.Error("Error was expected")
	}
}

func TestBlockUntilReadyStatusLocalhost(t *testing.T) {
	file, err := ioutil.TempFile("", "splitio_tests")
	if err != nil {
		t.Error("Couldn't create temporary file for localhost client tests: ", err)
		return
	}

	file.Write([]byte("feature1 on\n"))
	file.Sync()

	sdkConf := conf.Default()
	sdkConf.SplitFile = file.Name()

	impTest := &ImpressionListenerTest{}
	sdkConf.Advanced.ImpressionListener = impTest

	factory, _ := NewSplitFactory("localhost", sdkConf)

	client := factory.Client()
	manager := factory.Manager()

	if len(manager.SplitNames()) != 0 {
		t.Error("It should not return splits")
	}

	if client.factory.IsReady() {
		t.Error("Client should not be ready")
	}

	err = client.Track("something", "something", "something", nil, nil)
	if err != nil {
		t.Error("It should not return error")
	}

	attributes := make(map[string]interface{})
	attributes["One"] = "test"

	if client.Treatment("something", "something", attributes) != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	if client.Treatment("something", "something", nil) != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	features := []string{"something"}
	result := client.Treatments("something", features, nil)
	if result["something"] != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	err = client.BlockUntilReady(1)
	if err != nil {
		t.Error("Error was not expected")
	}

	if !client.factory.IsReady() {
		t.Error("Client should be ready")
	}

	if !manager.factory.IsReady() {
		t.Error("Manager should be ready")
	}

	err = client.Track("something", "something", "something", nil, nil)
	if err != nil {
		t.Error("It should not return error")
	}

	feature1 := client.Treatment("asd", "feature1", nil)
	if feature1 != "on" {
		t.Error("Feature1 retrieved incorrectly")
	}

	if manager.SplitNames()[0] != "feature1" {
		t.Error("It should return splits")
	}
}

func TestBlockUntilReadyStatusLocalhostOnDestroy(t *testing.T) {
	file, err := ioutil.TempFile("", "splitio_tests")
	if err != nil {
		t.Error("Couldn't create temporary file for localhost client tests: ", err)
		return
	}

	file.Write([]byte("feature1 on\n"))
	file.Sync()

	sdkConf := conf.Default()
	sdkConf.SplitFile = file.Name()

	factory, _ := NewSplitFactory("localhost", sdkConf)

	client := factory.Client()
	manager := factory.Manager()

	if len(manager.SplitNames()) != 0 {
		t.Error("It should not return splits")
	}

	if client.factory.IsReady() {
		t.Error("Client should not be ready")
	}

	err = client.BlockUntilReady(1)
	if err != nil {
		t.Error("Error was not expected")
	}

	if !client.factory.IsReady() {
		t.Error("Client should be ready")
	}

	if !manager.factory.IsReady() {
		t.Error("Manager should be ready")
	}

	client.Destroy()

	if !client.factory.IsDestroyed() {
		t.Error("Client should be destroyed")
	}

	if !manager.factory.IsDestroyed() {
		t.Error("Manager should be destroyed")
	}

	err = manager.BlockUntilReady(1)
	expected := "SDK Initialization: Client is destroyed"
	if err == nil || err.Error() != expected {
		t.Error("It should return an error")
	}
}

func TestBlockUntilReadyRedis(t *testing.T) {
	sdkConf := conf.Default()
	sdkConf.OperationMode = "redis-consumer"

	factory, _ := NewSplitFactory("something", sdkConf)

	if !factory.IsReady() {
		t.Error("Factory should be ready immediately")
	}

	client := factory.Client()
	if !client.factory.IsReady() {
		t.Error("Client should be ready immediately")
	}
	err := client.BlockUntilReady(1)
	if err != nil {
		t.Error("Error was not expected")
	}

	manager := factory.Manager()
	if !manager.factory.IsReady() {
		t.Error("Manager should be ready immediately")
	}
	err = manager.BlockUntilReady(1)
	if err != nil {
		t.Error("Error was not expected")
	}
}

func TestBlockUntilReadyInMemoryError(t *testing.T) {
	sdkConf := conf.Default()
	impTest := &ImpressionListenerTest{}
	sdkConf.Advanced.ImpressionListener = impTest

	attributes := make(map[string]interface{})
	attributes["One"] = "test"

	expectedVersion := "go-" + splitio.Version

	factory, _ := NewSplitFactory("something", sdkConf)

	if factory.IsReady() {
		t.Error("Factory should not be ready")
	}

	client := factory.Client()
	if client.factory.IsReady() {
		t.Error("Client should not be ready")
	}

	if client.Treatment("not_ready", "not_ready", attributes) != evaluator.Control {
		t.Error("Wrong evaluation")
	}
	if !compareListener(ilResult["not_ready"].(map[string]interface{}), "not_ready", "not_ready", "not ready", "control", int64(0), "", "test", cfg.InstanceName, expectedVersion) {
		t.Error("Impression should match")

	}
	ilResult = make(map[string]interface{})

	if client.Treatment("something", "something", nil) != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	features := []string{"something"}
	result := client.Treatments("something", features, nil)
	if result["something"] != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	err := client.Track("something", "something", "something", nil, nil)
	if err != nil {
		t.Error("It should not return error")
	}

	expected := "Client Instantiation: Client is not ready yet"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	err = client.BlockUntilReady(1)
	if err == nil {
		t.Error("It should return error")
	}

	if err != nil && err.Error() != "SDK Initialization failed" {
		t.Error("Wrong error")
	}
}

func TestBlockUntilReadyInMemory(t *testing.T) {
	mockedSplit1 := dtos.SplitDTO{
		Algo:                  2,
		ChangeNumber:          123,
		DefaultTreatment:      "default",
		Killed:                false,
		Name:                  "split",
		Seed:                  1234,
		Status:                "ACTIVE",
		TrafficAllocation:     1,
		TrafficAllocationSeed: -1667452163,
		TrafficTypeName:       "tt1",
		Conditions: []dtos.ConditionDTO{
			{
				ConditionType: "ROLLOUT",
				Label:         "in segment all",
				MatcherGroup: dtos.MatcherGroupDTO{
					Combiner: "AND",
					Matchers: []dtos.MatcherDTO{
						{
							MatcherType:        "ALL_KEYS",
							Whitelist:          nil,
							Negate:             false,
							UserDefinedSegment: nil,
						},
					},
				},
				Partitions: []dtos.PartitionDTO{
					{
						Size:      100,
						Treatment: "on",
					},
				},
			},
		},
	}
	mockedSplit2 := dtos.SplitDTO{Name: "split2", Killed: true, Status: "ACTIVE"}
	mockedSplit3 := dtos.SplitDTO{Name: "split3", Killed: true, Status: "INACTIVE"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(4 * time.Second)
		if r.URL.Path != "/splits" && r.Method != "GET" {
			t.Error("Invalid request. Should be GET to /splits")
		}

		splitChanges := dtos.SplitChangesDTO{
			Splits: []dtos.SplitDTO{mockedSplit1, mockedSplit2, mockedSplit3},
			Since:  3,
			Till:   3,
		}

		raw, err := json.Marshal(splitChanges)
		if err != nil {
			t.Error("Error building json")
			return
		}

		w.Write(raw)
	}))
	defer ts.Close()

	segmentMock, _ := ioutil.ReadFile("../../testdata/segment_mock.json")

	tss := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		fmt.Fprintln(w, fmt.Sprintf(string(segmentMock)))
	}))
	defer tss.Close()

	sdkConf := conf.Default()
	sdkConf.Advanced.EventsURL = tss.URL
	sdkConf.Advanced.SdkURL = ts.URL
	impTest := &ImpressionListenerTest{}
	sdkConf.Advanced.ImpressionListener = impTest

	attributes := make(map[string]interface{})
	attributes["One"] = "test"

	expectedVersion := "go-" + splitio.Version

	factory, _ := NewSplitFactory("something", sdkConf)

	if factory.IsReady() {
		t.Error("Factory should not be ready")
	}

	client := factory.Client()
	if client.factory.IsReady() {
		t.Error("Client should not be ready")
	}

	manager := factory.Manager()
	if manager.factory.IsReady() {
		t.Error("Manager should not be ready")
	}

	if len(manager.SplitNames()) != 0 {
		t.Error("It should not return splits")
	}

	if client.Treatment("not_ready2", "not_ready2", attributes) != evaluator.Control {
		t.Error("Wrong evaluation")
	}
	if !compareListener(ilResult["not_ready2"].(map[string]interface{}), "not_ready2", "not_ready2", "not ready", "control", int64(0), "", "test", cfg.InstanceName, expectedVersion) {
		t.Error("Impression should match")
	}

	result := client.Treatments("not_ready3", []string{"not_ready3"}, attributes)
	if result["not_ready3"] != evaluator.Control {
		t.Error("Wrong evaluation")
	}
	if !compareListener(ilResult["not_ready3"].(map[string]interface{}), "not_ready3", "not_ready3", "not ready", "control", int64(0), "", "test", cfg.InstanceName, expectedVersion) {
		t.Error("Impression should match")
	}
	ilResult = make(map[string]interface{})

	err := client.Track("something", "something", "something", nil, nil)
	if err != nil {
		t.Error("It should not return error")
	}

	expected := "Client Instantiation: Client is not ready yet"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	err = client.BlockUntilReady(1)
	if err == nil {
		t.Error("It should return error")
	}

	expected2 := "SDK Initialization: time of 1 exceeded"
	if err != nil && err.Error() != expected2 {
		t.Error("Wrong message error")
	}

	if client.factory.IsReady() {
		t.Error("Client should not be ready")
	}

	err = manager.BlockUntilReady(2)
	if err == nil {
		t.Error("It should return error")
	}

	expected2 = "SDK Initialization: time of 2 exceeded"
	if err != nil && err.Error() != expected2 {
		t.Error("Wrong message error")
	}

	err = client.BlockUntilReady(2)
	if err != nil && err.Error() != expected2 {
		t.Error("Wrong message error")
	}

	if !client.factory.IsReady() || !manager.factory.IsReady() {
		t.Error("Both client and manager should be ready")
	}

	if len(manager.SplitNames()) != 2 {
		t.Error("It should return Splits")
	}

	if client.Treatment("aaaaaaklmnbv", "split", nil) != "on" {
		t.Error("Treatment error")
	}

	client.Destroy()
}

var valid = &dtos.SplitDTO{
	Algo:                  2,
	ChangeNumber:          1494593336752,
	DefaultTreatment:      "off",
	Killed:                false,
	Name:                  "valid",
	Seed:                  -1992295819,
	Status:                "ACTIVE",
	TrafficAllocation:     100,
	TrafficAllocationSeed: -285565213,
	TrafficTypeName:       "user",
	Configurations:        map[string]string{"on": "{\"color\": \"blue\",\"size\": 13}"},
	Conditions: []dtos.ConditionDTO{
		{
			ConditionType: "ROLLOUT",
			Label:         "default rule",
			MatcherGroup: dtos.MatcherGroupDTO{
				Combiner: "AND",
				Matchers: []dtos.MatcherDTO{
					{
						KeySelector: &dtos.KeySelectorDTO{
							TrafficType: "user",
							Attribute:   nil,
						},
						MatcherType: "IN_SEGMENT",
						Whitelist:   nil,
						Negate:      false,
						UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
							SegmentName: "employees",
						},
					},
				},
			},
			Partitions: []dtos.PartitionDTO{
				{
					Size:      100,
					Treatment: "on",
				},
			},
		},
	},
}

var killed = &dtos.SplitDTO{
	Algo:                  2,
	ChangeNumber:          1494593336752,
	DefaultTreatment:      "defTreatment",
	Killed:                true,
	Name:                  "killed",
	Seed:                  -1992295819,
	Status:                "ACTIVE",
	TrafficAllocation:     100,
	TrafficAllocationSeed: -285565213,
	TrafficTypeName:       "user",
	Configurations:        map[string]string{"defTreatment": "{\"color\": \"orange\",\"size\": 15}"},
	Conditions: []dtos.ConditionDTO{
		{
			ConditionType: "ROLLOUT",
			Label:         "default rule",
			MatcherGroup: dtos.MatcherGroupDTO{
				Combiner: "AND",
				Matchers: []dtos.MatcherDTO{
					{
						KeySelector: &dtos.KeySelectorDTO{
							TrafficType: "user",
							Attribute:   nil,
						},
						MatcherType: "IN_SEGMENT",
						Whitelist:   nil,
						Negate:      false,
						UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
							SegmentName: "employees",
						},
					},
				},
			},
			Partitions: []dtos.PartitionDTO{
				{
					Size:      100,
					Treatment: "off",
				},
			},
		},
	},
}

var noConfig = &dtos.SplitDTO{
	Algo:                  2,
	ChangeNumber:          1494593336752,
	DefaultTreatment:      "defTreatment",
	Killed:                false,
	Name:                  "noConfig",
	Seed:                  -1992295819,
	Status:                "ACTIVE",
	TrafficAllocation:     100,
	TrafficAllocationSeed: -285565213,
	TrafficTypeName:       "user",
	Conditions: []dtos.ConditionDTO{
		{
			ConditionType: "ROLLOUT",
			Label:         "default rule",
			MatcherGroup: dtos.MatcherGroupDTO{
				Combiner: "AND",
				Matchers: []dtos.MatcherDTO{
					{
						KeySelector: &dtos.KeySelectorDTO{
							TrafficType: "user",
							Attribute:   nil,
						},
						MatcherType: "IN_SEGMENT",
						Whitelist:   nil,
						Negate:      false,
						UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
							SegmentName: "employees",
						},
					},
				},
			},
			Partitions: []dtos.PartitionDTO{
				{
					Size:      100,
					Treatment: "off",
				},
			},
		},
	},
}

type mockStorage struct{}

func (s *mockStorage) Get(
	feature string,
) *dtos.SplitDTO {
	switch feature {
	default:
	case "valid":
		return valid
	case "killed":
		return killed
	}
	return nil
}
func (s *mockStorage) GetAll() []dtos.SplitDTO                   { return make([]dtos.SplitDTO, 0) }
func (s *mockStorage) SegmentNames() *set.ThreadUnsafeSet        { return nil }
func (s *mockStorage) SplitNames() []string                      { return make([]string, 0) }
func (s *mockStorage) TrafficTypeExists(trafficType string) bool { return true }

type mockSegmentStorage struct{}

func (i *mockSegmentStorage) Get(feature string) *set.ThreadUnsafeSet {
	switch feature {
	default:
	case "employees":
		return set.NewSet("user1")
	}
	return nil
}

func isInvalidImpression(client SplitClient, key string, feature string, treatment string) bool {
	impressionsQueue := client.impressions.(storage.ImpressionStorage)
	impressions, _ := impressionsQueue.PopN(cfg.Advanced.ImpressionsBulkSize)
	i := impressions[0]

	if i.FeatureName != feature || i.KeyName != key || treatment != i.Treatment {
		return true
	}
	return false
}

func TestClient(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	evaluator := evaluator.NewEvaluator(
		&mockStorage{},
		&mockSegmentStorage{},
		nil,
		logger,
	)

	factory := &SplitFactory{cfg: cfg}
	client := SplitClient{
		evaluator:   evaluator,
		impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
		validator:   inputValidation{logger: logger},
		factory:     factory,
	}

	factory.status.Store(sdkStatusReady)

	// Assertions Treatment
	if client.Treatment("user1", "valid", nil) != "on" {
		t.Error("Unexpected Treatment Result")
	}
	if isInvalidImpression(client, "user1", "valid", "on") {
		t.Error("Wrong impression saved")
	}

	if client.Treatment("invalid", "valid", nil) != "off" {
		t.Error("Unexpected Treatment Result")
	}
	if isInvalidImpression(client, "invalid", "valid", "off") {
		t.Error("Wrong impression saved")
	}

	if client.Treatment("invalid", "invalid", nil) != "control" {
		t.Error("Unexpected Treatment Result")
	}

	if client.Treatment("invalid", "killed", nil) != "defTreatment" {
		t.Error("Unexpected Treatment Result")
	}
	if isInvalidImpression(client, "invalid", "killed", "defTreatment") {
		t.Error("Wrong impression saved")
	}

	// Assertion Treatments
	treatments := client.Treatments("user1", []string{"valid", "invalid", "killed"}, nil)
	if treatments["invalid"] != "control" {
		t.Error("Unexpected treatment result")
	}
	if treatments["killed"] != "defTreatment" {
		t.Error("Unexpected treatment result")
	}
	if treatments["valid"] != "on" {
		t.Error("Unexpected treatment result")
	}
	client.impressions.(storage.ImpressionStorage).PopN(cfg.Advanced.ImpressionsBulkSize)

	// Assertion TreatmentWithConfig
	result := client.TreatmentWithConfig("user1", "valid", nil)
	if result.Treatment != "on" {
		t.Error("Unexpected Treatment Result")
	}
	if *result.Config != "{\"color\": \"blue\",\"size\": 13}" {
		t.Error("Unexpected Config Result")
	}
	if isInvalidImpression(client, "user1", "valid", "on") {
		t.Error("Wrong impression saved")
	}

	result = client.TreatmentWithConfig("invalid", "valid", nil)
	if result.Treatment != "off" {
		t.Error("Unexpected Treatment Result")
	}
	if result.Config != nil {
		t.Error("Unexpected Config Result")
	}
	if isInvalidImpression(client, "invalid", "valid", result.Treatment) {
		t.Error("Wrong impression saved")
	}

	result = client.TreatmentWithConfig("invalid", "invalid", nil)
	if result.Treatment != "control" {
		t.Error("Unexpected Treatment Result")
	}
	if result.Config != nil {
		t.Error("Unexpected Config Result")
	}

	result = client.TreatmentWithConfig("invalid", "killed", nil)
	if result.Treatment != "defTreatment" {
		t.Error("Unexpected Treatment Result")
	}
	if *result.Config != "{\"color\": \"orange\",\"size\": 15}" {
		t.Error("Unexpected Config Result")
	}
	if isInvalidImpression(client, "invalid", "killed", result.Treatment) {
		t.Error("Wrong impression saved")
	}

	// Assertion TreatmentsWithConfig
	treatmentsWithConfigs := client.TreatmentsWithConfig("user1", []string{"valid", "invalid", "killed"}, nil)
	if treatmentsWithConfigs["invalid"].Treatment != "control" {
		t.Error("Unexpected treatment result")
	}
	if treatmentsWithConfigs["invalid"].Config != nil {
		t.Error("Unexpected Config Result")
	}
	if treatmentsWithConfigs["killed"].Treatment != "defTreatment" {
		t.Error("Unexpected treatment result")
	}
	if *treatmentsWithConfigs["killed"].Config != "{\"color\": \"orange\",\"size\": 15}" {
		t.Error("Unexpected Config Result")
	}
	if treatmentsWithConfigs["valid"].Treatment != "on" {
		t.Error("Unexpected treatment result")
	}
	if *treatmentsWithConfigs["valid"].Config != "{\"color\": \"blue\",\"size\": 13}" {
		t.Error("Unexpected Config Result")
	}
}

func TestLocalhostModeYAML(t *testing.T) {
	sdkConf := conf.Default()
	sdkConf.SplitFile = "../../testdata/splits.yaml"
	factory, _ := NewSplitFactory("localhost", sdkConf)
	client := factory.Client()
	manager := factory.Manager()

	_ = client.BlockUntilReady(5)

	if !client.isReady() {
		t.Error("Localhost should be ready")
	}

	if client.factory.cfg.OperationMode != "localhost" {
		t.Error("Localhost operation mode should be set when received apikey is 'localhost'")
	}

	if len(manager.Splits()) != 4 {
		t.Error("Error grabbing splits for localhost mode")
	}

	result := client.Treatment("only_key", "my_feature", nil)
	if result != "off" {
		t.Error("Treatment retrieved incorrectly")
	}

	result = client.Treatment("invalid_key", "my_feature", nil)
	if result != "control" {
		t.Error("Treatment retrieved incorrectly")
	}

	result = client.Treatment("key", "my_feature", nil)
	if result != "on" {
		t.Error("Treatment retrieved incorrectly")
	}

	result = client.Treatment("key2", "other_feature", nil)
	if result != "on" {
		t.Error("Treatment retrieved incorrectly")
	}

	result = client.Treatment("test", "other_feature_2", nil)
	if result != "on" {
		t.Error("Treatment retrieved incorrectly", result)
	}

	result = client.Treatment("key", "other_feature_3", nil)
	if result != "off" {
		t.Error("Treatment retrieved incorrectly")
	}

	result = client.Treatment("key_whitelist", "other_feature_3", nil)
	if result != "on" {
		t.Error("Treatment retrieved incorrectly")
	}

	resultWithConfigs := client.TreatmentWithConfig("only_key", "my_feature", nil)
	if resultWithConfigs.Treatment != "off" {
		t.Error("Treatment retrieved incorrectly")
	}
	if *resultWithConfigs.Config != "{\"desc\" : \"this applies only to OFF and only for only_key. The rest will receive ON\"}" {
		t.Error("Wronf config returned")
	}

	resultWithConfigs = client.TreatmentWithConfig("key", "my_feature", nil)
	if resultWithConfigs.Treatment != "on" {
		t.Error("Treatment retrieved incorrectly")
	}
	if *resultWithConfigs.Config != "{\"desc\" : \"this applies only to ON treatment\"}" {
		t.Error("Wronf config returned")
	}

	resultWithConfigs = client.TreatmentWithConfig("key3", "other_feature", nil)
	if resultWithConfigs.Treatment != "on" {
		t.Error("Treatment retrieved incorrectly")
	}
	if resultWithConfigs.Config != nil {
		t.Error("Config should be nil")
	}

	resultTreatments := client.Treatments("only_key", []string{"my_feature", "other_feature"}, nil)
	if resultTreatments["my_feature"] != "off" {
		t.Error("Wrong Treatment result")
	}
	if resultTreatments["other_feature"] != "control" {
		t.Error("Wrong Treatment result")
	}

	resultTreatmentsWithConfig := client.TreatmentsWithConfig("only_key", []string{"my_feature", "other_feature"}, nil)
	if resultTreatmentsWithConfig["my_feature"].Treatment != "off" {
		t.Error("Wrong Treatment result")
	}
	if *resultTreatmentsWithConfig["my_feature"].Config != "{\"desc\" : \"this applies only to OFF and only for only_key. The rest will receive ON\"}" {
		t.Error("Wrong Config result")
	}
	if resultTreatmentsWithConfig["other_feature"].Treatment != "control" {
		t.Error("Wrong Treatment result")
	}
	if resultTreatmentsWithConfig["other_feature"].Config != nil {
		t.Error("Config should be nil")
	}
}

func getRedisConfWithIP(IPAddressesEnabled bool) *redisdb.PrefixedRedisClient {
	// Create prefixed client for adding Split
	prefixedClient, _ := redisdb.NewPrefixedRedisClient(&conf.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Password: "",
		Prefix:   "testPrefix",
	})

	splitStorage := redisdb.NewRedisSplitStorage(prefixedClient, logger)
	splitStorage.PutMany([]dtos.SplitDTO{*valid}, 1494593336752)

	// Set default configs to connect Client with redis
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.IPAddressesEnabled = IPAddressesEnabled
	cfg.Advanced.ImpressionListener = &ImpressionListenerTest{}
	cfg.OperationMode = "redis-consumer"
	cfg.Redis = conf.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Password: "",
		Prefix:   "testPrefix",
	}

	factory, _ := NewSplitFactory("apikey", cfg)
	client := factory.Client()

	// Deletes previous data
	prefixedClient.Del("SPLITIO.impressions", "SPLITIO.events")

	// Calls treatments to generate one valid impression
	client.Treatment("user1", "valid", nil)
	client.Track("user1", "my-traffic", "my-event", nil, nil)

	return prefixedClient
}

func deleteDataGenerated(prefixedClient *redisdb.PrefixedRedisClient) {
	// Deletes generated data
	keys, _ := prefixedClient.Keys(fmt.Sprintf("SPLITIO/go-%s/NA/latency.sdk.getTreatment.bucket.*", splitio.Version))
	keys = append(keys, "SPLITIO.impressions", "SPLITIO.events", "SPLITIO.split.valid", "SPLITIO.splits.till")
	prefixedClient.Del(keys...)
}

func TestRedisClientWithIPDisabled(t *testing.T) {
	prefixedClient := getRedisConfWithIP(false)
	// Grabs created impression
	jsonImpr, _ := prefixedClient.LRange("SPLITIO.impressions", 0, 1).Result()
	impression := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(jsonImpr[0]), &impression)
	metadata := impression["m"]

	// Checks if metadata was created with "NA" values
	if metadata["i"] != "NA" || metadata["n"] != "NA" {
		t.Error("Instance Name and Machine IP should have 'NA' values")
	}

	// Grabs created event
	jsonEvent, _ := prefixedClient.LRange("SPLITIO.events", 0, 1).Result()
	event := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(jsonEvent[0]), &event)
	metadata = event["m"]
	// Checks if metadata was created with "NA" values
	if metadata["i"] != "NA" || metadata["n"] != "NA" {
		t.Error("Instance Name and Machine IP should have 'NA' values")
	}

	deleteDataGenerated(prefixedClient)
}

func TestRedisClientWithIPEnabled(t *testing.T) {
	prefixedClient := getRedisConfWithIP(true)
	// Grabs created impression
	jsonImpr, _ := prefixedClient.LRange("SPLITIO.impressions", 0, 1).Result()
	impression := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(jsonImpr[0]), &impression)
	metadata := impression["m"]

	// Checks if metadata was created with "NA" values
	if metadata["i"] == "NA" || metadata["n"] == "NA" {
		t.Error("Instance Name and Machine IP should not have 'NA' values")
	}

	// Grabs created event
	jsonEvent, _ := prefixedClient.LRange("SPLITIO.events", 0, 1).Result()
	event := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(jsonEvent[0]), &event)
	metadata = event["m"]
	// Checks if metadata was created with "NA" values
	if metadata["i"] == "NA" || metadata["n"] == "NA" {
		t.Error("Instance Name and Machine IP should not have 'NA' values")
	}

	deleteDataGenerated(prefixedClient)
}

func getInMemoryClientWithIP(IPAddressesEnabled bool, ts *httptest.Server) SplitClient {
	// Set default configs to connect Client with redis
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.IPAddressesEnabled = IPAddressesEnabled
	cfg.Advanced.EventsURL = ts.URL
	cfg.Advanced.SdkURL = ts.URL
	cfg.LoggerConfig.LogLevel = logging.LevelDebug
	cfg.TaskPeriods.ImpressionSync = 3
	cfg.TaskPeriods.EventsSync = 3

	factory, _ := NewSplitFactory("test", cfg)
	client := factory.Client()
	client.BlockUntilReady(5)
	return *client
}

func TestClientWithIPEnabled(t *testing.T) {
	var splitsMock, _ = ioutil.ReadFile("../../testdata/splits_mock.json")
	var splitMock, _ = ioutil.ReadFile("../../testdata/split_mock.json")

	postChannel := make(chan string, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/splitChanges":
			fmt.Fprintln(w, fmt.Sprintf(string(splitsMock), splitMock))
			return
		case "/segmentChanges/___TEST___":
			w.Header().Add("Content-Encoding", "gzip")
			gzw := gzip.NewWriter(w)
			defer gzw.Close()
			fmt.Fprintln(gzw, "Hello, client")
			return
		case "/testImpressions/bulk":
		case "/events/bulk":
			if r.Header.Get("SplitSDKMachineIP") == "NA" || r.Header.Get("SplitSDKMachineName") == "NA" {
				t.Error("It should not be NA when IPAddressesEnabled has been set")
			}

			rBody, _ := ioutil.ReadAll(r.Body)
			var dataInPost []map[string]interface{}
			err := json.Unmarshal(rBody, &dataInPost)
			if err != nil {
				t.Error(err)
				return
			}

			if len(dataInPost) < 1 {
				t.Error("It should send data")
			}
			fmt.Fprintln(w, "ok")
			postChannel <- "finished"
		case "/segmentChanges":
		default:
			t.Error(r.URL.Path)
			fmt.Fprintln(w, "ok")
			return
		}
	}))
	defer ts.Close()

	client := getInMemoryClientWithIP(true, ts)
	// Calls treatments to generate one valid impression
	client.Track("user1", "my-traffic", "my-event", nil, nil)
	client.Treatment("user1", "DEMO_MURMUR2", nil)

	select {
	case <-postChannel:
		return
	case <-time.After(4 * time.Second):
		t.Error("The test couldn't send impressions to check headers")
		return
	}
}

func TestClientWithIPDisabled(t *testing.T) {
	var splitsMock, _ = ioutil.ReadFile("../../testdata/splits_mock.json")
	var splitMock, _ = ioutil.ReadFile("../../testdata/split_mock.json")

	postChannel := make(chan string, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/splitChanges":
			fmt.Fprintln(w, fmt.Sprintf(string(splitsMock), splitMock))
			return
		case "/segmentChanges/___TEST___":
			w.Header().Add("Content-Encoding", "gzip")
			gzw := gzip.NewWriter(w)
			defer gzw.Close()
			fmt.Fprintln(gzw, "Hello, client")
			return
		case "/testImpressions/bulk":
		case "/events/bulk":
			for header := range r.Header {
				if (header == "SplitSDKMachineIP") || (header == "SplitSDKMachineName") {
					t.Error("Should not insert one of SplitSDKMachineIP, SplitSDKMachineName")
				}
			}

			rBody, _ := ioutil.ReadAll(r.Body)
			var dataInPost []map[string]interface{}
			err := json.Unmarshal(rBody, &dataInPost)
			if err != nil {
				t.Error(err)
				return
			}

			if len(dataInPost) < 1 {
				t.Error("It should send data")
			}
			fmt.Fprintln(w, "ok")
			postChannel <- "finished"
		case "/segmentChanges":
		default:
			t.Error(r.URL.Path)
			fmt.Fprintln(w, "ok")
			return
		}
	}))
	defer ts.Close()

	client := getInMemoryClientWithIP(false, ts)

	// Calls treatments to generate one valid impression
	client.Track("user1", "my-traffic", "my-event", nil, nil)
	client.Treatment("user1", "DEMO_MURMUR2", nil)

	select {
	case <-postChannel:
		return
	case <-time.After(4 * time.Second):
		t.Error("The test couldn't send impressions to check headers")
		return
	}
}
