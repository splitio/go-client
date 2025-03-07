package client

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/splitio/go-client/v6/splitio/conf"
	commonsCfg "github.com/splitio/go-split-commons/v6/conf"
	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/go-split-commons/v6/flagsets"
	"github.com/splitio/go-split-commons/v6/healthcheck/application"
	"github.com/splitio/go-split-commons/v6/provisional"
	"github.com/splitio/go-split-commons/v6/provisional/strategy"
	"github.com/splitio/go-split-commons/v6/service/api"
	authMocks "github.com/splitio/go-split-commons/v6/service/mocks"
	"github.com/splitio/go-split-commons/v6/storage/inmemory/mutexmap"
	"github.com/splitio/go-split-commons/v6/storage/inmemory/mutexqueue"
	"github.com/splitio/go-split-commons/v6/storage/mocks"
	"github.com/splitio/go-split-commons/v6/storage/redis"
	"github.com/splitio/go-split-commons/v6/synchronizer"
	"github.com/splitio/go-toolkit/v5/logging"
)

type MockWriter struct {
	mutex    sync.RWMutex
	messages []string
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = append(m.messages, string(p[:]))
	return len(p), nil
}

func (m *MockWriter) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = nil
}

func (m *MockWriter) Length() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.messages)
}

func (m *MockWriter) Matches(expected string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, msg := range m.messages {
		if strings.Contains(msg, expected) {
			m.messages = nil
			return true
		}
	}
	m.messages = nil
	return false
}

var mW MockWriter

func getMockedLogger() logging.LoggerInterface {
	return logging.NewLogger(&logging.LoggerOptions{
		LogLevel:      logging.LevelInfo,
		ErrorWriter:   &mW,
		WarningWriter: &mW,
		InfoWriter:    &mW,
		DebugWriter:   nil,
		VerboseWriter: nil,
	})
}

func getClient() SplitClient {
	logger := getMockedLogger()
	cfg := conf.Default()
	telemetryMockedStorage := mocks.MockTelemetryStorage{
		RecordImpressionsStatsCall: func(dataType int, count int64) {},
		RecordLatencyCall:          func(method string, latency time.Duration) {},
	}

	impressionObserver, _ := strategy.NewImpressionObserver(500)
	impressionsCounter := strategy.NewImpressionsCounter()
	impressionsStrategy := strategy.NewOptimizedImpl(impressionObserver, impressionsCounter, telemetryMockedStorage, true)
	impressionManager := provisional.NewImpressionManager(impressionsStrategy).(*provisional.ImpressionManagerImpl)

	factory := &SplitFactory{cfg: cfg, impressionManager: impressionManager,
		storages: sdkStorages{
			runtimeTelemetry:    telemetryMockedStorage,
			initTelemetry:       telemetryMockedStorage,
			evaluationTelemetry: telemetryMockedStorage,
		}}

	client := SplitClient{
		evaluator:   &mockEvaluator{},
		impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger, telemetryMockedStorage),
		logger:      logger,
		validator: inputValidation{
			logger: logger,
			splitStorage: mocks.MockSplitStorage{
				TrafficTypeExistsCall: func(trafficType string) bool {
					switch trafficType {
					case "trafictype":
						return true
					default:
						return false
					}
				},
			},
		},
		events: mocks.MockEventStorage{
			PushCall: func(event dtos.EventDTO, size int) error { return nil },
		},
		factory:             factory,
		impressionManager:   impressionManager,
		initTelemetry:       telemetryMockedStorage,
		evaluationTelemetry: telemetryMockedStorage,
		runtimeTelemetry:    telemetryMockedStorage,
	}
	factory.status.Store(sdkStatusReady)
	return client
}

func TestFactoryWithNilApiKey(t *testing.T) {
	cfg := conf.Default()
	cfg.Logger = getMockedLogger()
	_, err := NewSplitFactory("", cfg)

	if err == nil {
		t.Error("Should be error")
	}

	expected := "factory instantiation: you passed an empty SDK key, SDK key must be a non-empty string"
	if !mW.Matches(expected) {
		t.Error("Error is distinct from the expected one")
	}
}

func getLongKey() string {
	m := ""
	for n := 0; n <= 256; n++ {
		m += "m"
	}
	return m
}

func TestValidationEmpty(t *testing.T) {
	client := getClient()
	mW.Reset()
	expectedTreatment(client.Treatment("key", "feature", nil), "TreatmentA", t)
	if mW.Length() > 0 {
		t.Error("Wrong message")
	}
	mW.Reset()
}

func TestTreatmentValidatorOnKeys(t *testing.T) {
	client := getClient()
	// Nil
	expectedTreatment(client.Treatment(nil, "feature", nil), "control", t)
	if !mW.Matches("Treatment: you passed a nil key, key must be a non-empty string") {
		t.Error("Wrong message")
	}

	// Boolean
	expectedTreatment(client.Treatment(true, "feature", nil), "control", t)
	if !mW.Matches("Treatment: you passed an invalid key, key must be a non-empty string") {
		t.Error("Wrong message")
	}

	// Trimmed
	expectedTreatment(client.Treatment("     ", "feature", nil), "control", t)
	if !mW.Matches("Treatment: you passed an empty key, key must be a non-empty string") {
		t.Error("Wrong message")
	}

	// Long
	expectedTreatment(client.Treatment(getLongKey(), "feature", nil), "control", t)
	if !mW.Matches("Treatment: key too long - must be 250 characters or less") {
		t.Error("Wrong message")
	}

	// Int
	expectedTreatment(client.Treatment(123, "feature", nil), "TreatmentA", t)
	if !mW.Matches("Treatment: key %!s(int=123) is not of type string, converting") {
		t.Error("Wrong message")
	}

	// Int32
	expectedTreatment(client.Treatment(int32(123), "feature", nil), "TreatmentA", t)
	if !mW.Matches("Treatment: key %!s(int32=123) is not of type string, converting") {
		t.Error("Wrong message")
	}

	// Int 64
	expectedTreatment(client.Treatment(int64(123), "feature", nil), "TreatmentA", t)
	if !mW.Matches("Treatment: key %!s(int64=123) is not of type string, converting") {
		t.Error("Wrong message")
	}

	// Float
	expectedTreatment(client.Treatment(1.3, "feature", nil), "TreatmentA", t)
	if !mW.Matches("Treatment: key %!s(float64=1.3) is not of type string, converting") {
		t.Error("Wrong message")
	}

	// NaN
	expectedTreatment(client.Treatment(math.NaN, "feature", nil), "control", t)
	if !mW.Matches("Treatment: you passed an invalid key, key must be a non-empty string") {
		t.Error("Wrong message")
	}

	// Inf
	expectedTreatment(client.Treatment(math.Inf, "feature", nil), "control", t)
	if !mW.Matches("Treatment: you passed an invalid key, key must be a non-empty string") {
		t.Error("Wrong message")
	}
}

func getKey(matchingKey string, bucketingKey string) *Key {
	return &Key{
		MatchingKey:  matchingKey,
		BucketingKey: bucketingKey,
	}
}

func TestTreatmentValidatorWithKeyObject(t *testing.T) {
	client := getClient()
	// Empty
	expectedTreatment(client.Treatment(getKey("", "bucketing"), "feature", nil), "control", t)
	if !mW.Matches("Treatment: you passed an empty matchingKey, matchingKey must be a non-empty string") {
		t.Error("Wrong message")
	}

	// Long
	expectedTreatment(client.Treatment(getKey(getLongKey(), "bucketing"), "feature", nil), "control", t)
	if !mW.Matches("Treatment: matchingKey too long - must be 250 characters or less") {
		t.Error("Wrong message")
	}

	// Empty Bucketing
	expectedTreatment(client.Treatment(getKey("matching", ""), "feature", nil), "control", t)
	if !mW.Matches("Treatment: you passed an empty bucketingKey, bucketingKey must be a non-empty string") {
		t.Error("Wrong message")
	}

	// Long Bucketing
	expectedTreatment(client.Treatment(getKey("matching", getLongKey()), "feature", nil), "control", t)
	if !mW.Matches("Treatment: bucketingKey too long - must be 250 characters or less") {
		t.Error("Wrong message")
	}

	// Ok
	mW.Reset()
	expectedTreatment(client.Treatment(getKey("matching", "bucketing"), "feature", nil), "TreatmentA", t)
	if mW.Length() > 0 {
		t.Error("Wrong message")
	}
	mW.Reset()
}

func TestTreatmentValidatorOnFeatureName(t *testing.T) {
	client := getClient()
	// Empty
	expectedTreatment(client.Treatment("key", "", nil), "control", t)
	if !mW.Matches("Treatment: you passed an empty featureFlagName, flag name must be a non-empty string") {
		t.Error("Wrong message")
	}

	// Trimmed
	expectedTreatment(client.Treatment("key", "  feature   ", nil), "TreatmentA", t)
	if !mW.Matches("Treatment: featureFlagName '  feature   ' has extra whitespace, trimming") {
		t.Error("Wrong message")
	}

	// Non Existent
	expectedTreatment(client.Treatment("key", "feature_non_existent", nil), "control", t)
	if !mW.Matches("Treatment: you passed feature_non_existent that does not exist in this environment, please double check what feature flags exist in the Split user interface") {
		t.Error("Wrong message")
	}

	// Non Existent
	expectedTreatmentAndConfig(client.TreatmentWithConfig("key", "feature_non_existent", nil), "control", "", t)
	if !mW.Matches("TreatmentWithConfig: you passed feature_non_existent that does not exist in this environment, please double check what feature flags exist in the Split user interface") {
		t.Error("Wrong message")
	}
}

func expectedTreatments(key interface{}, features []string, length int, t *testing.T) map[string]string {
	client := getClient()
	result := client.Treatments(key, features, nil)
	if len(result) != length {
		t.Error("Wrong len of elements")
	}
	return result
}

func TestTreatmentsValidator(t *testing.T) {
	client := getClient()
	// Empty features
	expectedTreatments("key", []string{""}, 0, t)
	if !mW.Matches("Treatments: featureFlagNames must be a non-empty array") {
		t.Error("Wrong message")
	}

	// Inf
	result := expectedTreatments(math.Inf, []string{"feature"}, 1, t)
	expectedTreatment(result["feature"], "control", t)
	if !mW.Matches("Treatments: you passed an invalid key, key must be a non-empty string") {
		t.Error("Wrong message")
	}

	// Float
	result = expectedTreatments(1.3, []string{"feature"}, 1, t)
	expectedTreatment(result["feature"], "TreatmentA", t)
	if !mW.Matches("Treatments: key %!s(float64=1.3) is not of type string, converting") {
		t.Error("Wrong message")
	}

	// Trimmed
	result = expectedTreatments("key", []string{" some_feature  "}, 1, t)
	expectedTreatment(result["some_feature"], "control", t)
	if !mW.Matches("Treatments: featureFlagName ' some_feature  ' has extra whitespace, trimming") {
		t.Error("Wrong message")
	}

	// Non Existent
	result = expectedTreatments("key", []string{"feature_non_existent"}, 1, t)
	expectedTreatment(result["feature_non_existent"], "control", t)
	if !mW.Matches("Treatments: you passed feature_non_existent that does not exist in this environment, please double check what feature flags exist in the Split user interface") {
		t.Error("Wrong message")
	}

	// Non Existent Config
	resultWithConfig := client.TreatmentsWithConfig("key", []string{"feature_non_existent"}, nil)
	expectedTreatmentAndConfig(resultWithConfig["feature_non_existent"], "control", "", t)
	if !mW.Matches("TreatmentsWithConfig: you passed feature_non_existent that does not exist in this environment, please double check what feature flags exist in the Split user interface") {
		t.Error("Wrong message")
	}
}

func TestValidatorOnDestroy(t *testing.T) {
	telemetryMockedStorage := mocks.MockTelemetryStorage{
		RecordSessionLengthCall: func(session int64) {},
	}
	logger := getMockedLogger()
	localConfig := &synchronizer.LocalConfig{RefreshEnabled: false}
	sync, _ := synchronizer.NewSynchronizerManager(
		synchronizer.NewLocal(localConfig, &api.SplitAPI{}, mocks.MockSplitStorage{}, mocks.MockSegmentStorage{}, logger, telemetryMockedStorage, &application.Dummy{}),
		logger,
		commonsCfg.AdvancedConfig{},
		authMocks.MockAuthClient{},
		mocks.MockSplitStorage{},
		make(chan int, 1),
		telemetryMockedStorage,
		dtos.Metadata{},
		nil,
		&application.Dummy{},
	)
	factory := &SplitFactory{
		cfg:         conf.Default(),
		syncManager: sync,
		storages: sdkStorages{
			initTelemetry:       telemetryMockedStorage,
			runtimeTelemetry:    telemetryMockedStorage,
			evaluationTelemetry: telemetryMockedStorage,
		},
	}
	factory.status.Store(sdkStatusReady)
	var client2 = SplitClient{
		evaluator:           &mockEvaluator{},
		impressions:         mutexqueue.NewMQImpressionsStorage(5000, make(chan string, 1), logger, telemetryMockedStorage),
		logger:              logger,
		validator:           inputValidation{logger: logger},
		factory:             factory,
		initTelemetry:       telemetryMockedStorage,
		evaluationTelemetry: telemetryMockedStorage,
		runtimeTelemetry:    telemetryMockedStorage,
	}

	var manager = SplitManager{
		logger:        logger,
		validator:     inputValidation{logger: logger},
		factory:       factory,
		initTelemetry: telemetryMockedStorage,
	}

	client2.Destroy()

	expectedTreatment(client2.Treatment("key", "  feature   ", nil), "control", t)
	if !mW.Matches("Client has already been destroyed - no calls possible") {
		t.Error("Wrong message")
	}

	result := client2.Treatments("key", []string{"some_feature"}, nil)
	expectedTreatment(result["some_feature"], "control", t)
	if !mW.Matches("Client has already been destroyed - no calls possible") {
		t.Error("Wrong message")
	}

	expectedTrack(client2.Track("key", "trafficType", "eventType", 0, nil), "Client has already been destroyed - no calls possible", t)

	manager.Split("feature")
	if !mW.Matches("Client has already been destroyed - no calls possible") {
		t.Error("Wrong message")
	}
}

func expectedTrack(err error, expected string, t *testing.T) {
	if err != nil && err.Error() != expected {
		t.Error("Wrong error", err.Error())
	}
	if !mW.Matches(expected) {
		t.Error("Wrong message")
	}
}

func makeBigString(length int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	asRuneSlice := make([]rune, length)
	for index := range asRuneSlice {
		asRuneSlice[index] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(asRuneSlice)
}

func TestTrackValidators(t *testing.T) {
	client := getClient()
	// Empty key
	expectedTrack(client.Track("", "trafficType", "eventType", nil, nil), "Track: you passed an empty key, key must be a non-empty string", t)

	// Long key
	expectedTrack(client.Track(getLongKey(), "trafficType", "eventType", nil, nil), "Track: key too long - must be 250 characters or less", t)

	// Empty event type
	expectedTrack(client.Track("key", "trafficType", "", nil, nil), "Track: you passed an empty event type, event type must be a non-empty string", t)

	// Not match regex
	expected := "Track: you passed //, event name must adhere to " +
		"the regular expression ^[a-zA-Z0-9][-_.:a-zA-Z0-9]{0,79}$. This means an event " +
		"name must be alphanumeric, cannot be more than 80 characters long, and can " +
		"only include a dash, underscore, period, or colon as separators of " +
		"alphanumeric characters"
	expectedTrack(client.Track("key", "trafficType", "//", nil, nil), expected, t)

	// Empty traffic type
	expectedTrack(client.Track("key", "", "eventType", nil, nil), "Track: you passed an empty traffic type, traffic type must be a non-empty string", t)

	// Not matching traffic type
	expected = "Track: traffic type traffic does not have any corresponding feature flags in this environment, make sure you’re tracking your events to a valid traffic type defined in the Split user interface"
	expectedTrack(client.Track("key", "traffic", "eventType", nil, nil), expected, t)

	// Uppercase traffic type
	expectedTrack(client.Track("key", "traficTYPE", "eventType", nil, nil), "Track: traffic type should be all lowercase - converting string to lowercase", t)

	// Traffic Type No Ocurrences
	err := client.Track("key", "trafficTypeNoOcurrences", "eventType", nil, nil)
	if !mW.Matches("Track: traffic type traffictypenoocurrences does not have any corresponding feature flags in this environment, make sure you’re tracking your events to a valid traffic type defined in the Split user interface") {
		t.Error("Wrong message")
	}
	if err != nil {
		t.Error("Should not be error")
	}

	// Value
	expectedTrack(client.Track("key", "traffic", "eventType", true, nil), "Track: value must be a number", t)

	// Properties
	props := make(map[string]interface{})
	for i := 0; i < 301; i++ {
		props[fmt.Sprintf("prop-%d", i)] = "asd"
	}
	expectedTrack(client.Track("key", "traffic", "eventType", 1, props), "Track: Event has more than 300 properties. Some of them will be trimmed when processed", t)

	// Properties > 32kb
	props2 := make(map[string]interface{})
	for i := 0; i < 299; i++ {
		props2[fmt.Sprintf("%s%d", makeBigString(255), i)] = makeBigString(255)
	}
	expectedTrack(client.Track("key", "traffic", "eventType", nil, props2), "The maximum size allowed for the properties is 32kb. Event not queued", t)

	// Ok
	err = client.Track("key", "traffic", "eventType", 1, nil)

	if err != nil {
		t.Error("Should not return error")
	}
}

func TestLocalhostTrafficType(t *testing.T) {
	sdkConf := conf.Default()
	sdkConf.SplitFile = "../../testdata/splits.yaml"
	factory, _ := NewSplitFactory(conf.Localhost, sdkConf)
	client := factory.Client()

	_ = client.BlockUntilReady(1)

	factory.status.Store(sdkStatusInitializing)

	if client.isReady() {
		t.Error("Localhost should not be ready")
	}

	err := client.Track("key", "traffic", "eventType", nil, nil)

	if err != nil {
		t.Error("It should not inform any err")
	}

	mW.Reset()
	if mW.Length() > 0 {
		t.Error("Wrong message")
	}
	mW.Reset()
}

func TestInMemoryFactoryFlagSets(t *testing.T) {
	var splitsMock, _ = ioutil.ReadFile("../../testdata/splits_mock.json")
	var splitMock, _ = ioutil.ReadFile("../../testdata/split_mock.json")

	postChannel := make(chan string, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/splitChanges":
			if r.RequestURI != "/splitChanges?s=1.1&since=-1&sets=a%2Cc%2Cd" {
				t.Error("wrong RequestURI for flag sets")
			}
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
		case "/metrics/config":
			rBody, _ := ioutil.ReadAll(r.Body)
			var dataInPost dtos.Config
			err := json.Unmarshal(rBody, &dataInPost)
			if err != nil {
				t.Error(err)
				return
			}
			if dataInPost.FlagSetsInvalid != 4 {
				t.Error("invalid flag sets should be 4")
			}
			if dataInPost.FlagSetsTotal != 7 {
				t.Error("total flag sets should be 7")
			}
		default:
			fmt.Fprintln(w, "ok")
			return
		}
	}))
	defer ts.Close()
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	cfg.IPAddressesEnabled = true
	cfg.Advanced.EventsURL = ts.URL
	cfg.Advanced.SdkURL = ts.URL
	cfg.Advanced.TelemetryServiceURL = ts.URL
	cfg.Advanced.AuthServiceURL = ts.URL
	cfg.Advanced.ImpressionListener = &ImpressionListenerTest{}
	cfg.TaskPeriods.ImpressionSync = 60
	cfg.TaskPeriods.EventsSync = 60
	cfg.Advanced.StreamingEnabled = false
	cfg.Advanced.FlagSetsFilter = []string{"a", "_b", "a", "a", "c", "d", "_d"}

	factory, _ := NewSplitFactory("test", cfg)
	client := factory.Client()
	errBlock := client.BlockUntilReady(15)

	if errBlock != nil {
		t.Error("client should be ready")
	}

	if !client.isReady() {
		t.Error("InMemory should be ready")
	}

	mW.Reset()
	if mW.Length() > 0 {
		t.Error("Wrong message")
	}
	mW.Reset()

	client.Destroy()
}

func TestConsumerFactoryFlagSets(t *testing.T) {
	logger := getMockedLogger()
	sdkConf := conf.Default()
	sdkConf.OperationMode = conf.RedisConsumer
	sdkConf.Advanced.FlagSetsFilter = []string{"a", "b"}
	sdkConf.Logger = logger

	factory, _ := NewSplitFactory("something", sdkConf)
	if !mW.Matches("FlagSets filter is not applicable for Consumer modes where the SDK does not keep rollout data in sync. FlagSet filter was discarded") {
		t.Error("Wrong message")
	}
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

	client.Destroy()
}

func TestNotReadyYet(t *testing.T) {
	nonReadyUsages := 0
	logger := getMockedLogger()
	telemetryStorage := mocks.MockTelemetryStorage{
		RecordNonReadyUsageCall: func() {
			nonReadyUsages++
		},
		RecordExceptionCall: func(method string) {},
	}
	factoryNotReady := &SplitFactory{}
	clientNotReady := SplitClient{
		evaluator:   &mockEvaluator{},
		impressions: mutexqueue.NewMQImpressionsStorage(5000, make(chan string, 1), logger, mocks.MockTelemetryStorage{}),
		logger:      logger,
		validator: inputValidation{
			logger:       logger,
			splitStorage: mocks.MockSplitStorage{},
		},
		events: mocks.MockEventStorage{
			PushCall: func(event dtos.EventDTO, size int) error { return nil },
		},
		factory:             factoryNotReady,
		initTelemetry:       telemetryStorage,
		evaluationTelemetry: telemetryStorage,
	}
	flagSetFilter := flagsets.NewFlagSetFilter([]string{})
	maganerNotReady := SplitManager{
		initTelemetry: telemetryStorage,
		factory:       factoryNotReady,
		logger:        logger,
		splitStorage:  mutexmap.NewMMSplitStorage(flagSetFilter),
	}

	factoryNotReady.status.Store(sdkStatusInitializing)

	expectedMessage := "{operation}: the SDK is not ready, results may be incorrect. Make sure to wait for SDK readiness before using this method"
	expectedMessage1 := "{operation}: the SDK is not ready, results may be incorrect for feature flag feature. Make sure to wait for SDK readiness before using this method"
	expectedMessage2 := "{operation}: the SDK is not ready, results may be incorrect for feature flags feature, feature_2. Make sure to wait for SDK readiness before using this method"

	clientNotReady.Treatment("test", "feature", nil)
	if !mW.Matches(strings.Replace(expectedMessage1, "{operation}", "Treatment", 1)) {
		t.Error("Wrong message")
	}

	clientNotReady.Treatments("test", []string{"feature", "feature_2"}, nil)
	if !mW.Matches(strings.Replace(expectedMessage2, "{operation}", "Treatments", 1)) {
		t.Error("Wrong message")
	}

	clientNotReady.TreatmentWithConfig("test", "feature", nil)
	if !mW.Matches(strings.Replace(expectedMessage1, "{operation}", "TreatmentWithConfig", 1)) {
		t.Error("Wrong message")
	}

	clientNotReady.TreatmentsWithConfig("test", []string{"feature", "feature_2"}, nil)
	if !mW.Matches(strings.Replace(expectedMessage2, "{operation}", "TreatmentsWithConfig", 1)) {
		t.Error("Wrong message", mW.messages)
	}
	expected := "Track: the SDK is not ready, results may be incorrect. Make sure to wait for SDK readiness before using this method"
	expectedTrack(clientNotReady.Track("key", "traffic", "eventType", nil, nil), expected, t)

	maganerNotReady.Split("feature")
	if !mW.Matches(strings.Replace(expectedMessage, "{operation}", "Split", 1)) {
		t.Error("Wrong message")
	}

	maganerNotReady.Splits()
	if !mW.Matches(strings.Replace(expectedMessage, "{operation}", "Splits", 1)) {
		t.Error("Wrong message")
	}

	maganerNotReady.SplitNames()
	if !mW.Matches(strings.Replace(expectedMessage, "{operation}", "SplitNames", 1)) {
		t.Error("Wrong message")
	}

	if nonReadyUsages != 8 {
		t.Error("It should track a non ready usage")
	}
}

func TestManagerWithEmptySplit(t *testing.T) {
	flagSetFilter := flagsets.NewFlagSetFilter([]string{})
	splitStorage := mutexmap.NewMMSplitStorage(flagSetFilter)
	factory := SplitFactory{}
	manager := SplitManager{
		splitStorage: splitStorage,
		logger:       getMockedLogger(),
	}

	factory.status.Store(sdkStatusReady)
	manager.factory = &factory

	manager.Split("")
	if !mW.Matches("Split: you passed an empty featureFlagName, flag name must be a non-empty string") {
		t.Error("Wrong message")
	}

	manager.Split("non_existent")
	if !mW.Matches("Split: you passed non_existent that does not exist in this environment, please double check what feature flags exist in the Split user interface.") {
		t.Error("Wrong message")
	}
}
