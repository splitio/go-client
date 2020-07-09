package client

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/engine/evaluator/impressionlabels"
	evaluatorMock "github.com/splitio/go-client/splitio/engine/evaluator/mocks"
	impressionlistener "github.com/splitio/go-client/splitio/impressionListener"
	redisCfg "github.com/splitio/go-split-commons/conf"
	"github.com/splitio/go-split-commons/dtos"
	authMocks "github.com/splitio/go-split-commons/service/mocks"
	"github.com/splitio/go-split-commons/storage"
	"github.com/splitio/go-split-commons/storage/mocks"
	"github.com/splitio/go-split-commons/storage/mutexmap"
	"github.com/splitio/go-split-commons/storage/mutexqueue"
	"github.com/splitio/go-split-commons/storage/redis"
	"github.com/splitio/go-split-commons/synchronizer"
	syncMock "github.com/splitio/go-split-commons/synchronizer/mocks"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
	predis "github.com/splitio/go-toolkit/redis"
)

type mockEvaluator struct{}

func (e *mockEvaluator) EvaluateFeature(
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
func (e *mockEvaluator) EvaluateFeatures(
	key string,
	bucketingKey *string,
	features []string,
	attributes map[string]interface{},
) evaluator.Results {
	results := evaluator.Results{
		Evaluations:      make(map[string]evaluator.Result),
		EvaluationTimeNs: 0,
	}
	for _, feature := range features {
		switch feature {
		case "feature":
			results.Evaluations["feature"] = evaluator.Result{
				EvaluationTimeNs:  0,
				Label:             "aLabel",
				SplitChangeNumber: 123,
				Treatment:         "TreatmentA",
			}
			break
		case "feature2":
			results.Evaluations["feature2"] = evaluator.Result{
				EvaluationTimeNs:  0,
				Label:             "bLabel",
				SplitChangeNumber: 123,
				Treatment:         "TreatmentB",
			}
			break
		case "some_feature":
			results.Evaluations["some_feature"] = evaluator.Result{
				EvaluationTimeNs:  0,
				Label:             "bLabel",
				SplitChangeNumber: 123,
				Treatment:         evaluator.Control,
			}
			break
		default:
			results.Evaluations[feature] = evaluator.Result{
				EvaluationTimeNs:  0,
				Label:             impressionlabels.SplitNotFound,
				SplitChangeNumber: 123,
				Treatment:         evaluator.Control,
			}
			break
		}
	}
	return results
}

func getFactory() SplitFactory {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	return SplitFactory{
		cfg: cfg,
		storages: sdkStorages{
			impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
			telemetry:   mutexmap.NewMMMetricsStorage(),
			events:      mocks.MockEventStorage{},
		},
		logger: logger,
	}
}

func expectedTreatment(treatment string, expectedTreatment string, t *testing.T) {
	if treatment != expectedTreatment {
		t.Error("Expected: " + expectedTreatment + " actual: " + treatment)
	}
}

func expectedTreatmentAndConfig(treatmentResult TreatmentResult, expectedTreatment string, expectedConfig string, t *testing.T) {
	if treatmentResult.Treatment != expectedTreatment {
		t.Error("Expected: " + treatmentResult.Treatment + " actual: " + expectedTreatment)
	}
	if expectedConfig != "" {
		if *treatmentResult.Config != expectedConfig {
			t.Error("Expected Config: " + *treatmentResult.Config + " actual: " + expectedConfig)
		}
	} else {
		if treatmentResult.Config != nil {
			t.Error("Config was not nil")
		}
	}

}

func TestClientGetTreatment(t *testing.T) {
	factory := getFactory()
	client := factory.Client()
	client.evaluator = &mockEvaluator{}
	factory.status.Store(sdkStatusReady)

	expectedTreatment(client.Treatment("key", "feature", nil), "TreatmentA", t)
	impressionsQueue := client.impressions.(storage.ImpressionStorage)
	impressions, _ := impressionsQueue.PopN(5000)
	impression := impressions[0]
	if impression.Label != "aLabel" {
		t.Error("Impression should have label when labelsEnabled is true")
	}

	client.factory.cfg.LabelsEnabled = false
	expectedTreatment(client.Treatment("key", "feature2", nil), "TreatmentB", t)

	impressions, _ = impressionsQueue.PopN(5000)
	impression = impressions[0]
	if impression.Label != "" {
		t.Error("Impression should have label when labelsEnabled is true")
	}
}

func TestTreatments(t *testing.T) {
	factory := getFactory()
	client := factory.Client()
	client.evaluator = &mockEvaluator{}
	factory.status.Store(sdkStatusReady)

	res := client.Treatments("user1", []string{"feature", "notFeature"}, nil)

	expectedTreatment(res["feature"], "TreatmentA", t)
	expectedTreatment(res["notFeature"], evaluator.Control, t)
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
	factory, err := NewSplitFactory("localhost", sdkConf)
	if err != nil {
		t.Error(err)
	}
	client := factory.Client()
	client.BlockUntilReady(1)

	if factory.cfg.OperationMode != "localhost" {
		t.Error("Localhost operation mode should be set when received apikey is 'localhost'")
	}

	expectedTreatment(client.Treatment("asd", "feature1", nil), "on", t)
	expectedTreatment(client.Treatment("asd", "feature2", nil), "off", t)

	if client.Track("somekey", "somett", "somee", nil, nil) != nil {
		t.Error("It should be ok")
	}

	client.Destroy()
	file.Close()
	os.Remove(file.Name())
}

func TestClientGetTreatmentConsideringValidationInputs(t *testing.T) {
	factory := getFactory()
	client := factory.Client()
	client.evaluator = &mockEvaluator{}
	factory.status.Store(sdkStatusReady)

	expectedTreatment(client.Treatment(nil, "feature", nil), evaluator.Control, t)
	expectedTreatment(client.Treatment(true, "feature", nil), evaluator.Control, t)
	expectedTreatment(client.Treatment(123, "feature", nil), "TreatmentA", t)
	expectedTreatment(client.Treatment("key", "feature", nil), "TreatmentA", t)
	expectedTreatment(client.Treatment(&Key{
		MatchingKey:  "key",
		BucketingKey: "bucketing",
	}, "feature", nil), "TreatmentA", t)
}

func TestClientPanicking(t *testing.T) {
	factory := getFactory()

	client := factory.Client()
	client.evaluator = evaluatorMock.MockEvaluator{
		EvaluateFeatureCall: func(key string, bucketingKey *string, feature string, attributes map[string]interface{}) *evaluator.Result {
			panic("Testing panicking")
		},
		EvaluateFeaturesCall: func(key string, bucketingKey *string, features []string, attributes map[string]interface{}) evaluator.Results {
			panic("Testing panicking")
		},
	}
	factory.status.Store(sdkStatusReady)

	expectedTreatment(client.Treatment("key", "some", nil), evaluator.Control, t)
}

func TestClientDestroy(t *testing.T) {
	var periodicDataRecordingStopped int64
	var periodicDataFetchingStopped int64
	logger := logging.NewLogger(nil)

	sync, _ := synchronizer.NewSynchronizerManager(
		syncMock.MockSynchronizer{
			StopPeriodicDataRecordingCall: func() {
				atomic.AddInt64(&periodicDataRecordingStopped, 1)
			},
			StopPeriodicFetchingCall: func() {
				atomic.AddInt64(&periodicDataFetchingStopped, 1)
			},
		},
		logger,
		redisCfg.AdvancedConfig{},
		authMocks.MockAuthClient{},
		mocks.MockSplitStorage{},
		make(chan int, 1),
	)

	factory := &SplitFactory{
		syncManager: sync,
		cfg:         conf.Default(),
	}
	client := SplitClient{
		logger:  logger,
		factory: factory,
	}

	time.Sleep(1 * time.Second)
	client.Destroy()
	time.Sleep(1 * time.Second)

	if client.Treatment("key", "feature", nil) != evaluator.Control {
		t.Error("Single .Treatment() call should return control")
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

	if !factory.IsDestroyed() {
		t.Error("It should be destroyed")
	}

	if periodicDataRecordingStopped != 1 {
		t.Error("It should call StopPeriodicDataRecordingCall")
	}

	if periodicDataFetchingStopped != 1 {
		t.Error("It should call StopPeriodicFetchingCall")
	}
}

// TEST CLIENT WITH IMPRESSION LISTENER //
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

func getClientForListener() SplitClient {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	impTest := &ImpressionListenerTest{}
	impresionL := impressionlistener.NewImpressionListenerWrapper(impTest, dtos.Metadata{
		SDKVersion:  "go-" + splitio.Version,
		MachineIP:   "123.123.123.123",
		MachineName: "ip-123-123-123-123",
	})

	factory := &SplitFactory{
		cfg: cfg,
		storages: sdkStorages{
			impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
			telemetry:   mutexmap.NewMMMetricsStorage(),
			events:      mocks.MockEventStorage{},
		},
		logger: logger,
		metadata: dtos.Metadata{
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

	return client
}
func TestImpressionListener(t *testing.T) {
	client := getClientForListener()
	cfg := conf.Default()

	attributes := make(map[string]interface{})
	attributes["One"] = "test"

	expectedTreatment(client.Treatment("user1", "feature", attributes), "TreatmentA", t)
	expectedVersion := "go-" + splitio.Version

	if !compareListener(ilResult["feature"].(map[string]interface{}), "feature", "user1", "aLabel", "TreatmentA", int64(123), "", "test", cfg.InstanceName, expectedVersion) {
		t.Error("Impression should match")
	}
	ilResult = make(map[string]interface{})

	delete(ilResult, "feature")
}

func TestImpressionListenerForTreatments(t *testing.T) {
	client := getClientForListener()
	cfg := conf.Default()

	attributes := make(map[string]interface{})
	attributes["One"] = "test"

	res := client.Treatments("user1", []string{"feature", "feature2"}, attributes)

	expectedTreatment(res["feature"], "TreatmentA", t)
	expectedTreatment(res["feature2"], "TreatmentB", t)

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

// TEST BLOCK UNTIL READY //
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

	expectedTreatment(client.Treatment("something", "something", attributes), evaluator.Control, t)

	features := []string{"something"}
	result := client.Treatments("something", features, nil)
	expectedTreatment(result["something"], evaluator.Control, t)

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

	expectedTreatment(client.Treatment("asd", "feature1", nil), "on", t)

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

	expectedTreatment(client.Treatment("not_ready", "not_ready", attributes), evaluator.Control, t)
	if !compareListener(ilResult["not_ready"].(map[string]interface{}), "not_ready", "not_ready", "not ready", "control", int64(0), "", "test", sdkConf.InstanceName, expectedVersion) {
		t.Error("Impression should match")

	}
	ilResult = make(map[string]interface{})

	expectedTreatment(client.Treatment("something", "something", attributes), evaluator.Control, t)

	features := []string{"something"}
	result := client.Treatments("something", features, nil)
	expectedTreatment(result["something"], evaluator.Control, t)

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

	expectedTreatment(client.Treatment("not_ready2", "not_ready2", attributes), evaluator.Control, t)
	if !compareListener(ilResult["not_ready2"].(map[string]interface{}), "not_ready2", "not_ready2", "not ready", "control", int64(0), "", "test", sdkConf.InstanceName, expectedVersion) {
		t.Error("Impression should match")
	}

	result := client.Treatments("not_ready3", []string{"not_ready3"}, attributes)
	expectedTreatment(result["not_ready3"], evaluator.Control, t)
	if !compareListener(ilResult["not_ready3"].(map[string]interface{}), "not_ready3", "not_ready3", "not ready", "control", int64(0), "", "test", sdkConf.InstanceName, expectedVersion) {
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

func isInvalidImpression(client SplitClient, key string, feature string, treatment string) bool {
	impressionsQueue := client.impressions.(storage.ImpressionStorage)
	impressions, _ := impressionsQueue.PopN(5000)
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
		mocks.MockSplitStorage{
			SplitCall: func(splitName string) *dtos.SplitDTO {
				switch splitName {
				default:
				case "valid":
					return valid
				case "killed":
					return killed
				}
				return nil
			},
			FetchManyCall: func(splitNames []string) map[string]*dtos.SplitDTO {
				splits := make(map[string]*dtos.SplitDTO)
				for _, feature := range splitNames {
					switch feature {
					default:
					case "valid":
						splits[feature] = valid
						break
					case "killed":
						splits["killed"] = killed
						break
					}
				}
				return splits
			},
		},
		mocks.MockSegmentStorage{
			KeysCall: func(segmentName string) *set.ThreadUnsafeSet {
				switch segmentName {
				default:
				case "employees":
					return set.NewSet("user1")
				}
				return nil
			},
		},
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
	expectedTreatment(client.Treatment("user1", "valid", nil), "on", t)
	if isInvalidImpression(client, "user1", "valid", "on") {
		t.Error("Wrong impression saved")
	}

	expectedTreatment(client.Treatment("invalid", "valid", nil), "off", t)
	if isInvalidImpression(client, "invalid", "valid", "off") {
		t.Error("Wrong impression saved")
	}

	expectedTreatment(client.Treatment("invalid", "invalid", nil), "control", t)
	if client.Treatment("invalid", "invalid", nil) != "control" {
		t.Error("Unexpected Treatment Result")
	}

	expectedTreatment(client.Treatment("invalid", "killed", nil), "defTreatment", t)
	if isInvalidImpression(client, "invalid", "killed", "defTreatment") {
		t.Error("Wrong impression saved")
	}

	// Assertion Treatments
	treatments := client.Treatments("user1", []string{"valid", "invalid", "killed"}, nil)
	expectedTreatment(treatments["invalid"], "control", t)
	expectedTreatment(treatments["killed"], "defTreatment", t)
	expectedTreatment(treatments["valid"], "on", t)
	client.impressions.(storage.ImpressionStorage).PopN(cfg.Advanced.ImpressionsBulkSize)

	// Assertion TreatmentWithConfig
	expectedTreatmentAndConfig(client.TreatmentWithConfig("user1", "valid", nil), "on", "{\"color\": \"blue\",\"size\": 13}", t)
	if isInvalidImpression(client, "user1", "valid", "on") {
		t.Error("Wrong impression saved")
	}

	expectedTreatmentAndConfig(client.TreatmentWithConfig("invalid", "valid", nil), "off", "", t)
	if isInvalidImpression(client, "invalid", "valid", "off") {
		t.Error("Wrong impression saved")
	}

	expectedTreatmentAndConfig(client.TreatmentWithConfig("invalid", "invalid", nil), "control", "", t)
	expectedTreatmentAndConfig(client.TreatmentWithConfig("invalid", "killed", nil), "defTreatment", "{\"color\": \"orange\",\"size\": 15}", t)
	if isInvalidImpression(client, "invalid", "killed", "defTreatment") {
		t.Error("Wrong impression saved")
	}

	// Assertion TreatmentsWithConfig
	treatmentsWithConfigs := client.TreatmentsWithConfig("user1", []string{"valid", "invalid", "killed"}, nil)
	expectedTreatmentAndConfig(treatmentsWithConfigs["invalid"], "control", "", t)
	expectedTreatmentAndConfig(treatmentsWithConfigs["killed"], "defTreatment", "{\"color\": \"orange\",\"size\": 15}", t)
	expectedTreatmentAndConfig(treatmentsWithConfigs["valid"], "on", "{\"color\": \"blue\",\"size\": 13}", t)
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

	expectedTreatment(client.Treatment("only_key", "my_feature", nil), "off", t)
	expectedTreatment(client.Treatment("invalid_key", "my_feature", nil), "control", t)
	expectedTreatment(client.Treatment("key", "my_feature", nil), "on", t)
	expectedTreatment(client.Treatment("key2", "other_feature", nil), "on", t)
	expectedTreatment(client.Treatment("test", "other_feature_2", nil), "on", t)
	expectedTreatment(client.Treatment("key", "other_feature_3", nil), "off", t)
	expectedTreatment(client.Treatment("key_whitelist", "other_feature_3", nil), "on", t)

	expectedTreatmentAndConfig(client.TreatmentWithConfig("only_key", "my_feature", nil), "off", "{\"desc\" : \"this applies only to OFF and only for only_key. The rest will receive ON\"}", t)
	expectedTreatmentAndConfig(client.TreatmentWithConfig("key", "my_feature", nil), "on", "{\"desc\" : \"this applies only to ON treatment\"}", t)
	expectedTreatmentAndConfig(client.TreatmentWithConfig("key3", "other_feature", nil), "on", "", t)

	resultTreatments := client.Treatments("only_key", []string{"my_feature", "other_feature"}, nil)
	expectedTreatment(resultTreatments["my_feature"], "off", t)
	expectedTreatment(resultTreatments["other_feature"], "control", t)

	resultTreatmentsWithConfig := client.TreatmentsWithConfig("only_key", []string{"my_feature", "other_feature"}, nil)
	expectedTreatmentAndConfig(resultTreatmentsWithConfig["my_feature"], "off", "{\"desc\" : \"this applies only to OFF and only for only_key. The rest will receive ON\"}", t)
	expectedTreatmentAndConfig(resultTreatmentsWithConfig["other_feature"], "control", "", t)
}

func getRedisConfWithIP(IPAddressesEnabled bool) *predis.PrefixedRedisClient {
	// Create prefixed client for adding Split
	prefixedClient, _ := redis.NewRedisClient(&redisCfg.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Password: "",
		Prefix:   "testPrefix",
	}, logging.NewLogger(&logging.LoggerOptions{}))

	raw, err := json.Marshal(*valid)
	if err != nil {
		return nil
	}
	prefixedClient.Set("SPLITIO.split.valid", raw, 0)
	prefixedClient.Set("SPLITIO.splits.till", 1494593336752, 0)

	// Set default configs to connect Client with redis
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.IPAddressesEnabled = IPAddressesEnabled
	cfg.Advanced.ImpressionListener = &ImpressionListenerTest{}
	cfg.OperationMode = "redis-consumer"
	cfg.Redis = redisCfg.RedisConfig{
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

func deleteDataGenerated(prefixedClient *predis.PrefixedRedisClient) {
	// Deletes generated data
	keys, _ := prefixedClient.Keys(fmt.Sprintf("SPLITIO/go-%s/*/latency.sdk.getTreatment.bucket.*", splitio.Version))
	keys = append(keys, "SPLITIO.impressions", "SPLITIO.events", "SPLITIO.split.valid", "SPLITIO.splits.till")
	prefixedClient.Del(keys...)
}

func TestRedisClientWithIPDisabled(t *testing.T) {
	prefixedClient := getRedisConfWithIP(false)
	// Grabs created impression
	resImpression, _ := prefixedClient.LRange("SPLITIO.impressions", 0, 1)
	impression := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(resImpression[0]), &impression)
	metadata := impression["m"]
	// Checks if metadata was created with "NA" values
	if metadata["i"] != "NA" || metadata["n"] != "NA" {
		t.Error("Instance Name and Machine IP should have 'NA' values")
	}
	listenerData, _ := ilResult["valid"].(map[string]interface{})
	if listenerData["InstanceName"] != "NA" {
		t.Error("InstanceName should be 'NA")
	}

	// Grabs created event
	resEvent, _ := prefixedClient.LRange("SPLITIO.events", 0, 1)
	event := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(resEvent[0]), &event)
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
	resImpression, _ := prefixedClient.LRange("SPLITIO.impressions", 0, 1)
	impression := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(resImpression[0]), &impression)
	metadata := impression["m"]
	// Checks if metadata was created with "NA" values
	if metadata["i"] == "NA" || metadata["n"] == "NA" {
		t.Error("Instance Name and Machine IP should not have 'NA' values")
	}
	listenerData, _ := ilResult["valid"].(map[string]interface{})
	if listenerData["InstanceName"] == "NA" {
		t.Error("InstanceName should not be 'NA")
	}

	// Grabs created event
	resEvent, _ := prefixedClient.LRange("SPLITIO.events", 0, 1)
	event := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(resEvent[0]), &event)
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
	cfg.Advanced.ImpressionListener = &ImpressionListenerTest{}
	cfg.LoggerConfig.LogLevel = logging.LevelDebug
	cfg.TaskPeriods.ImpressionSync = 3
	cfg.TaskPeriods.EventsSync = 3

	factory, _ := NewSplitFactory("test", cfg)
	client := factory.Client()
	client.BlockUntilReady(2)
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
			fmt.Fprintln(w, "ok")
			return
		}
	}))
	defer ts.Close()

	client := getInMemoryClientWithIP(true, ts)
	// Calls treatments to generate one valid impression
	client.Track("user1", "my-traffic", "my-event", nil, nil)
	client.Treatment("user1", "DEMO_MURMUR2", nil)

	listenerData, _ := ilResult["DEMO_MURMUR2"].(map[string]interface{})
	if listenerData["InstanceName"] == "NA" {
		t.Error("InstanceName should not be 'NA")
	}

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
			fmt.Fprintln(w, "ok")
			return
		}
	}))
	defer ts.Close()

	client := getInMemoryClientWithIP(false, ts)

	// Calls treatments to generate one valid impression
	client.Track("user1", "my-traffic", "my-event", nil, nil)
	client.Treatment("user1", "DEMO_MURMUR2", nil)

	listenerData, _ := ilResult["DEMO_MURMUR2"].(map[string]interface{})
	if listenerData["InstanceName"] != "NA" {
		t.Error("InstanceName should be 'NA")
	}

	select {
	case <-postChannel:
		return
	case <-time.After(4 * time.Second):
		t.Error("The test couldn't send impressions to check headers")
		return
	}
}
