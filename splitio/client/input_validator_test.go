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
	"os"
	"testing"
	"time"

	"github.com/splitio/go-client/v6/splitio/conf"

	commonsCfg "github.com/splitio/go-split-commons/v9/conf"
	"github.com/splitio/go-split-commons/v9/dtos"
	"github.com/splitio/go-split-commons/v9/flagsets"
	"github.com/splitio/go-split-commons/v9/healthcheck/application"
	"github.com/splitio/go-split-commons/v9/provisional"
	"github.com/splitio/go-split-commons/v9/provisional/strategy"
	"github.com/splitio/go-split-commons/v9/service/api"
	authMocks "github.com/splitio/go-split-commons/v9/service/mocks"
	"github.com/splitio/go-split-commons/v9/storage/filter"
	"github.com/splitio/go-split-commons/v9/storage/inmemory/mutexmap"
	"github.com/splitio/go-split-commons/v9/storage/inmemory/mutexqueue"
	"github.com/splitio/go-split-commons/v9/storage/mocks"
	"github.com/splitio/go-split-commons/v9/storage/redis"
	"github.com/splitio/go-split-commons/v9/synchronizer"
	"github.com/splitio/go-toolkit/v5/logging"
	lMock "github.com/splitio/go-toolkit/v5/logging/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getClient(logger logging.LoggerInterface) SplitClient {
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
	loggerMock := &lMock.LoggerMock{}
	loggerMock.On("Error", []interface{}{"factory instantiation: you passed an empty SDK key, SDK key must be a non-empty string"}).Return()
	cfg.Logger = loggerMock

	_, err := NewSplitFactory("", cfg)
	assert.NotNil(t, err)

	loggerMock.AssertExpectations(t)
}

func getLongKey() string {
	m := ""
	for n := 0; n <= 256; n++ {
		m += "m"
	}
	return m
}

func TestValidationEmpty(t *testing.T) {
	loggerMock := &lMock.LoggerMock{}
	client := getClient(loggerMock)
	expectedTreatment(client.Treatment("key", "feature", nil), "TreatmentA", t)
	loggerMock.AssertExpectations(t)
}

func TestTreatmentValidatorOnKeys(t *testing.T) {
	loggerMock := &lMock.LoggerMock{}
	client := getClient(loggerMock)

	// Nil
	loggerMock.On("Error", []interface{}{"Treatment: you passed a nil key, key must be a non-empty string"}).Return().Once()
	assert.Equal(t, "control", client.Treatment(nil, "feature", nil))

	// Boolean
	loggerMock.On("Error", []interface{}{"Treatment: you passed an invalid key, key must be a non-empty string"}).Return().Once()
	assert.Equal(t, "control", client.Treatment(true, "feature", nil))

	// Trimmed
	loggerMock.On("Error", []interface{}{"Treatment: you passed an empty key, key must be a non-empty string"}).Return().Once()
	assert.Equal(t, client.Treatment("     ", "feature", nil), "control")

	// Long
	loggerMock.On("Error", []interface{}{"Treatment: key too long - must be 250 characters or less"}).Return().Once()
	assert.Equal(t, client.Treatment(getLongKey(), "feature", nil), "control")

	// Int
	loggerMock.On("Warning", []interface{}{"Treatment: key 123 is not of type string, converting"}).Return().Once()
	assert.Equal(t, "TreatmentA", client.Treatment(123, "feature", nil))

	// Int32
	loggerMock.On("Warning", []interface{}{"Treatment: key 123 is not of type string, converting"}).Return().Once()
	assert.Equal(t, "TreatmentA", client.Treatment(int32(123), "feature", nil))

	// Int 64
	loggerMock.On("Warning", []interface{}{"Treatment: key 123 is not of type string, converting"}).Return().Once()
	assert.Equal(t, "TreatmentA", client.Treatment(int64(123), "feature", nil))

	// Float
	loggerMock.On("Warning", []interface{}{"Treatment: key 1.3 is not of type string, converting"}).Return().Once()
	assert.Equal(t, "TreatmentA", client.Treatment(1.3, "feature", nil))

	// NaN
	loggerMock.On("Error", []interface{}{"Treatment: you passed an invalid key, key must be a non-empty string"}).Return().Once()
	assert.Equal(t, "control", client.Treatment(math.NaN(), "feature", nil))

	// Inf
	loggerMock.On("Error", []interface{}{"Treatment: you passed an invalid key, key must be a non-empty string"}).Return().Once()
	assert.Equal(t, "control", client.Treatment(math.Inf, "feature", nil))

	loggerMock.AssertExpectations(t)
}

func getKey(matchingKey string, bucketingKey string) *Key {
	return &Key{
		MatchingKey:  matchingKey,
		BucketingKey: bucketingKey,
	}
}

func TestTreatmentValidatorWithKeyObject(t *testing.T) {
	loggerMock := &lMock.LoggerMock{}
	client := getClient(loggerMock)

	// Empty
	loggerMock.On("Error", []interface{}{"Treatment: you passed an empty matchingKey, matchingKey must be a non-empty string"}).Return().Once()
	assert.Equal(t, "control", client.Treatment(getKey("", "bucketing"), "feature", nil))

	// Long
	loggerMock.On("Error", []interface{}{"Treatment: matchingKey too long - must be 250 characters or less"}).Return().Once()
	assert.Equal(t, "control", client.Treatment(getKey(getLongKey(), "bucketing"), "feature", nil))

	// Empty Bucketing
	loggerMock.On("Error", []interface{}{"Treatment: you passed an empty bucketingKey, bucketingKey must be a non-empty string"}).Return().Once()
	assert.Equal(t, "control", client.Treatment(getKey("matching", ""), "feature", nil))

	// Long Bucketing
	loggerMock.On("Error", []interface{}{"Treatment: bucketingKey too long - must be 250 characters or less"}).Return().Once()
	assert.Equal(t, "control", client.Treatment(getKey("matching", getLongKey()), "feature", nil))

	// Ok
	assert.Equal(t, "TreatmentA", client.Treatment(getKey("matching", "bucketing"), "feature", nil))

	loggerMock.AssertExpectations(t)
}

func TestTreatmentValidatorOnFeatureName(t *testing.T) {
	loggerMock := &lMock.LoggerMock{}
	client := getClient(loggerMock)

	// Empty
	loggerMock.On("Error", []interface{}{"Treatment: you passed an empty featureFlagName, flag name must be a non-empty string"}).Return().Once()
	assert.Equal(t, "control", client.Treatment("key", "", nil))

	// Trimmed
	loggerMock.On("Warning", []interface{}{"Treatment: featureFlagName '  feature   ' has extra whitespace, trimming"}).Return().Once()
	assert.Equal(t, "TreatmentA", client.Treatment("key", "  feature   ", nil))

	// Non Existent
	loggerMock.On("Warning", []interface{}{"Treatment: you passed feature_non_existent that does not exist in this environment, please double check what feature flags exist in the Split user interface."}).Return().Once()
	assert.Equal(t, "control", client.Treatment("key", "feature_non_existent", nil))

	// Non Existent
	loggerMock.On("Warning", []interface{}{"TreatmentWithConfig: you passed feature_non_existent that does not exist in this environment, please double check what feature flags exist in the Split user interface."}).Return().Once()
	results := client.TreatmentWithConfig("key", "feature_non_existent", nil)
	assert.Equal(t, "control", results.Treatment)
	assert.Nil(t, results.Config)

	loggerMock.AssertExpectations(t)
}

func TestTreatmentsValidator(t *testing.T) {
	logger := &lMock.LoggerMock{}
	client := getClient(logger)

	// Empty features
	logger.On("Error", []interface{}{"Treatments: you passed an empty featureFlagName, flag name must be a non-empty string"}).Return().Once()
	logger.On("Error", []interface{}{"Treatments: featureFlagNames must be a non-empty array"}).Return().Once()
	assert.Equal(t, map[string]string{}, client.Treatments("key", []string{""}, nil))

	// Inf
	logger.On("Error", []interface{}{"Treatments: you passed an invalid key, key must be a non-empty string"}).Return().Once()
	assert.Equal(t, map[string]string{"feature": "control"}, client.Treatments(math.Inf, []string{"feature"}, nil))

	// Float
	logger.On("Warning", []interface{}{"Treatments: key 1.3 is not of type string, converting"}).Return().Once()
	assert.Equal(t, map[string]string{"feature": "TreatmentA"}, client.Treatments(1.3, []string{"feature"}, nil))

	// Trimmed
	logger.On("Warning", []interface{}{"Treatments: featureFlagName ' some_feature  ' has extra whitespace, trimming"}).Return().Once()
	assert.Equal(t, map[string]string{"some_feature": "control"}, client.Treatments("key", []string{" some_feature  "}, nil))

	// Non Existent
	logger.On("Warning", []interface{}{"Treatments: you passed feature_non_existent that does not exist in this environment, please double check what feature flags exist in the Split user interface."}).Return().Once()
	assert.Equal(t, map[string]string{"feature_non_existent": "control"}, client.Treatments("key", []string{"feature_non_existent"}, nil))

	// Non Existent Config
	logger.On("Warning", []interface{}{"TreatmentsWithConfig: you passed feature_non_existent that does not exist in this environment, please double check what feature flags exist in the Split user interface."}).Return().Once()
	result := client.TreatmentsWithConfig("key", []string{"feature_non_existent"}, nil)
	assert.Equal(t, map[string]TreatmentResult{"feature_non_existent": {Treatment: "control", Config: nil}}, result)

	logger.AssertExpectations(t)
}

func TestValidatorOnDestroy(t *testing.T) {
	telemetryMockedStorage := mocks.MockTelemetryStorage{
		RecordSessionLengthCall: func(session int64) {},
	}
	logger := &lMock.LoggerMock{}
	localConfig := &synchronizer.LocalConfig{RefreshEnabled: false}
	sync, _ := synchronizer.NewSynchronizerManager(
		synchronizer.NewLocal(localConfig, &api.SplitAPI{}, mocks.MockSplitStorage{}, mocks.MockSegmentStorage{}, nil, &mocks.MockRuleBasedSegmentStorage{}, logger, telemetryMockedStorage, &application.Dummy{}),
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

	logger.On("Info", mock.Anything).Return()
	client2.Destroy()

	logger.On("Error", []interface{}{"Client has already been destroyed - no calls possible"}).Return().Once()
	assert.Equal(t, "control", client2.Treatment("key", "  feature   ", nil))

	logger.On("Error", []interface{}{"Client has already been destroyed - no calls possible"}).Return().Once()
	result := client2.Treatments("key", []string{"some_feature"}, nil)
	assert.Equal(t, map[string]string{"some_feature": "control"}, result)

	logger.On("Error", []interface{}{"Client has already been destroyed - no calls possible"}).Return().Once()
	assert.ErrorContains(t, client2.Track("key", "trafficType", "eventType", 0, nil), "Client has already been destroyed - no calls possible")

	logger.On("Error", []interface{}{"Client has already been destroyed - no calls possible"}).Return().Once()
	manager.Split("feature")

	logger.AssertExpectations(t)
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
	logger := &lMock.LoggerMock{}
	client := getClient(logger)
	// Empty key
	logger.On("Error", []interface{}{"Track: you passed an empty key, key must be a non-empty string"}).Return().Once()
	assert.ErrorContains(t, client.Track("", "trafficType", "eventType", nil, nil), "Track: you passed an empty key, key must be a non-empty string")

	// Long key
	logger.On("Error", []interface{}{"Track: key too long - must be 250 characters or less"}).Return().Once()
	assert.ErrorContains(t, client.Track(getLongKey(), "trafficType", "eventType", nil, nil), "Track: key too long - must be 250 characters or less")

	// Empty event type
	logger.On("Error", []interface{}{"Track: you passed an empty event type, event type must be a non-empty string"}).Return().Once()
	assert.ErrorContains(t, client.Track("key", "trafficType", "", nil, nil), "Track: you passed an empty event type, event type must be a non-empty string")

	// Not match regex
	expected := "Track: you passed //, event name must adhere to " +
		"the regular expression ^[a-zA-Z0-9][-_.:a-zA-Z0-9]{0,79}$. This means an event " +
		"name must be alphanumeric, cannot be more than 80 characters long, and can " +
		"only include a dash, underscore, period, or colon as separators of " +
		"alphanumeric characters"
	logger.On("Error", []interface{}{expected}).Return().Once()
	assert.ErrorContains(t, client.Track("key", "trafficType", "//", nil, nil), expected)

	// Empty traffic type
	logger.On("Error", []interface{}{"Track: you passed an empty traffic type, traffic type must be a non-empty string"}).Return().Once()
	assert.ErrorContains(t, client.Track("key", "", "eventType", nil, nil), "Track: you passed an empty traffic type, traffic type must be a non-empty string")

	// Not matching traffic type
	expected = "Track: traffic type traffic does not have any corresponding feature flags in this environment, make sure you’re tracking your events to a valid traffic type defined in the Split user interface"
	logger.On("Warning", []interface{}{expected}).Return().Once()
	assert.Nil(t, client.Track("key", "traffic", "eventType", nil, nil), expected)

	// Uppercase traffic type
	logger.On("Warning", []interface{}{"Track: traffic type should be all lowercase - converting string to lowercase"}).Return().Once()
	assert.Nil(t, client.Track("key", "traficTYPE", "eventType", nil, nil))

	// Traffic Type No Ocurrences
	logger.On("Warning", []interface{}{"Track: traffic type traffictypenoocurrences does not have any corresponding feature flags in this environment, make sure you’re tracking your events to a valid traffic type defined in the Split user interface"}).Return().Once()
	assert.Nil(t, client.Track("key", "traffictypenoocurrences", "eventType", nil, nil))

	// Value
	logger.On("Warning", mock.Anything).Return().Once()
	logger.On("Error", []interface{}{"Track: value must be a number"}).Return().Once()
	assert.ErrorContains(t, client.Track("key", "traffic", "eventType", true, nil), "Track: value must be a number", t)

	// Properties
	props := make(map[string]interface{})
	for i := 0; i < 301; i++ {
		props[fmt.Sprintf("prop-%d", i)] = "asd"
	}
	logger.On("Warning", mock.Anything).Return().Once()
	logger.On("Warning", []interface{}{"Track: Event has more than 300 properties. Some of them will be trimmed when processed"}).Return().Once()
	assert.Nil(t, client.Track("key", "traffic", "eventType", 1, props))

	// Properties > 32kb
	props2 := make(map[string]interface{})
	for i := 0; i < 299; i++ {
		props2[fmt.Sprintf("%s%d", makeBigString(255), i)] = makeBigString(255)
	}
	logger.On("Warning", mock.Anything).Return().Once()
	logger.On("Error", []interface{}{"Track: The maximum size allowed for the properties is 32kb. Event not queued"}).Return().Once()
	assert.ErrorContains(t, client.Track("key", "traffic", "eventType", nil, props2), "The maximum size allowed for the properties is 32kb. Event not queued", t)

	// Ok
	logger.On("Warning", mock.Anything).Return().Once()
	assert.Nil(t, client.Track("key", "trafictype", "eventType", nil, nil))
	assert.Nil(t, client.Track("key", "traffic", "eventType", 1, nil))

	logger.AssertExpectations(t)
}

func TestLocalhostTrafficType(t *testing.T) {
	sdkConf := conf.Default()
	sdkConf.SplitFile = "../../testdata/splits.yaml"
	factory, _ := NewSplitFactory(conf.Localhost, sdkConf)
	client := factory.Client()

	_ = client.BlockUntilReady(1)

	factory.status.Store(sdkStatusInitializing)

	assert.False(t, client.isReady(), "Localhost should not be ready")
	assert.Nil(t, client.Track("key", "traffic", "eventType", nil, nil))

	client.Destroy()
}

func TestInMemoryFactoryFlagSets(t *testing.T) {
	var splitsMock, _ = os.ReadFile("../../testdata/splits_mock.json")
	var splitMock, _ = os.ReadFile("../../testdata/split_mock.json")

	postChannel := make(chan string, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/splitChanges":
			assert.Equal(t, "/splitChanges?s=1.3&since=-1&rbSince=-1&sets=a%2Cc%2Cd", r.RequestURI)
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
			assert.NotSubset(t, r.Header, map[string][]string{"SplitSDKMachineIP": {}, "SplitSDKMachineName": {}})

			rBody, _ := ioutil.ReadAll(r.Body)
			var dataInPost []map[string]interface{}
			err := json.Unmarshal(rBody, &dataInPost)
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, len(dataInPost), 1)
			fmt.Fprintln(w, "ok")
			postChannel <- "finished"
		case "/segmentChanges":
		case "/metrics/config":
			rBody, _ := ioutil.ReadAll(r.Body)
			var dataInPost dtos.Config
			err := json.Unmarshal(rBody, &dataInPost)
			assert.Nil(t, err)
			assert.Equal(t, int64(4), dataInPost.FlagSetsInvalid, "invalid flag sets should be 4")
			assert.Equal(t, int64(7), dataInPost.FlagSetsTotal, "total flag sets should be 7")
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
	assert.Nil(t, errBlock, "client should be ready")
	assert.True(t, client.isReady(), "InMemory should be ready")

	client.Destroy()
}

func TestConsumerFactoryFlagSets(t *testing.T) {
	logger := &lMock.LoggerMock{}
	sdkConf := conf.Default()
	sdkConf.OperationMode = conf.RedisConsumer
	sdkConf.Advanced.FlagSetsFilter = []string{"a", "b"}
	sdkConf.Logger = logger

	logger.On("Debug", mock.Anything)
	logger.On("Warning", []interface{}{"Factory Instantiation: You already have an instance of the Split factory. Make sure you definitely want this additional instance. We recommend keeping only one instance of the factory at all times (Singleton pattern) and reusing it throughout your application."})
	logger.On("Warning", []interface{}{"FlagSets filter is not applicable for Consumer modes where the SDK does not keep rollout data in sync. FlagSet filter was discarded"}).Return().Once()
	factory, _ := NewSplitFactory("something", sdkConf)
	assert.True(t, factory.IsReady())
	client := factory.Client()
	assert.True(t, factory.IsReady())

	err := client.BlockUntilReady(1)
	assert.Nil(t, err)

	manager := factory.Manager()
	assert.True(t, manager.factory.IsReady())
	err = manager.BlockUntilReady(1)
	assert.Nil(t, err)

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
	logger := &lMock.LoggerMock{}
	telemetryStorage := mocks.MockTelemetryStorage{
		RecordLatencyCall: func(method string, latency time.Duration) {},
		RecordNonReadyUsageCall: func() {
			nonReadyUsages++
		},
		RecordExceptionCall: func(method string) {},
	}
	impressionsCounter := strategy.NewImpressionsCounter()
	filter := filter.NewBloomFilter(100, 0.01)
	uniqueKeysTracker := strategy.NewUniqueKeysTracker(filter)
	noneStrategy := strategy.NewNoneImpl(impressionsCounter, uniqueKeysTracker, false)
	impressionManager := provisional.NewImpressionManagerImp(noneStrategy, noneStrategy)
	factoryNotReady := &SplitFactory{cfg: conf.Default(), impressionManager: impressionManager}
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
		impressionManager:   impressionManager,
	}
	flagSetFilter := flagsets.NewFlagSetFilter([]string{})
	managerNotReady := SplitManager{
		initTelemetry: telemetryStorage,
		factory:       factoryNotReady,
		logger:        logger,
		splitStorage:  mutexmap.NewMMSplitStorage(flagSetFilter),
	}

	factoryNotReady.status.Store(sdkStatusInitializing)

	logger.On("Warning", []interface{}{"Treatment: the SDK is not ready, results may be incorrect for feature flag feature. Make sure to wait for SDK readiness before using this method"}).Return().Once()
	assert.Equal(t, "control", clientNotReady.Treatment("test", "feature", nil))

	logger.On("Warning", []interface{}{"Treatments: the SDK is not ready, results may be incorrect for feature flags feature, feature_2. Make sure to wait for SDK readiness before using this method"}).Return().Once()
	assert.Equal(t, map[string]string{"feature": "control", "feature_2": "control"}, clientNotReady.Treatments("test", []string{"feature", "feature_2"}, nil))

	logger.On("Warning", []interface{}{"TreatmentWithConfig: the SDK is not ready, results may be incorrect for feature flag feature. Make sure to wait for SDK readiness before using this method"}).Return().Once()
	assert.Equal(t, TreatmentResult{Treatment: "control", Config: nil}, clientNotReady.TreatmentWithConfig("test", "feature", nil))

	logger.On("Warning", []interface{}{"TreatmentsWithConfig: the SDK is not ready, results may be incorrect for feature flags feature, feature_2. Make sure to wait for SDK readiness before using this method"}).Return().Once()
	assert.Equal(t, map[string]TreatmentResult{"feature": {Treatment: "control", Config: nil}, "feature_2": {Treatment: "control", Config: nil}}, clientNotReady.TreatmentsWithConfig("test", []string{"feature", "feature_2"}, nil))

	logger.On("Warning", []interface{}{"Track: the SDK is not ready, results may be incorrect. Make sure to wait for SDK readiness before using this method"}).Return().Once()
	assert.Nil(t, clientNotReady.Track("key", "traffic", "eventType", nil, nil))

	logger.On("Warning", []interface{}{"Split: the SDK is not ready, results may be incorrect. Make sure to wait for SDK readiness before using this method"}).Return().Once()
	logger.On("Error", []interface{}{"Split: you passed feature that does not exist in this environment, please double check what feature flags exist in the Split user interface."}).Return().Once()
	assert.Nil(t, managerNotReady.Split("feature"))

	logger.On("Warning", []interface{}{"SplitNames: the SDK is not ready, results may be incorrect. Make sure to wait for SDK readiness before using this method"}).Return().Once()
	assert.Len(t, managerNotReady.SplitNames(), 0)

	logger.On("Warning", []interface{}{"Splits: the SDK is not ready, results may be incorrect. Make sure to wait for SDK readiness before using this method"}).Return().Once()
	assert.Len(t, managerNotReady.Splits(), 0)

	assert.Equal(t, 8, nonReadyUsages, "It should track non ready usages")
	logger.AssertExpectations(t)
}

func TestManagerWithEmptySplit(t *testing.T) {
	flagSetFilter := flagsets.NewFlagSetFilter([]string{})
	splitStorage := mutexmap.NewMMSplitStorage(flagSetFilter)
	factory := SplitFactory{}
	logger := &lMock.LoggerMock{}
	manager := SplitManager{
		splitStorage: splitStorage,
		logger:       logger,
	}

	factory.status.Store(sdkStatusReady)
	manager.factory = &factory

	logger.On("Error", []interface{}{"Split: you passed an empty featureFlagName, flag name must be a non-empty string"}).Return().Once()
	assert.Nil(t, manager.Split(""))

	logger.On("Error", []interface{}{"Split: you passed non_existent that does not exist in this environment, please double check what feature flags exist in the Split user interface."}).Return().Once()
	assert.Nil(t, manager.Split("non_existent"))

	logger.AssertExpectations(t)
}
