package client

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/splitio/go-client/v6/splitio"
	"github.com/splitio/go-client/v6/splitio/conf"
	impressionlistener "github.com/splitio/go-client/v6/splitio/impressionListener"

	commonsCfg "github.com/splitio/go-split-commons/v6/conf"
	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/go-split-commons/v6/engine/evaluator"
	"github.com/splitio/go-split-commons/v6/engine/evaluator/impressionlabels"
	evaluatorMock "github.com/splitio/go-split-commons/v6/engine/evaluator/mocks"
	"github.com/splitio/go-split-commons/v6/healthcheck/application"
	"github.com/splitio/go-split-commons/v6/provisional"
	"github.com/splitio/go-split-commons/v6/provisional/strategy"
	authMocks "github.com/splitio/go-split-commons/v6/service/mocks"
	"github.com/splitio/go-split-commons/v6/storage"
	"github.com/splitio/go-split-commons/v6/storage/inmemory"
	"github.com/splitio/go-split-commons/v6/storage/inmemory/mutexqueue"
	"github.com/splitio/go-split-commons/v6/storage/mocks"
	"github.com/splitio/go-split-commons/v6/storage/redis"
	"github.com/splitio/go-split-commons/v6/synchronizer"
	syncMock "github.com/splitio/go-split-commons/v6/synchronizer/mocks"
	"github.com/splitio/go-split-commons/v6/telemetry"
	"github.com/splitio/go-split-commons/v6/util"

	"github.com/splitio/go-toolkit/v5/datastructures/set"
	"github.com/splitio/go-toolkit/v5/logging"
	predis "github.com/splitio/go-toolkit/v5/redis"
)

type mockEvaluator struct{}

// EvaluateFeatureByFlagSets implements evaluator.Interface.
func (*mockEvaluator) EvaluateFeatureByFlagSets(key string, bucketingKey *string, flagSets []string, attributes map[string]interface{}) evaluator.Results {
	panic("unimplemented")
}

func (e *mockEvaluator) EvaluateFeature(
	key string,
	bucketingKey *string,
	feature string,
	attributes map[string]interface{},
) *evaluator.Result {
	switch feature {
	case "feature":
		return &evaluator.Result{
			EvaluationTime:    0,
			Label:             "aLabel",
			SplitChangeNumber: 123,
			Treatment:         "TreatmentA",
		}
	case "feature2":
		return &evaluator.Result{
			EvaluationTime:    0,
			Label:             "bLabel",
			SplitChangeNumber: 123,
			Treatment:         "TreatmentB",
		}
	case "some_feature":
		return &evaluator.Result{
			EvaluationTime:    0,
			Label:             "bLabel",
			SplitChangeNumber: 123,
			Treatment:         evaluator.Control,
		}
	default:
		return &evaluator.Result{
			EvaluationTime:    0,
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
		Evaluations:    make(map[string]evaluator.Result),
		EvaluationTime: 0,
	}
	for _, feature := range features {
		switch feature {
		case "feature":
			results.Evaluations["feature"] = evaluator.Result{
				EvaluationTime:    0,
				Label:             "aLabel",
				SplitChangeNumber: 123,
				Treatment:         "TreatmentA",
			}
		case "feature2":
			results.Evaluations["feature2"] = evaluator.Result{
				EvaluationTime:    0,
				Label:             "bLabel",
				SplitChangeNumber: 123,
				Treatment:         "TreatmentB",
			}
		case "some_feature":
			results.Evaluations["some_feature"] = evaluator.Result{
				EvaluationTime:    0,
				Label:             "bLabel",
				SplitChangeNumber: 123,
				Treatment:         evaluator.Control,
			}
		default:
			results.Evaluations[feature] = evaluator.Result{
				EvaluationTime:    0,
				Label:             impressionlabels.SplitNotFound,
				SplitChangeNumber: 123,
				Treatment:         evaluator.Control,
			}
		}
	}
	return results
}

func getFactory() SplitFactory {
	telemetryStorage, _ := inmemory.NewTelemetryStorage()
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	impressionObserver, _ := strategy.NewImpressionObserver(500)
	impressionsCounter := strategy.NewImpressionsCounter()
	impressionsStrategy := strategy.NewOptimizedImpl(impressionObserver, impressionsCounter, telemetryStorage, false)
	impressionManager := provisional.NewImpressionManager(impressionsStrategy)

	return SplitFactory{
		cfg: cfg,
		storages: sdkStorages{
			impressions:         mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger, telemetryStorage),
			events:              mocks.MockEventStorage{},
			initTelemetry:       telemetryStorage,
			runtimeTelemetry:    telemetryStorage,
			evaluationTelemetry: telemetryStorage,
		},
		impressionManager: impressionManager,
		logger:            logger,
	}
}

func getFactoryByFlagSets() SplitFactory {
	telemetryStorage, _ := inmemory.NewTelemetryStorage()
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.Advanced.FlagSetsFilter = []string{"set1", "set2"}
	logger := logging.NewLogger(nil)

	impressionObserver, _ := strategy.NewImpressionObserver(500)
	impressionsCounter := strategy.NewImpressionsCounter()
	impressionsStrategy := strategy.NewOptimizedImpl(impressionObserver, impressionsCounter, telemetryStorage, false)
	impressionManager := provisional.NewImpressionManager(impressionsStrategy)

	return SplitFactory{
		cfg: cfg,
		storages: sdkStorages{
			impressions:         mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger, telemetryStorage),
			events:              mocks.MockEventStorage{},
			initTelemetry:       telemetryStorage,
			runtimeTelemetry:    telemetryStorage,
			evaluationTelemetry: telemetryStorage,
		},
		impressionManager: impressionManager,
		logger:            logger,
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

func TestClientGetTreatmentByFlagSet(t *testing.T) {
	factory := getFactoryByFlagSets()
	client := factory.Client()
	client.evaluator = evaluatorMock.MockEvaluator{
		EvaluateFeatureByFlagSetsCall: func(key string, bucketingKey *string, flagSets []string, attributes map[string]interface{}) evaluator.Results {
			results := evaluator.Results{
				Evaluations:    make(map[string]evaluator.Result),
				EvaluationTime: 0,
			}
			for _, flagSet := range flagSets {
				switch flagSet {
				case "set1":
					results.Evaluations["feature"] = evaluator.Result{
						EvaluationTime:    0,
						Label:             "aLabel",
						SplitChangeNumber: 123,
						Treatment:         "TreatmentA",
					}
				default:
					t.Error("Should be set1 or set2")
				}
			}
			return results
		},
	}
	factory.status.Store(sdkStatusReady)

	res := client.TreatmentsByFlagSet("user1", "set1", nil)

	expectedTreatment(res["feature"], "TreatmentA", t)
}

func TestClientGetTreatmentByFlagSets(t *testing.T) {
	factory := getFactory()
	client := factory.Client()
	client.evaluator = evaluatorMock.MockEvaluator{
		EvaluateFeatureByFlagSetsCall: func(key string, bucketingKey *string, flagSets []string, attributes map[string]interface{}) evaluator.Results {
			results := evaluator.Results{
				Evaluations:    make(map[string]evaluator.Result),
				EvaluationTime: 0,
			}
			for _, flagSet := range flagSets {
				switch flagSet {
				case "set1":
					results.Evaluations["feature"] = evaluator.Result{
						EvaluationTime:    0,
						Label:             "aLabel",
						SplitChangeNumber: 123,
						Treatment:         "TreatmentA",
					}
				case "set2":
					results.Evaluations["feature2"] = evaluator.Result{
						EvaluationTime:    0,
						Label:             "bLabel",
						SplitChangeNumber: 123,
						Treatment:         "TreatmentB",
					}
				default:
					t.Error("Should be set1 or set2")
				}
			}
			return results
		},
	}
	factory.status.Store(sdkStatusReady)

	res := client.TreatmentsByFlagSets("user1", []string{"set1", "set2"}, nil)

	expectedTreatment(res["feature"], "TreatmentA", t)
	expectedTreatment(res["feature2"], "TreatmentB", t)
}

func TestClientGetTreatmentWithConfigByFlagSet(t *testing.T) {
	factory := getFactory()
	client := factory.Client()
	client.evaluator = evaluatorMock.MockEvaluator{
		EvaluateFeatureByFlagSetsCall: func(key string, bucketingKey *string, flagSets []string, attributes map[string]interface{}) evaluator.Results {
			results := evaluator.Results{
				Evaluations:    make(map[string]evaluator.Result),
				EvaluationTime: 0,
			}
			for _, flagSet := range flagSets {
				switch flagSet {
				case "set1":
					results.Evaluations["feature"] = evaluator.Result{
						EvaluationTime:    0,
						Label:             "aLabel",
						SplitChangeNumber: 123,
						Treatment:         "TreatmentA",
					}
				default:
					t.Error("Should be set1 or set2")
				}
			}
			return results
		},
	}
	factory.status.Store(sdkStatusReady)

	res := client.TreatmentsWithConfigByFlagSet("user1", "set1", nil)

	expectedTreatment(res["feature"].Treatment, "TreatmentA", t)
}

func TestClientGetTreatmentWithConfigByFlagSets(t *testing.T) {
	factory := getFactory()
	client := factory.Client()
	client.evaluator = evaluatorMock.MockEvaluator{
		EvaluateFeatureByFlagSetsCall: func(key string, bucketingKey *string, flagSets []string, attributes map[string]interface{}) evaluator.Results {
			results := evaluator.Results{
				Evaluations:    make(map[string]evaluator.Result),
				EvaluationTime: 0,
			}
			for _, flagSet := range flagSets {
				switch flagSet {
				case "set1":
					results.Evaluations["feature"] = evaluator.Result{
						EvaluationTime:    0,
						Label:             "aLabel",
						SplitChangeNumber: 123,
						Treatment:         "TreatmentA",
					}
				case "set2":
					results.Evaluations["feature2"] = evaluator.Result{
						EvaluationTime:    0,
						Label:             "bLabel",
						SplitChangeNumber: 123,
						Treatment:         "TreatmentB",
					}
				default:
					t.Error("Should be set1 or set2")
				}
			}
			return results
		},
	}
	factory.status.Store(sdkStatusReady)

	res := client.TreatmentsWithConfigByFlagSets("user1", []string{"set1", "set2"}, nil)

	expectedTreatment(res["feature"].Treatment, "TreatmentA", t)
	expectedTreatment(res["feature2"].Treatment, "TreatmentB", t)
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
	factory, err := NewSplitFactory(conf.Localhost, sdkConf)
	if err != nil {
		t.Error(err)
	}
	client := factory.Client()
	client.BlockUntilReady(1)

	if factory.cfg.OperationMode != conf.Localhost {
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
	call := 0
	telemetryMockedStorage := mocks.MockTelemetryStorage{
		RecordExceptionCall: func(method string) {
			switch call {
			case 0:
				if method != telemetry.Treatment {
					t.Error("Should be Treatment")
				}
			case 1:
				if method != telemetry.Treatments {
					t.Error("Should be Treatments")
				}
			case 2:
				if method != telemetry.TreatmentWithConfig {
					t.Error("Should be TreatmentWithConfig")
				}
			case 3:
				if method != telemetry.TreatmentsWithConfig {
					t.Error("Should be TreatmentsWithConfig")
				}
			case 4:
				if method != telemetry.Track {
					t.Error("Should be Track")
				}
			}

			call++
		},
	}
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	impressionObserver, _ := strategy.NewImpressionObserver(500)
	impressionsCounter := strategy.NewImpressionsCounter()
	impressionsStrategy := strategy.NewOptimizedImpl(impressionObserver, impressionsCounter, telemetryMockedStorage, false)
	impressionManager := provisional.NewImpressionManager(impressionsStrategy)

	factory := SplitFactory{
		cfg: cfg,
		storages: sdkStorages{
			impressions:         mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger, telemetryMockedStorage),
			events:              mocks.MockEventStorage{},
			initTelemetry:       telemetryMockedStorage,
			runtimeTelemetry:    telemetryMockedStorage,
			evaluationTelemetry: telemetryMockedStorage,
		},
		impressionManager: impressionManager,
		logger:            logger,
	}

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
	expectedTreatment(client.Treatments("key", []string{"some"}, nil)["some"], evaluator.Control, t)
	expectedTreatment(client.TreatmentWithConfig("key", "some", nil).Treatment, evaluator.Control, t)
	expectedTreatment(client.TreatmentsWithConfig("key", []string{"some"}, nil)["some"].Treatment, evaluator.Control, t)

	err := client.Track("some", "some", "some", nil, nil)
	if err == nil || err.Error() != "Track is panicking. Please check logs" {
		t.Error("It should panic")
	}
}

func TestClientDestroy(t *testing.T) {
	var periodicDataRecordingStopped int64
	var periodicDataFetchingStopped int64
	logger := logging.NewLogger(nil)
	telemetryStorage, _ := inmemory.NewTelemetryStorage()

	sync, _ := synchronizer.NewSynchronizerManager(
		&syncMock.MockSynchronizer{
			SyncAllCall:                    func() error { return nil },
			StartPeriodicDataRecordingCall: func() {},
			StartPeriodicFetchingCall:      func() {},
			StopPeriodicDataRecordingCall:  func() { atomic.AddInt64(&periodicDataRecordingStopped, 1) },
			StopPeriodicFetchingCall:       func() { atomic.AddInt64(&periodicDataFetchingStopped, 1) },
			RefreshRatesCall:               func() (time.Duration, time.Duration) { return time.Minute, time.Minute },
		},
		logger,
		commonsCfg.AdvancedConfig{StreamingEnabled: false},
		authMocks.MockAuthClient{},
		mocks.MockSplitStorage{},
		make(chan int, 1),
		telemetryStorage,
		dtos.Metadata{},
		nil,
		&application.Dummy{},
	)
	sync.Start()

	factory := &SplitFactory{
		syncManager: sync,
		cfg:         conf.Default(),
		storages: sdkStorages{
			initTelemetry:       telemetryStorage,
			runtimeTelemetry:    telemetryStorage,
			evaluationTelemetry: telemetryStorage,
		},
	}
	client := SplitClient{
		logger:              logger,
		factory:             factory,
		runtimeTelemetry:    telemetryStorage,
		initTelemetry:       telemetryStorage,
		evaluationTelemetry: telemetryStorage,
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
	ilTest["Pt"] = data.Impression.Pt

	ilResult[data.Impression.FeatureName] = ilTest
}

func compareListener(ilTest map[string]interface{}, f string, k string, l string, t string, c int64, b string, a string, i string, v string) bool {
	if ilTest["FeatureName"] != f || ilTest["KeyName"] != k || ilTest["Label"] != l || ilTest["Treatment"] != t || ilTest["ChangeNumber"] != c || ilTest["BucketingKey"] != b {
		return false
	}
	if ilTest["Version"] != v {
		return false
	}
	if ilTest["InstanceName"] != i {
		return false
	}
	attr1, _ := ilTest["Attributes"].(map[string]interface{})
	return attr1["One"] == a
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
	telemetryMockedStorage := mocks.MockTelemetryStorage{
		RecordImpressionsStatsCall: func(dataType int, count int64) {},
		RecordLatencyCall:          func(method string, latency time.Duration) {},
	}
	impressionStorage := mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger, telemetryMockedStorage)

	impressionObserver, _ := strategy.NewImpressionObserver(500)
	impressionsCounter := strategy.NewImpressionsCounter()
	impressionsStrategy := strategy.NewOptimizedImpl(impressionObserver, impressionsCounter, telemetryMockedStorage, true)
	impressionManager := provisional.NewImpressionManager(impressionsStrategy)

	factory := &SplitFactory{
		cfg: cfg,
		storages: sdkStorages{
			impressions:         impressionStorage,
			events:              mocks.MockEventStorage{},
			evaluationTelemetry: telemetryMockedStorage,
		},
		logger: logger,
		metadata: dtos.Metadata{
			SDKVersion: "go-" + splitio.Version,
		},
		impressionManager: impressionManager,
	}

	client := SplitClient{
		evaluator:           &mockEvaluator{},
		impressions:         impressionStorage,
		logger:              logger,
		impressionListener:  impresionL,
		factory:             factory,
		impressionManager:   impressionManager,
		evaluationTelemetry: telemetryMockedStorage,
	}

	factory.status.Store(sdkStatusReady)

	return client
}
func TestImpressionListener(t *testing.T) {
	client := getClientForListener()

	attributes := make(map[string]interface{})
	attributes["One"] = "test"

	expectedTreatment(client.Treatment("user1", "feature", attributes), "TreatmentA", t)
	expectedVersion := "go-" + splitio.Version

	if !compareListener(ilResult["feature"].(map[string]interface{}), "feature", "user1", "aLabel", "TreatmentA", int64(123), "", "test", "ip-123-123-123-123", expectedVersion) {
		t.Error("Impression should match")
	}
	ilResult = make(map[string]interface{})

	delete(ilResult, "feature")
}

func TestImpressionListenerForTreatments(t *testing.T) {
	client := getClientForListener()

	attributes := make(map[string]interface{})
	attributes["One"] = "test"

	res := client.Treatments("user1", []string{"feature", "feature2"}, attributes)

	expectedTreatment(res["feature"], "TreatmentA", t)
	expectedTreatment(res["feature2"], "TreatmentB", t)

	if len(ilResult) != 2 {
		t.Error("Error on ImpressionListener")
	}

	expectedVersion := "go-" + splitio.Version

	if !compareListener(ilResult["feature"].(map[string]interface{}), "feature", "user1", "aLabel", "TreatmentA", int64(123), "", "test", "ip-123-123-123-123", expectedVersion) {
		t.Error("Impression should match")
	}

	if !compareListener(ilResult["feature2"].(map[string]interface{}), "feature2", "user1", "bLabel", "TreatmentB", int64(123), "", "test", "ip-123-123-123-123", expectedVersion) {
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

	factory, _ := NewSplitFactory(conf.Localhost, sdkConf)

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

	factory, _ := NewSplitFactory(conf.Localhost, sdkConf)

	client := factory.Client()
	manager := factory.Manager()

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

	factory, _ := NewSplitFactory(conf.Localhost, sdkConf)

	client := factory.Client()
	manager := factory.Manager()

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
	sdkConf.OperationMode = conf.RedisConsumer

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

	prefixedClient, _ := redis.NewRedisClient(&commonsCfg.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Prefix:   "",
	}, logging.NewLogger(&logging.LoggerOptions{}))
	deleteDataGenerated(prefixedClient)
}

func TestBlockUntilReadyInMemoryError(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(404)
	}))
	defer ts.Close()

	sdkConf := conf.Default()
	sdkConf.Advanced.SdkURL = ts.URL
	sdkConf.Advanced.EventsURL = ts.URL
	sdkConf.Advanced.AuthServiceURL = ts.URL
	sdkConf.Advanced.TelemetryServiceURL = ts.URL
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
	if !compareListener(
		ilResult["not_ready"].(map[string]interface{}),
		"not_ready",
		"not_ready",
		"not ready",
		"control",
		int64(0),
		"",
		"test",
		sdkConf.InstanceName, expectedVersion,
	) {
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

	err = client.BlockUntilReady(5)
	if err == nil {
		t.Error("It should return error")
	}

	if err != nil && err.Error() != "SDK Initialization failed" {
		t.Error("Wrong error. Got: ", err.Error())
	}
}

func TestBlockUntilReadyInMemoryOk(t *testing.T) {
	metricsInitCalled := 0
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
					Matchers: []dtos.MatcherDTO{{MatcherType: "ALL_KEYS"}},
				},
				Partitions: []dtos.PartitionDTO{{Size: 100, Treatment: "on"}},
			},
		},
	}
	mockedSplit2 := dtos.SplitDTO{Name: "split2", Killed: true, Status: "ACTIVE"}
	mockedSplit3 := dtos.SplitDTO{Name: "split3", Killed: true, Status: "INACTIVE"}

	sdkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		if r.URL.Path != "/splitChanges" || r.Method != "GET" {
			t.Error("Invalid request. Should be GET to /splitChanges")
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
	defer sdkServer.Close()

	eventsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer eventsServer.Close()

	telemetryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/metrics/config":
			metricsInitCalled++
			rBody, _ := ioutil.ReadAll(r.Body)
			var dataInPost dtos.Config
			err := json.Unmarshal(rBody, &dataInPost)
			if err != nil {
				t.Error(err)
				return
			}

			if dataInPost.NonReadyUsages != 4 {
				t.Error("It should call 4 methods in non ready state")
			}
			if dataInPost.BurTimeouts != 2 {
				t.Error("It should excedeed two times")
			}
			if dataInPost.OperationMode != telemetry.Standalone {
				t.Error("It should be Standalone")
			}
			if dataInPost.Storage != telemetry.Memory {
				t.Error("It should initiate in memory mode")
			}
			if dataInPost.TimeUntilReady < 3000 {
				t.Error("It should took more than timers set in test")
			}
			if !dataInPost.ImpressionsListenerEnabled {
				t.Error("It should have impression listener")
			}
		}
		fmt.Fprintln(w, "ok")
	}))
	defer telemetryServer.Close()

	sdkConf := conf.Default()
	sdkConf.LoggerConfig.StandardLoggerFlags = log.Llongfile
	sdkConf.Advanced.EventsURL = eventsServer.URL
	sdkConf.Advanced.SdkURL = sdkServer.URL
	sdkConf.Advanced.TelemetryServiceURL = telemetryServer.URL
	sdkConf.Advanced.StreamingEnabled = false
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

	err = manager.BlockUntilReady(1)
	if err == nil {
		t.Error("It should return error")
	}

	expected2 = "SDK Initialization: time of 1 exceeded"
	if err == nil || err.Error() != expected2 {
		t.Error("Wrong message error. Got:", err.Error())
	}

	err = client.BlockUntilReady(2)
	if err != nil {
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

	if metricsInitCalled != 1 {
		t.Error("It should send init data")
	}
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
	cfg.ImpressionsMode = commonsCfg.ImpressionsModeDebug
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
					case "killed":
						splits["killed"] = killed
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
			SegmentContainsKeyCall: func(segmentName, key string) (bool, error) {
				if segmentName != "employees" {
					return false, fmt.Errorf("segment not found")
				}

				if key == "user1" {
					return true, nil
				}
				return false, nil
			},
		},
		nil,
		logger,
	)

	mockedTelemetryStorage := mocks.MockTelemetryStorage{
		RecordImpressionsStatsCall: func(dataType int, count int64) {},
		RecordLatencyCall:          func(method string, latency time.Duration) {},
	}

	impressionObserver, _ := strategy.NewImpressionObserver(500)
	impressionsStrategy := strategy.NewDebugImpl(impressionObserver, true)
	impressionManager := provisional.NewImpressionManager(impressionsStrategy)

	factory := &SplitFactory{cfg: cfg, impressionManager: impressionManager}
	client := SplitClient{
		evaluator:           evaluator,
		impressions:         mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger, mockedTelemetryStorage),
		logger:              logger,
		validator:           inputValidation{logger: logger},
		factory:             factory,
		impressionManager:   impressionManager,
		evaluationTelemetry: mockedTelemetryStorage,
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
	factory, _ := NewSplitFactory(conf.Localhost, sdkConf)
	client := factory.Client()
	manager := factory.Manager()

	_ = client.BlockUntilReady(5)

	if !client.isReady() {
		t.Error("Localhost should be ready")
	}

	if client.factory.cfg.OperationMode != conf.Localhost {
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

func TestLocalhostModeJSON(t *testing.T) {
	sdkConf := conf.Default()
	sdkConf.SplitFile = "../../testdata/splits.json"
	sdkConf.SegmentDirectory = "../../testdata/segments"
	factory, _ := NewSplitFactory(conf.Localhost, sdkConf)
	client := factory.Client()
	manager := factory.Manager()

	_ = client.BlockUntilReady(5)

	if !client.isReady() {
		t.Error("Localhost should be ready")
	}

	if client.factory.cfg.OperationMode != conf.Localhost {
		t.Error("Localhost operation mode should be set when received apikey is 'localhost'")
	}

	if len(manager.Splits()) != 4 {
		t.Error("Error grabbing splits for localhost mode")
	}

	expectedTreatment(client.Treatment("key", "non_existent_feature", nil), "control", t)
	expectedTreatment(client.Treatment("key", "feature_flag_1", nil), "on", t)
	expectedTreatment(client.Treatment("example1", "feature_flag_2", nil), "some_treatment", t)
	expectedTreatment(client.Treatment("key", "feature_flag_2", nil), "on", t)
	expectedTreatment(client.Treatment("key", "feature_flag_3", nil), "off", t)
	expectedTreatment(client.Treatment("key", "feature_flag_4", nil), "control", t)

	expectedTreatmentAndConfig(client.TreatmentWithConfig("key", "feature_flag_2", nil), "on", "{\"color\":\"red\"}", t)
	expectedTreatmentAndConfig(client.TreatmentWithConfig("example1", "feature_flag_2", nil), "some_treatment", "{\"color\":\"white\"}", t)
	expectedTreatmentAndConfig(client.TreatmentWithConfig("key", "feature_flag_3", nil), "off", "", t)

	resultTreatments := client.Treatments("example1", []string{"non_existent_feature", "feature_flag_1", "feature_flag_2", "feature_flag_3", "feature_flag_4"}, nil)
	expectedTreatment(resultTreatments["non_existent_feature"], "control", t)
	expectedTreatment(resultTreatments["feature_flag_1"], "on", t)
	expectedTreatment(resultTreatments["feature_flag_2"], "some_treatment", t)
	expectedTreatment(resultTreatments["feature_flag_3"], "off", t)
	expectedTreatment(resultTreatments["feature_flag_4"], "control", t)

	resultTreatmentsWithConfig := client.TreatmentsWithConfig("example1", []string{"feature_flag_2", "feature_flag_3"}, nil)
	expectedTreatmentAndConfig(resultTreatmentsWithConfig["feature_flag_2"], "some_treatment", "{\"color\":\"white\"}", t)
	expectedTreatmentAndConfig(resultTreatmentsWithConfig["feature_flag_3"], "off", "", t)
}

func getRedisConfWithIP(IPAddressesEnabled bool) (*predis.PrefixedRedisClient, *SplitClient) {
	// Create prefixed client for adding Split
	prefixedClient, _ := redis.NewRedisClient(&commonsCfg.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Password: "",
		Prefix:   "testPrefix",
	}, logging.NewLogger(&logging.LoggerOptions{}))

	raw, err := json.Marshal(*valid)
	if err != nil {
		return nil, nil
	}
	prefixedClient.Set("SPLITIO.split.valid", raw, 0)
	prefixedClient.Set("SPLITIO.splits.till", 1494593336752, 0)

	// Set default configs to connect Client with redis
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.IPAddressesEnabled = IPAddressesEnabled
	cfg.Advanced.ImpressionListener = &ImpressionListenerTest{}
	cfg.OperationMode = conf.RedisConsumer
	cfg.Redis = commonsCfg.RedisConfig{
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

	return prefixedClient, client
}

func deleteDataGenerated(prefixedClient *predis.PrefixedRedisClient) {
	// Deletes generated data
	keys := []string{"SPLITIO.impressions", "SPLITIO.events", "SPLITIO.split.valid", "SPLITIO.splits.till", "SPLITIO.telemetry.config", "SPLITIO.telemetry.init"}
	prefixedClient.Del(keys...)
}

func TestRedisClientWithIPDisabled(t *testing.T) {
	prefixedClient, splitClient := getRedisConfWithIP(false)

	// Grabs created event
	resEvent, _ := prefixedClient.LRange("SPLITIO.events", 0, 1)
	event := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(resEvent[0]), &event)
	metadata := event["m"]
	// Checks if metadata was created with "NA" values
	if metadata["i"] != "NA" || metadata["n"] != "NA" {
		t.Error("Instance Name and Machine IP should have 'NA' values")
	}

	splitClient.Destroy()

	// Grabs created impression
	resImpression, _ := prefixedClient.LRange("SPLITIO.impressions", 0, 1)
	impression := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(resImpression[0]), &impression)
	metadata = impression["m"]
	// Checks if metadata was created with "NA" values
	if metadata["i"] != "NA" || metadata["n"] != "NA" {
		t.Error("Instance Name and Machine IP should have 'NA' values")
	}
	listenerData, _ := ilResult["valid"].(map[string]interface{})
	if listenerData["InstanceName"] != "NA" {
		t.Error("InstanceName should be 'NA")
	}

	deleteDataGenerated(prefixedClient)
}

func TestRedisClientWithIPEnabled(t *testing.T) {
	prefixedClient, splitClient := getRedisConfWithIP(true)

	// Grabs created event
	resEvent, _ := prefixedClient.LRange("SPLITIO.events", 0, 1)
	event := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(resEvent[0]), &event)
	metadata := event["m"]
	// Checks if metadata was created with "NA" values
	if metadata["i"] == "NA" || metadata["n"] == "NA" {
		t.Error("Instance Name and Machine IP should not have 'NA' values")
	}

	splitClient.Destroy()

	// Grabs created impression
	resImpression, _ := prefixedClient.LRange("SPLITIO.impressions", 0, 1)
	impression := make(map[string]map[string]interface{})
	json.Unmarshal([]byte(resImpression[0]), &impression)
	metadata = impression["m"]
	// Checks if metadata was created with "NA" values
	if metadata["i"] == "NA" || metadata["n"] == "NA" {
		t.Error("Instance Name and Machine IP should not have 'NA' values")
	}
	listenerData, _ := ilResult["valid"].(map[string]interface{})
	if listenerData["InstanceName"] == "NA" {
		t.Error("InstanceName should not be 'NA")
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
	cfg.Advanced.TelemetryServiceURL = ts.URL
	cfg.Advanced.AuthServiceURL = ts.URL
	cfg.Advanced.ImpressionListener = &ImpressionListenerTest{}
	cfg.TaskPeriods.ImpressionSync = 1
	cfg.TaskPeriods.EventsSync = 1
	cfg.ImpressionsMode = "Debug"
	cfg.Advanced.StreamingEnabled = false

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

func TestClientOptimized(t *testing.T) {
	var isDestroyCalled int64
	var splitsMock, _ = ioutil.ReadFile("../../testdata/splits_mock_2.json")

	postChannel := make(chan string, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/splitChanges":
			fmt.Fprintln(w, string(splitsMock))
			return
		case "/testImpressions/bulk":
			if r.Header.Get("SplitSDKImpressionsMode") != commonsCfg.ImpressionsModeOptimized {
				t.Error("Wrong header")
			}

			if atomic.LoadInt64(&isDestroyCalled) == 1 {
				rBody, _ := ioutil.ReadAll(r.Body)
				var dataInPost []map[string]interface{}
				err := json.Unmarshal(rBody, &dataInPost)
				if err != nil {
					t.Error(err)
					return
				}
				if len(dataInPost) != 2 {
					t.Error("It should send two impressions in optimized mode")
				}
				for _, ki := range dataInPost {
					if asISlice, ok := ki["i"].([]interface{}); !ok || len(asISlice) != 1 {
						t.Error("It should send only one impression per featureName", dataInPost)
					}
				}
			}

			fmt.Fprintln(w, "ok")
		case "/testImpressions/count":
			fmt.Fprintln(w, "ok")
			if atomic.LoadInt64(&isDestroyCalled) == 1 {
				rBody, _ := ioutil.ReadAll(r.Body)

				var dataInPost map[string][]map[string]interface{}
				err := json.Unmarshal(rBody, &dataInPost)
				if err != nil {
					t.Error(err)
					return
				}

				for _, v := range dataInPost["pf"] {
					if int64(v["m"].(float64)) != util.TruncateTimeFrame(time.Now().UTC().UnixNano()) {
						t.Error("Wrong timeFrame")
					}
					switch v["f"] {
					case "DEMO_MURMUR2":
						if v["rc"].(float64) != 2 {
							t.Error("Wrong rc")
						}
						if int64(v["m"].(float64)) != util.TruncateTimeFrame(time.Now().UTC().UnixNano()) {
							t.Error("Wrong timeFrame")
						}
					case "DEMO_MURMUR":
						if v["rc"].(float64) != 1 {
							t.Error("Wrong rc")
						}
					}
				}
				postChannel <- "finished"
			}
		case "/events/bulk":
			fmt.Fprintln(w, "ok")
		case "/segmentChanges":
			fallthrough
		default:
			fmt.Fprintln(w, "ok")
		}
	}))
	defer ts.Close()

	impTest := &ImpressionListenerTest{}
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.Advanced.EventsURL = ts.URL
	cfg.Advanced.SdkURL = ts.URL
	cfg.Advanced.TelemetryServiceURL = ts.URL
	cfg.Advanced.AuthServiceURL = ts.URL
	cfg.Advanced.ImpressionListener = impTest

	factory, _ := NewSplitFactory("test", cfg)
	client := factory.Client()
	client.BlockUntilReady(2)

	// Calls treatments to generate one valid impression
	time.Sleep(300 * time.Millisecond) // Let's wait until first call of recorders have finished
	client.Treatment("user1", "DEMO_MURMUR2", nil)
	impL1, _ := ilResult["DEMO_MURMUR2"].(map[string]interface{})
	if impL1["Pt"].(int64) != 0 {
		t.Error("Pt should be 0")
	}
	client.Treatment("user1", "DEMO_MURMUR2", nil)
	impL2, _ := ilResult["DEMO_MURMUR2"].(map[string]interface{})
	if impL2["Pt"] != impL1["Time"] {
		t.Error("Pt should be equal to previos imp1")
	}
	client.Treatments("user1", []string{"DEMO_MURMUR2", "DEMO_MURMUR"}, nil)
	impL3, _ := ilResult["DEMO_MURMUR2"].(map[string]interface{})
	if impL3["Pt"] != impL2["Time"] {
		t.Error("Pt should be equal to previos imp2")
	}
	impL4, _ := ilResult["DEMO_MURMUR"].(map[string]interface{})
	if impL4["Pt"].(int64) != 0 {
		t.Error("Pt should be equal to 0")
	}

	atomic.AddInt64(&isDestroyCalled, 1)
	client.Destroy()

	select {
	case <-postChannel:
		return
	case <-time.After(4 * time.Second):
		t.Error("The test couldn't send impressions to check headers")
		return
	}
}

func TestClientNone(t *testing.T) {
	var isDestroyCalled int64
	var splitsMock, _ = ioutil.ReadFile("../../testdata/splits_mock_2.json")

	postChannel := make(chan string, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/splitChanges":
			fmt.Fprintln(w, string(splitsMock))
			return
		case "/testImpressions/bulk":
			t.Error("Should not post impressions")
		case "/testImpressions/count":
			fmt.Fprintln(w, "ok")
			if atomic.LoadInt64(&isDestroyCalled) == 1 {
				rBody, _ := ioutil.ReadAll(r.Body)

				var dataInPost map[string][]map[string]interface{}
				err := json.Unmarshal(rBody, &dataInPost)
				if err != nil {
					t.Error(err)
					return
				}

				for _, v := range dataInPost["pf"] {
					if int64(v["m"].(float64)) != util.TruncateTimeFrame(time.Now().UTC().UnixNano()) {
						t.Error("Wrong timeFrame")
					}
					switch v["f"] {
					case "DEMO_MURMUR2":
						if v["rc"].(float64) != 4 {
							t.Error("Wrong rc")
						}
						if int64(v["m"].(float64)) != util.TruncateTimeFrame(time.Now().UTC().UnixNano()) {
							t.Error("Wrong timeFrame")
						}
					case "DEMO_MURMUR":
						if v["rc"].(float64) != 1 {
							t.Error("Wrong rc")
						}
					}
				}
			}
		case "/keys/ss":
			rBody, _ := ioutil.ReadAll(r.Body)

			var uniques dtos.Uniques
			err := json.Unmarshal(rBody, &uniques)
			if err != nil {
				t.Error(err)
				return
			}

			if len(uniques.Keys) != 2 {
				t.Error("Length should be 2")
			}
			for _, key := range uniques.Keys {
				if key.Feature == "DEMO_MURMUR2" && len(key.Keys) != 2 {
					t.Error("Length should be 2")
				}
				if key.Feature == "DEMO_MURMUR" && len(key.Keys) != 1 {
					t.Error("Length should be 1")
				}
			}

			postChannel <- "finished"
		case "/events/bulk":
			fmt.Fprintln(w, "ok")
		case "/segmentChanges":
			fallthrough
		default:
			fmt.Fprintln(w, "ok")
		}
	}))
	defer ts.Close()

	impTest := &ImpressionListenerTest{}
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.Advanced.StreamingEnabled = false
	cfg.Advanced.EventsURL = ts.URL
	cfg.Advanced.SdkURL = ts.URL
	cfg.Advanced.TelemetryServiceURL = ts.URL
	cfg.Advanced.ImpressionListener = impTest
	cfg.ImpressionsMode = commonsCfg.ImpressionsModeNone

	factory, _ := NewSplitFactory("test", cfg)
	client := factory.Client()
	client.BlockUntilReady(2)

	// Calls treatments to generate one valid impression
	time.Sleep(300 * time.Millisecond) // Let's wait until first call of recorders have finished
	client.Treatment("user1", "DEMO_MURMUR2", nil)
	client.Treatment("user1", "DEMO_MURMUR2", nil)
	client.Treatments("user1", []string{"DEMO_MURMUR2", "DEMO_MURMUR"}, nil)
	client.Treatment("user2", "DEMO_MURMUR2", nil)

	atomic.AddInt64(&isDestroyCalled, 1)
	client.Destroy()

	select {
	case <-postChannel:
		return
	case <-time.After(4 * time.Second):
		t.Error("The test couldn't send impressions to check headers")
		return
	}
}

func TestClientDebug(t *testing.T) {
	var isDestroyCalled = false
	var splitsMock, _ = ioutil.ReadFile("../../testdata/splits_mock_2.json")

	postChannel := make(chan string, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/splitChanges":
			fmt.Fprintln(w, string(splitsMock))
			return
		case "/testImpressions/bulk":
			if r.Header.Get("SplitSDKImpressionsMode") != commonsCfg.ImpressionsModeDebug {
				t.Error("Wrong header")
			}

			if isDestroyCalled {
				rBody, _ := ioutil.ReadAll(r.Body)
				var dataInPost []map[string]interface{}
				err := json.Unmarshal(rBody, &dataInPost)
				if err != nil {
					t.Error(err)
					return
				}
				if len(dataInPost) != 1 {
					t.Error("It should send two impressions in optimized mode")
				}
				if len(dataInPost[0]["i"].([]interface{})) != 3 {
					t.Error("It should send only one impression per featureName")
				}
			}

			fmt.Fprintln(w, "ok")
			postChannel <- "finished"
		case "/testImpressions/count":
			t.Error("It should not called this endpoint")
		case "/events/bulk":
			fmt.Fprintln(w, "ok")
		case "/segmentChanges":
			fallthrough
		default:
			fmt.Fprintln(w, "ok")
		}
	}))
	defer ts.Close()

	impTest := &ImpressionListenerTest{}
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.Advanced.EventsURL = ts.URL
	cfg.Advanced.SdkURL = ts.URL
	cfg.Advanced.TelemetryServiceURL = ts.URL
	cfg.Advanced.AuthServiceURL = ts.URL
	cfg.Advanced.ImpressionListener = impTest
	cfg.ImpressionsMode = "Debug"

	factory, _ := NewSplitFactory("test", cfg)
	client := factory.Client()
	client.BlockUntilReady(2)

	// Calls treatments to generate one valid impression
	time.Sleep(300 * time.Millisecond) // Let's wait until first call of recorders have finished
	client.Treatment("user1", "DEMO_MURMUR2", nil)
	impL1, _ := ilResult["DEMO_MURMUR2"].(map[string]interface{})
	if impL1["Pt"].(int64) != 0 {
		t.Error("Pt should be 0")
	}
	client.Treatment("user1", "DEMO_MURMUR2", nil)
	impL2, _ := ilResult["DEMO_MURMUR2"].(map[string]interface{})
	if impL2["Pt"] != impL1["Time"] {
		t.Error("Pt should be equal to previos imp1")
	}
	client.Treatments("user1", []string{"DEMO_MURMUR2"}, nil)
	impL3, _ := ilResult["DEMO_MURMUR2"].(map[string]interface{})
	if impL3["Pt"] != impL2["Time"] {
		t.Error("Pt should be equal to previos imp2")
	}

	isDestroyCalled = true
	client.Destroy()

	select {
	case <-postChannel:
		return
	case <-time.After(4 * time.Second):
		t.Error("The test couldn't send impressions to check headers")
		return
	}
}

func TestUnsupportedMatcherAndSemver(t *testing.T) {
	var isDestroyCalled = false
	var splitsMock, _ = ioutil.ReadFile("../../testdata/splits_mock_3.json")

	postChannel := make(chan string, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth":
			if r.URL.Query().Get("s") != "1.1" {
				t.Error("should be parameter s, for flags spec")
			}
			fmt.Fprintln(w, "{\"pushEnabled\": false, \"token\": \"token\"}")
			return
		case "/splitChanges":
			fmt.Fprintln(w, string(splitsMock))
			return
		case "/testImpressions/bulk":
			if r.Header.Get("SplitSDKImpressionsMode") != commonsCfg.ImpressionsModeOptimized {
				t.Error("Wrong header")
			}

			if isDestroyCalled {
				rBody, _ := ioutil.ReadAll(r.Body)
				var dataInPost []map[string]interface{}
				err := json.Unmarshal(rBody, &dataInPost)
				if err != nil {
					t.Error(err)
					return
				}
				if len(dataInPost) != 6 {
					t.Error("It should send two impressions in optimized mode")
				}
				for _, ki := range dataInPost {
					if asISlice, ok := ki["i"].([]interface{}); !ok || len(asISlice) != 1 {
						t.Error("It should send only one impression per featureName", dataInPost)
					}
					if ki["f"] == "unsupported" {
						message := ki["i"].([]interface{})[0].(map[string]interface{})["r"]
						if message != "targeting rule type unsupported by sdk" {
							t.Error("message sould be: targeting rule type unsupported by sdk")
						}
					}
				}
			}

			fmt.Fprintln(w, "ok")
			postChannel <- "finished"
		case "/testImpressions/count":
			fallthrough
		case "/keys/ss":
			fallthrough
		case "/events/bulk":
			fallthrough
		case "/segmentChanges":
			fallthrough
		default:
			fmt.Fprintln(w, "ok")
		}
	}))
	defer ts.Close()

	cfg := conf.Default()
	cfg.Advanced.AuthServiceURL = ts.URL
	cfg.Advanced.EventsURL = ts.URL
	cfg.Advanced.SdkURL = ts.URL
	cfg.Advanced.TelemetryServiceURL = ts.URL

	factory, _ := NewSplitFactory("test", cfg)
	client := factory.Client()
	client.BlockUntilReady(2)

	// Calls treatments to generate one valid impression
	time.Sleep(300 * time.Millisecond) // Let's wait until first call of recorders have finished
	attributes := make(map[string]interface{})
	attributes["version"] = "1.22.9"
	evaluation := client.Treatment("user1", "semver", attributes)
	if evaluation != "on" {
		t.Error("evaluation for semver should be on")
	}
	attributes["version"] = "2.0.0"
	evaluation = client.Treatment("user1", "semver1", attributes)
	if evaluation != "on" {
		t.Error("evaluation for semver should be on")
	}
	evaluation = client.Treatment("user1", "semver2", attributes)
	if evaluation != "on" {
		t.Error("evaluation for semver should be on")
	}
	attributes["version"] = "1.0.0"
	evaluation = client.Treatment("user1", "semver3", attributes)
	if evaluation != "on" {
		t.Error("evaluation for semver should be on")
	}
	attributes["version"] = "2.1.0"
	evaluation = client.Treatment("user1", "semver4", attributes)
	if evaluation != "on" {
		t.Error("evaluation for semver should be on")
	}
	evaluation = client.Treatment("user1", "unsupported", nil)
	if evaluation != "control" {
		t.Error("evaluation for unsupported should be control")
	}

	isDestroyCalled = true
	client.Destroy()

	select {
	case <-postChannel:
		return
	case <-time.After(4 * time.Second):
		t.Error("The test couldn't send impressions to check headers")
		return
	}
}

func TestTelemetryMemory(t *testing.T) {
	factoryInstances = make(map[string]int64)
	var metricsInitCalled int64
	var metricsStatsCalled int64

	sdkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		splitChanges := dtos.SplitChangesDTO{
			Splits: []dtos.SplitDTO{
				{Name: "split1", Killed: true, Status: "ACTIVE", DefaultTreatment: "on"},
				{Name: "split2", Killed: true, Status: "ACTIVE"},
				{Name: "split3", Killed: true, Status: "INACTIVE"},
			},
			Since: 3,
			Till:  3,
		}

		raw, err := json.Marshal(splitChanges)
		if err != nil {
			t.Error("Error building json")
			return
		}

		w.Write(raw)
	}))
	defer sdkServer.Close()

	eventsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer eventsServer.Close()

	telemetryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/metrics/config":
			atomic.AddInt64(&metricsInitCalled, 1)
			rBody, _ := ioutil.ReadAll(r.Body)
			var dataInPost dtos.Config
			err := json.Unmarshal(rBody, &dataInPost)
			if err != nil {
				t.Error(err)
				return
			}

			if dataInPost.Storage != telemetry.Memory {
				t.Error("It should initiate in memory mode")
			}
			if dataInPost.Rates.Events != 200 || dataInPost.Rates.Impressions != 300 || dataInPost.Rates.Segments != 100 || dataInPost.Rates.Splits != 60 || dataInPost.Rates.Telemetry != 3600 {
				t.Error("Wrong rates")
			}
			if dataInPost.NonReadyUsages != 0 {
				t.Error("It should not be non ready usages")
			}
			if dataInPost.BurTimeouts != 0 {
				t.Error("It should not be bur timeouts")
			}
			if dataInPost.OperationMode != telemetry.Standalone {
				t.Error("It should be Standalone")
			}
			switch metricsInitCalled {
			case 1:
				if dataInPost.RedundantFactories != 0 {
					t.Error("It should be 0")
				}
				if dataInPost.ActiveFactories != 1 {
					t.Error("It should be 1")
				}
				if dataInPost.ImpressionsMode != telemetry.ImpressionsModeOptimized {
					t.Error("It should be Optimized")
				}
			case 2:
				if dataInPost.RedundantFactories != 1 {
					t.Error("It should be 1")
				}
				if dataInPost.ActiveFactories != 1 {
					t.Error("It should be 1")
				}
				if dataInPost.ImpressionsMode != telemetry.ImpressionsModeDebug {
					t.Error("It should be Debug")
				}
			case 3:
				if dataInPost.RedundantFactories != 1 {
					t.Error("It should be 1")
				}
				if dataInPost.ActiveFactories != 2 {
					t.Error("It should be 2")
				}
				if dataInPost.ImpressionsMode != telemetry.ImpressionsModeNone {
					t.Error("It should be None")
				}
			}

			if dataInPost.TimeUntilReady == 0 {
				t.Error("It should record ready")
			}
			if dataInPost.ImpressionsListenerEnabled {
				t.Error("It should not have impression listener")
			}
		case "/metrics/usage":
			atomic.AddInt64(&metricsStatsCalled, 1)
			rBody, _ := ioutil.ReadAll(r.Body)
			var dataInPost dtos.Stats
			err := json.Unmarshal(rBody, &dataInPost)
			if err != nil {
				t.Error(err)
				return
			}

			if dataInPost.LastSynchronizations.Splits == 0 || dataInPost.LastSynchronizations.Telemetry == 0 {
				t.Error("It should record lastSynchronizations")
			}
			if dataInPost.SplitCount != 2 {
				t.Error("It should have 2 splits")
			}
			if dataInPost.SessionLengthMs == 0 {
				t.Error("It should record sessionsLength")
			}
			if dataInPost.UpdatesFromSSE.Splits != 0 {
				t.Error("It should send ufs")
			}

			switch metricsStatsCalled {
			case 1:
				if dataInPost.ImpressionsQueued != 1 {
					t.Error("It should queue one impression")
				}
				if dataInPost.ImpressionsDeduped != 1 {
					t.Error("It should dedupe one impression. ", dataInPost.ImpressionsDeduped)
				}
				if dataInPost.EventsQueued != 1 {
					t.Error("It should queue one event")
				}
			case 2:
				if dataInPost.ImpressionsQueued != 2 {
					t.Error("It should queue 2 impressions")
				}
				if dataInPost.ImpressionsDeduped != 0 {
					t.Error("It should not dedupe impressions")
				}
				if dataInPost.EventsQueued != 1 {
					t.Error("It should queue one event")
				}
			case 3:
				if dataInPost.ImpressionsQueued != 0 {
					t.Error("It should not queue impressions")
				}
				if dataInPost.ImpressionsDeduped != 0 {
					t.Error("It should not dedupe impressions")
				}
				if dataInPost.EventsQueued != 0 {
					t.Error("It should not track event")
				}
			}
		}

		fmt.Fprintln(w, "ok")
	}))
	defer telemetryServer.Close()

	sdkConf := conf.Default()
	sdkConf.LoggerConfig.StandardLoggerFlags = log.Llongfile
	sdkConf.Advanced.EventsURL = eventsServer.URL
	sdkConf.Advanced.SdkURL = sdkServer.URL
	sdkConf.Advanced.TelemetryServiceURL = telemetryServer.URL
	sdkConf.Advanced.StreamingEnabled = false
	sdkConf.TaskPeriods.EventsSync = 200
	sdkConf.TaskPeriods.SegmentSync = 100

	factory, _ := NewSplitFactory("something", sdkConf)

	client := factory.Client()
	manager := factory.Manager()
	client.BlockUntilReady(1)
	if len(manager.SplitNames()) != 2 {
		t.Error("It should return splits")
	}
	client.Treatment("some", "split1", nil)
	client.Treatment("some", "split1", nil)
	client.Track("something", "something", "something", nil, nil)

	sdkConf.ImpressionsMode = "debug"
	factory2, _ := NewSplitFactory("something", sdkConf)
	manager2 := factory2.Manager()
	client2 := factory2.Client()
	client2.BlockUntilReady(1)
	if len(manager2.SplitNames()) != 2 {
		t.Error("It should return splits")
	}
	client2.Treatment("some", "split1", nil)
	client2.Treatment("some", "split1", nil)
	client2.Track("something", "something", "something", nil, nil)

	sdkConf.ImpressionsMode = "none"
	factory3, _ := NewSplitFactory("something2", sdkConf)
	manager3 := factory3.Manager()
	client3 := factory3.Client()
	client3.BlockUntilReady(1)
	if len(manager3.SplitNames()) != 2 {
		t.Error("It should return splits")
	}
	client3.Treatment("some", "split1", nil)
	client3.Treatment("some", "split1", nil)

	factory.Destroy()
	factory2.Destroy()
	factory3.Destroy()

	if atomic.LoadInt64(&metricsInitCalled) != 3 {
		t.Error("It should send init data")
	}
	if atomic.LoadInt64(&metricsStatsCalled) != 3 {
		t.Error("It should send stats data")
	}
}

func TestTelemetryRedis(t *testing.T) {
	factoryInstances = make(map[string]int64)
	sdkConf := conf.Default()
	sdkConf.OperationMode = conf.RedisConsumer

	factory, _ := NewSplitFactory("something", sdkConf)
	md := fmt.Sprintf("go-%s/%s/%s", splitio.Version, sdkConf.InstanceName, sdkConf.IPAddress)

	if !factory.IsReady() {
		t.Error("Factory should be ready immediately")
	}

	client := factory.Client()
	err := client.BlockUntilReady(1)
	if err != nil {
		t.Error("Error was not expected")
	}

	prefixedClient, _ := redis.NewRedisClient(&commonsCfg.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Prefix:   "",
	}, logging.NewLogger(&logging.LoggerOptions{}))
	data, err := prefixedClient.HGetAll("SPLITIO.telemetry.init")
	if err != nil {
		t.Error("It should not return err")
	}
	if len(data) != 1 {
		t.Error("It should store one")
	}

	forMd, ok := data[md]
	if !ok {
		t.Error("no config found for calculated metadata: ", md)
		t.Error(data)
	}

	var dataInRedis dtos.Config
	err = json.Unmarshal([]byte(forMd), &dataInRedis)
	if err != nil {
		t.Error("Should not return error umarshalling")
	}

	if dataInRedis.ActiveFactories != 1 {
		t.Error("Wrong value")
	}
	if dataInRedis.OperationMode != telemetry.Consumer {
		t.Error("It should be consumer")
	}
	if dataInRedis.Storage != telemetry.Redis {
		t.Error("It should be redis")
	}

	deleteDataGenerated(prefixedClient)
	factory.Destroy()
}

func TestClientNoneRedis(t *testing.T) {
	redisConfig := &commonsCfg.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Prefix:   "test-prefix-m",
	}

	prefixedClient, _ := redis.NewRedisClient(redisConfig, logging.NewLogger(&logging.LoggerOptions{}))
	raw, _ := json.Marshal(*valid)
	prefixedClient.Set("SPLITIO.split.valid", raw, 0)
	raw, _ = json.Marshal(*noConfig)
	prefixedClient.Set("SPLITIO.split.noConfig", raw, 0)

	impTest := &ImpressionListenerTest{}
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.Advanced.ImpressionListener = impTest
	cfg.ImpressionsMode = commonsCfg.ImpressionsModeNone
	cfg.OperationMode = conf.RedisConsumer
	cfg.Redis = *redisConfig

	factory, _ := NewSplitFactory("test", cfg)
	client := factory.Client()
	client.BlockUntilReady(2)

	// Calls treatments to generate one valid impression
	time.Sleep(300 * time.Millisecond) // Let's wait until first call of recorders have finished
	client.Treatment("user1", "valid", nil)
	client.Treatment("user2", "valid", nil)
	client.Treatment("user3", "valid", nil)
	client.Treatment("user1", "valid", nil)
	client.Treatment("user2", "valid", nil)
	client.Treatment("user3", "valid", nil)
	client.Treatment("user3", "noConfig", nil)
	client.Treatment("user3", "noConfig", nil)
	client.Destroy()

	// Validate unique keys
	uniques, _ := prefixedClient.LRange("SPLITIO.uniquekeys", 0, -1)

	keysDto := make([]dtos.Key, 0)
	for _, key := range uniques {
		var keyDto dtos.Key
		_ = json.Unmarshal([]byte(key), &keyDto)
		keysDto = append(keysDto, keyDto)
	}

	if len(keysDto) != 2 {
		t.Errorf("Lenght should be 2. Actual %d", len(keysDto))
	}

	for _, unique := range keysDto {
		if unique.Feature == "valid" && len(unique.Keys) != 3 {
			t.Error("Keys should be 3")
		}
		if unique.Feature == "noConfig" && len(unique.Keys) != 1 {
			t.Error("Keys should be 1")
		}
	}

	// Validate impression counts
	impressionscount, _ := prefixedClient.HGetAll("SPLITIO.impressions.count")

	for key, count := range impressionscount {
		if strings.HasPrefix(key, "valid::") && count != "6" {
			t.Error("Expected: 6. actual: " + count)
		}
		if strings.HasPrefix(key, "noConfig::") && count != "2" {
			t.Error("Expected: 2. actual: " + count)
		}
	}

	// Validate that impressions doesn't exist
	exist, _ := prefixedClient.Exists("SPLITIO.impressions")
	if exist != 0 {
		t.Error("SPLITIO.impressions should not exist")
	}

	// Clean redis
	keys, _ := prefixedClient.Keys("SPLITIO*")
	for _, k := range keys {
		prefixedClient.Del(k)
	}
}

func TestClientOptimizedRedis(t *testing.T) {
	redisConfig := &commonsCfg.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Prefix:   "test-prefix-m",
	}

	prefixedClient, _ := redis.NewRedisClient(redisConfig, logging.NewLogger(&logging.LoggerOptions{}))
	raw, _ := json.Marshal(*valid)
	prefixedClient.Set("SPLITIO.split.valid", raw, 0)
	raw, _ = json.Marshal(*noConfig)
	prefixedClient.Set("SPLITIO.split.noConfig", raw, 0)

	impTest := &ImpressionListenerTest{}
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.Advanced.ImpressionListener = impTest
	cfg.ImpressionsMode = commonsCfg.ImpressionsModeOptimized
	cfg.OperationMode = conf.RedisConsumer
	cfg.Redis = *redisConfig

	factory, _ := NewSplitFactory("test", cfg)
	client := factory.Client()
	client.BlockUntilReady(2)

	// Calls treatments to generate one valid impression
	time.Sleep(300 * time.Millisecond) // Let's wait until first call of recorders have finished
	client.Treatment("user1", "valid", nil)
	client.Treatment("user2", "valid", nil)
	client.Treatment("user3", "valid", nil)
	client.Treatment("user1", "valid", nil)
	client.Treatment("user2", "valid", nil)
	client.Treatment("user3", "valid", nil)
	client.Treatment("user3", "noConfig", nil)
	client.Treatment("user3", "noConfig", nil)
	client.Destroy()

	// Validate impressions
	impressions, _ := prefixedClient.LRange("SPLITIO.impressions", 0, -1)

	if len(impressions) != 4 {
		t.Error("Impression length shold be 4")
	}

	for _, imp := range impressions {
		var imprObject dtos.ImpressionQueueObject
		_ = json.Unmarshal([]byte(imp), &imprObject)

		if imprObject.Impression.KeyName == "user1" && imprObject.Impression.FeatureName == "valid" && imprObject.Impression.Pt != 0 {
			t.Error("Pt should be 0.")
		}
		if imprObject.Impression.KeyName == "user2" && imprObject.Impression.FeatureName == "valid" && imprObject.Impression.Pt != 0 {
			t.Error("Pt should be 0.")
		}
		if imprObject.Impression.KeyName == "user3" && imprObject.Impression.FeatureName == "valid" && imprObject.Impression.Pt != 0 {
			t.Error("Pt should be 0.")
		}
		if imprObject.Impression.KeyName == "user3" && imprObject.Impression.FeatureName == "noConfig" && imprObject.Impression.Pt != 0 {
			t.Error("Pt should be 0.")
		}
	}

	// Validate impression counts
	impressionscount, _ := prefixedClient.HGetAll("SPLITIO.impressions.count")

	for key, count := range impressionscount {
		if strings.HasPrefix(key, "valid::") && count != "3" {
			t.Error("Expected: 3. actual: " + count)
		}
		if strings.HasPrefix(key, "noConfig::") && count != "1" {
			t.Error("Expected: 1. actual: " + count)
		}
	}

	// Validate that uniquekeys doesn't exist
	exist, _ := prefixedClient.Exists("SPLITIO.uniquekeys")
	if exist != 0 {
		t.Error("SPLITIO.uniquekeys should not exist")
	}

	// Clean redis
	keys, _ := prefixedClient.Keys("SPLITIO*")
	for _, k := range keys {
		prefixedClient.Del(k)
	}
}

func TestClientDebugRedis(t *testing.T) {
	redisConfig := &commonsCfg.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Prefix:   "test-prefix-m",
	}

	prefixedClient, _ := redis.NewRedisClient(redisConfig, logging.NewLogger(&logging.LoggerOptions{}))
	raw, _ := json.Marshal(*valid)
	prefixedClient.Set("SPLITIO.split.valid", raw, 0)
	raw, _ = json.Marshal(*noConfig)
	prefixedClient.Set("SPLITIO.split.noConfig", raw, 0)

	impTest := &ImpressionListenerTest{}
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.Advanced.ImpressionListener = impTest
	cfg.ImpressionsMode = commonsCfg.ImpressionsModeDebug
	cfg.OperationMode = conf.RedisConsumer
	cfg.Redis = *redisConfig

	factory, _ := NewSplitFactory("test", cfg)
	client := factory.Client()
	client.BlockUntilReady(2)

	// Calls treatments to generate one valid impression
	time.Sleep(300 * time.Millisecond) // Let's wait until first call of recorders have finished
	client.Treatment("user1", "valid", nil)
	client.Treatment("user2", "valid", nil)
	client.Treatment("user3", "valid", nil)
	client.Treatment("user1", "valid", nil)
	client.Treatment("user2", "valid", nil)
	client.Treatment("user3", "valid", nil)
	client.Treatment("user3", "noConfig", nil)
	client.Treatment("user3", "noConfig", nil)
	client.Destroy()

	// Validate impressions
	impressions, _ := prefixedClient.LRange("SPLITIO.impressions", 0, -1)
	if len(impressions) != 8 {
		t.Errorf("Impression length should be 8. Actual %d", len(impressions))
	}

	// Validate impression counts
	impressionscount, _ := prefixedClient.HGetAll("SPLITIO.impressions.count")

	for key, count := range impressionscount {
		if strings.HasPrefix(key, "valid::") && count != "6" {
			t.Error("Expected: 6. actual: " + count)
		}
		if strings.HasPrefix(key, "noConfig::") && count != "2" {
			t.Error("Expected: 2. actual: " + count)
		}
	}

	// Validate that uniquekeys doesn't exist
	exist, _ := prefixedClient.Exists("SPLITIO.uniquekeys")
	if exist != 0 {
		t.Error("SPLITIO.uniquekeys should not exist")
	}
	exist, _ = prefixedClient.Exists("SPLITIO.impressions.count")
	if exist != 0 {
		t.Error("SPLITIO.impressions.count should not exist")
	}

	// Clean redis
	keys, _ := prefixedClient.Keys("SPLITIO*")
	for _, k := range keys {
		prefixedClient.Del(k)
	}
}

var semver string = "3.4.5"
var attribute string = "version"

var splitSemver = &dtos.SplitDTO{
	Algo:                  2,
	ChangeNumber:          1494593336752,
	DefaultTreatment:      "off",
	Killed:                false,
	Name:                  "semver",
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
							Attribute:   &attribute,
						},
						MatcherType: "EQUAL_TO_SEMVER",
						String:      &semver,
						Whitelist:   nil,
						Negate:      false,
					},
				},
			},
			Partitions: []dtos.PartitionDTO{
				{
					Size:      100,
					Treatment: "on",
				},
				{
					Size:      0,
					Treatment: "off",
				},
			},
		},
	},
}

var splitUnsupported = &dtos.SplitDTO{
	Algo:                  2,
	ChangeNumber:          1494593336752,
	DefaultTreatment:      "off",
	Killed:                false,
	Name:                  "unsupported",
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
						MatcherType: "UNSUPPORTED",
						Whitelist:   nil,
						Negate:      false,
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

func TestUnsupportedandSemverMatcherRedis(t *testing.T) {
	redisConfig := &commonsCfg.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Prefix:   "test-prefix-semver",
	}

	prefixedClient, _ := redis.NewRedisClient(redisConfig, logging.NewLogger(&logging.LoggerOptions{}))
	raw, _ := json.Marshal(*splitSemver)
	prefixedClient.Set("SPLITIO.split.semver", raw, 0)
	raw, _ = json.Marshal(*splitUnsupported)
	prefixedClient.Set("SPLITIO.split.unsupported", raw, 0)

	impTest := &ImpressionListenerTest{}
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.Advanced.ImpressionListener = impTest
	cfg.ImpressionsMode = commonsCfg.ImpressionsModeOptimized
	cfg.OperationMode = conf.RedisConsumer
	cfg.Redis = *redisConfig

	factory, _ := NewSplitFactory("test", cfg)
	client := factory.Client()
	client.BlockUntilReady(2)

	// Calls treatments to generate one valid impression
	time.Sleep(300 * time.Millisecond) // Let's wait until first call of recorders have finished
	attributes := make(map[string]interface{})
	attributes["version"] = "3.4.5"
	evaluation := client.Treatment("user1", "semver", attributes)
	if evaluation != "on" {
		t.Error("evaluation for semver should be on")
	}
	evaluation = client.Treatment("user2", "unsupported", nil)
	if evaluation != "control" {
		t.Error("evaluation for unsupported should be control")
	}
	client.Destroy()

	// Validate impressions
	impressions, _ := prefixedClient.LRange("SPLITIO.impressions", 0, -1)

	if len(impressions) != 2 {
		t.Error("Impression length shold be 2")
	}

	for _, imp := range impressions {
		var imprObject dtos.ImpressionQueueObject
		_ = json.Unmarshal([]byte(imp), &imprObject)

		if imprObject.Impression.KeyName == "user1" && imprObject.Impression.FeatureName == "semver" && imprObject.Impression.Pt != 0 {
			t.Error("Pt should be 0.")
		}
		if imprObject.Impression.KeyName == "user2" && imprObject.Impression.FeatureName == "unsupported" && imprObject.Impression.Pt != 0 {
			t.Error("Pt should be 0.")
		}
	}

	// Clean redis
	keys, _ := prefixedClient.Keys("SPLITIO*")
	for _, k := range keys {
		prefixedClient.Del(k)
	}
}
