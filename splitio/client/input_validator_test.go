package client

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-client/splitio/storage/mutexqueue"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
)

var strMsg string

type MockWriter struct {
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	strMsg = string(p[:])
	return 0, nil
}

var mW MockWriter
var options = &logging.LoggerOptions{
	LogLevel:      5,
	ErrorWriter:   &mW,
	WarningWriter: &mW,
	InfoWriter:    &mW,
	DebugWriter:   &mW,
	VerboseWriter: &mW,
}

type mockSplitStorage struct {
}

func (tt *mockSplitStorage) TrafficTypeExists(trafficType string) bool {
	switch trafficType {
	case "trafictype":
		return true
	default:
		return false
	}
}
func (tt *mockSplitStorage) Get(splitName string) *dtos.SplitDTO { return nil }
func (tt *mockSplitStorage) GetAll() []dtos.SplitDTO             { return []dtos.SplitDTO{} }
func (tt *mockSplitStorage) SplitNames() []string                { return []string{} }
func (tt *mockSplitStorage) SegmentNames() *set.ThreadUnsafeSet  { return nil }

var logger = logging.NewLogger(options)
var cfg = conf.Default()
var client = SplitClient{
	cfg:         cfg,
	evaluator:   &mockEvaluator{},
	impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
	metrics:     mutexmap.NewMMMetricsStorage(),
	logger:      logger,
	validator: inputValidation{
		logger:       logger,
		splitStorage: &mockSplitStorage{},
	},
	events: &mockEvents{},
}

var factory = SplitFactory{
	client: &client,
}

func init() {
	factory.status.Store(SdkReady)
	client.factory = &factory
}

func TestFactoryWithNilApiKey(t *testing.T) {
	cfg := conf.Default()
	cfg.Logger = logger
	_, err := NewSplitFactory("", cfg)

	if err == nil {
		t.Error("Should be error")
	}

	expected := "Factory instantiation: you passed an empty apikey, apikey must be a non-empty string"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithNilKey(t *testing.T) {
	result := client.Treatment(nil, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed a nil key, key must be a non-empty string"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithBooleanKey(t *testing.T) {
	result := client.Treatment(true, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an invalid key, key must be a non-empty string"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWitWhitespaceKey(t *testing.T) {
	result := client.Treatment("     ", "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an empty key, key must be a non-empty string"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithLengthKey(t *testing.T) {
	m := ""
	for n := 0; n <= 256; n++ {
		m += "m"
	}

	result := client.Treatment(m, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: key too long - must be 250 characters or less"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithStringKey(t *testing.T) {
	result := client.Treatment("key", "feature", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := ""
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithIntKey(t *testing.T) {
	result := client.Treatment(123, "feature", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := "Treatment: key %!s(int=123) is not of type string, converting"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithInt32Key(t *testing.T) {
	result := client.Treatment(int32(123), "feature", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := "Treatment: key %!s(int32=123) is not of type string, converting"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithInt64Key(t *testing.T) {
	result := client.Treatment(int64(123), "feature", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := "Treatment: key %!s(int64=123) is not of type string, converting"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWitFloatKey(t *testing.T) {
	result := client.Treatment(1.3, "feature", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := "Treatment: key %!s(float64=1.3) is not of type string, converting"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWitFloatNaNKey(t *testing.T) {
	result := client.Treatment(math.NaN, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an invalid key, key must be a non-empty string"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWitFloatInfKey(t *testing.T) {
	result := client.Treatment(math.Inf, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an invalid key, key must be a non-empty string"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithEmptyMatchingKey(t *testing.T) {
	var key = &Key{
		MatchingKey:  "",
		BucketingKey: "bucketing",
	}

	result := client.Treatment(key, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an empty matchingKey, matchingKey must be a non-empty string"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithLengthMatchingKey(t *testing.T) {
	m := ""
	for n := 0; n <= 256; n++ {
		m += "m"
	}

	var key = &Key{
		MatchingKey:  m,
		BucketingKey: "bucketing",
	}

	result := client.Treatment(key, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: matchingKey too long - must be 250 characters or less"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithEmptyBuckeitngKey(t *testing.T) {
	var key = &Key{
		MatchingKey:  "matching",
		BucketingKey: "",
	}

	result := client.Treatment(key, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an empty bucketingKey, bucketingKey must be a non-empty string"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithLengthBucketingKey(t *testing.T) {
	m := ""
	for n := 0; n <= 256; n++ {
		m += "m"
	}

	var key = &Key{
		MatchingKey:  "matching",
		BucketingKey: m,
	}

	result := client.Treatment(key, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: bucketingKey too long - must be 250 characters or less"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithKeyObject(t *testing.T) {
	var key = &Key{
		MatchingKey:  "matching",
		BucketingKey: "bucketing",
	}

	result := client.Treatment(key, "feature", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := ""
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorEmptyFeatureName(t *testing.T) {
	result := client.Treatment("key", "", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an empty featureName, featureName must be a non-empty string"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWhitespacesFeatureName(t *testing.T) {
	result := client.Treatment("key", "  feature   ", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := "Treatment: split name '  feature   ' has extra whitespace, trimming"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentValidatorWithFeatureNonExistant(t *testing.T) {
	result := client.Treatment("key", "feature_non_existant", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed feature_non_existant that does not exist in this environment, please double check what Splits exist in the web console."
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentConfigValidatorWithFeatureNonExistant(t *testing.T) {
	result := client.TreatmentWithConfig("key", "feature_non_existant", nil)

	if result.Treatment != "control" {
		t.Error("Should be control")
	}

	expected := "TreatmentWithConfig: you passed feature_non_existant that does not exist in this environment, please double check what Splits exist in the web console."
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentClientDestroyed(t *testing.T) {

	var client2 = SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
		metrics:     mutexmap.NewMMMetricsStorage(),
		logger:      logger,
		validator:   inputValidation{logger: logger},
		sync: &sdkSync{
			countersSync:   nil,
			gaugeSync:      nil,
			impressionSync: nil,
			latenciesSync:  nil,
			segmentSync:    nil,
			splitSync:      nil,
		},
	}

	factory := SplitFactory{
		client: &client2,
	}

	client2.factory = &factory

	client2.Destroy()

	result := client2.Treatment("key", "  feature   ", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Client has already been destroyed - no calls possible"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentsEmptyFeatures(t *testing.T) {
	features := []string{""}

	result := client.Treatments("key", features, nil)

	if len(result) != 0 {
		t.Error("Should not have elements")
	}

	expected := "Treatments: features must be a non-empty array"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentsValidatorWitFloatInfKey(t *testing.T) {
	features := []string{"feature"}

	result := client.Treatments(math.Inf, features, nil)

	if len(result) != 1 {
		t.Error("Should return values")
	}

	if result["feature"] != "control" {
		t.Error("Should return feature with control value")
	}

	expected := "Treatments: you passed an invalid key, key must be a non-empty string"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmenstValidatorWitFloatKey(t *testing.T) {
	features := []string{"feature"}

	result := client.Treatments(1.3, features, nil)

	if len(result) != 1 {
		t.Error("Should return elements")
	}

	if result["feature"] != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := "Treatments: key %!s(float64=1.3) is not of type string, converting"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentsWhitespaceFeatures(t *testing.T) {
	features := []string{" some_feature  "}

	result := client.Treatments("key", features, nil)

	if result["some_feature"] != "control" {
		t.Error("Wrong result")
	}

	expected := "Treatments: split name ' some_feature  ' has extra whitespace, trimming"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentsValidatorWithFeatureNonExistant(t *testing.T) {
	result := client.Treatments("key", []string{"feature_non_existant"}, nil)

	if result["feature_non_existant"] != "control" {
		t.Error("Should be control")
	}

	expected := "Treatments: you passed feature_non_existant that does not exist in this environment, please double check what Splits exist in the web console."
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}
func TestTreatmentsConfigValidatorWithFeatureNonExistant(t *testing.T) {
	result := client.TreatmentsWithConfig("key", []string{"feature_non_existant"}, nil)

	if result["feature_non_existant"].Treatment != "control" {
		t.Error("Should be control")
	}

	expected := "TreatmentsWithConfig: you passed feature_non_existant that does not exist in this environment, please double check what Splits exist in the web console."
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTreatmentsClientDestroyed(t *testing.T) {

	var client2 = SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
		metrics:     mutexmap.NewMMMetricsStorage(),
		logger:      logger,
		validator:   inputValidation{logger: logger},
		sync: &sdkSync{
			countersSync:   nil,
			gaugeSync:      nil,
			impressionSync: nil,
			latenciesSync:  nil,
			segmentSync:    nil,
			splitSync:      nil,
		},
	}

	factory := SplitFactory{
		client: &client2,
	}

	client2.factory = &factory

	client2.Destroy()

	features := []string{"some_feature"}

	result := client2.Treatments("key", features, nil)

	if result == nil {
		t.Error("Should return control treatments")
	}

	if result["some_feature"] != evaluator.Control {
		t.Error("Wrong treatment result")
	}

	expected := "Client has already been destroyed - no calls possible"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTrackValidatorWithEmptyKey(t *testing.T) {
	err := client.Track("", "trafficType", "eventType", nil, nil)

	expected := "Track: you passed an empty key, key must be a non-empty string"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTrackValidatorWithLengthKey(t *testing.T) {
	m := ""
	for n := 0; n <= 256; n++ {
		m += "m"
	}

	err := client.Track(m, "trafficType", "eventType", nil, nil)

	expected := "Track: key too long - must be 250 characters or less"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTrackValidatorWithEmptyEventName(t *testing.T) {
	err := client.Track("key", "trafficType", "", nil, nil)

	expected := "Track: you passed an empty event type, event type must be a non-empty string"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTrackValidatorWithNotConformEventName(t *testing.T) {
	err := client.Track("key", "trafficType", "//", nil, nil)

	expected := "Track: you passed //, event name must adhere to " +
		"the regular expression ^[a-zA-Z0-9][-_.:a-zA-Z0-9]{0,79}$. This means an event " +
		"name must be alphanumeric, cannot be more than 80 characters long, and can " +
		"only include a dash, underscore, period, or colon as separators of " +
		"alphanumeric characters"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTrackValidatorWithEmptyTrafficType(t *testing.T) {
	err := client.Track("key", "", "eventType", nil, nil)

	expected := "Track: you passed an empty traffic type, traffic type must be a non-empty string"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestLocalhostTrafficType(t *testing.T) {
	sdkConf := conf.Default()
	sdkConf.SplitFile = "../../testdata/splits.yaml"
	factory, _ := NewSplitFactory("localhost", sdkConf)
	client := factory.Client()

	_ = client.BlockUntilReady(1)

	factory.status.Store(SdkOnInitialization)

	if client.isReady() {
		t.Error("Localhost should not be ready")
	}

	err := client.Track("key", "traffic", "eventType", nil, nil)

	if err != nil {
		t.Error("It should not inform any err")
	}

	if len(strMsg) > 0 {
		t.Error("It should not inform any log")
	}
}

func TestTrafficTypeOnReady(t *testing.T) {
	err := client.Track("key", "traffic", "eventType", nil, nil)

	expected := "Track: traffic type traffic does not have any corresponding Splits in this environment, make sure you’re tracking your events to a valid traffic type defined in the Split console"
	if err != nil {
		t.Error("Wrong error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
}
func TestTrackNotReadyYetTrafficType(t *testing.T) {
	var clientNotReady = SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
		metrics:     mutexmap.NewMMMetricsStorage(),
		logger:      logger,
		validator: inputValidation{
			logger:       logger,
			splitStorage: &mockSplitStorage{},
		},
		events: &mockEvents{},
	}

	var factoryNotReady = SplitFactory{
		client: &clientNotReady,
	}

	factoryNotReady.status.Store(SdkOnInitialization)
	clientNotReady.factory = &factoryNotReady

	err := clientNotReady.Track("key", "traffic", "eventType", nil, nil)

	expected := "Track: the SDK is not ready, results may be incorrect. Make sure to wait for SDK readiness before using this method"
	if err != nil {
		t.Error("Wrong error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
}

func TestTrackValidatorWithUpperCaseTrafficType(t *testing.T) {
	err := client.Track("key", "traficTYPE", "eventType", nil, nil)

	expected := "Track: traffic type should be all lowercase - converting string to lowercase"
	if err != nil {
		t.Error("Should not be error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTrackValidatorReturning0Occurrences(t *testing.T) {
	err := client.Track("key", "trafficTypeNoOcurrences", "eventType", nil, nil)

	expected := "Track: traffic type traffictypenoocurrences does not have any corresponding Splits in this environment, make sure you’re tracking your events to a valid traffic type defined in the Split console"
	if err != nil {
		t.Error("Should not be error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTrackValidatorWitWrongTypeValue(t *testing.T) {
	err := client.Track("key", "traffic", "eventType", true, nil)

	expected := "Track: value must be a number"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTrackValidatorWithTooManyProperties(t *testing.T) {
	props := make(map[string]interface{})
	for i := 0; i < 301; i++ {
		props[fmt.Sprintf("prop-%d", i)] = "asd"
	}
	err := client.Track("key", "traffic", "eventType", 1, props)

	expected := "Track: Event has more than 300 properties. Some of them will be trimmed when processed"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func makeBigString(length int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	asRuneSlice := make([]rune, length)
	for index := range asRuneSlice {
		asRuneSlice[index] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(asRuneSlice)
}

func TestTrackValidatorEventTooBig(t *testing.T) {
	props := make(map[string]interface{})
	for i := 0; i < 299; i++ {
		props[fmt.Sprintf("%s%d", makeBigString(255), i)] = makeBigString(255)
	}
	err := client.Track("key", "traffic", "eventType", nil, props)

	expected := "The maximum size allowed for the properties is 32kb. Event not queued"
	if err == nil {
		t.Error("Should return an error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTrackValidator(t *testing.T) {
	err := client.Track("key", "traffic", "eventType", 1, nil)

	if err != nil {
		t.Error("Should not return error")
	}
}

func TestTrackClientDestroyed(t *testing.T) {

	var client2 = SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
		metrics:     mutexmap.NewMMMetricsStorage(),
		logger:      logger,
		validator:   inputValidation{logger: logger},
		sync: &sdkSync{
			countersSync:   nil,
			gaugeSync:      nil,
			impressionSync: nil,
			latenciesSync:  nil,
			segmentSync:    nil,
			splitSync:      nil,
		},
	}

	factory := SplitFactory{
		client: &client2,
	}

	client2.factory = &factory

	client2.Destroy()

	err := client2.Track("key", "trafficType", "eventType", 0, nil)

	expected := "Client has already been destroyed - no calls possible"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestManagerWithEmptySplit(t *testing.T) {
	splitStorage := mutexmap.NewMMSplitStorage()
	manager := SplitManager{
		splitStorage: splitStorage,
		logger:       logger,
	}
	factory := SplitFactory{
		manager: &manager,
	}

	factory.status.Store(SdkReady)
	manager.factory = &factory

	result := manager.Split("")

	expected := "Split: you passed an empty split name, split name must be a non-empty string"
	if result != nil {
		t.Error("Wrong result")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestManagerWithNonExistantSplit(t *testing.T) {
	splitStorage := mutexmap.NewMMSplitStorage()
	manager := SplitManager{
		splitStorage: splitStorage,
		logger:       logger,
		validator:    inputValidation{logger: logger},
	}
	factory := SplitFactory{
		manager: &manager,
	}

	factory.status.Store(SdkReady)
	manager.factory = &factory

	result := manager.Split("non_existant")

	expected := "Split: you passed non_existant that does not exist in this environment, please double check what Splits exist in the web console."
	if result != nil {
		t.Error("Wrong result")
	}

	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestDestroy(t *testing.T) {
	cfg.OperationMode = "redis-consumer"
	var client2 = SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
		metrics:     mutexmap.NewMMMetricsStorage(),
		logger:      logger,
		validator:   inputValidation{logger: logger},
		sync: &sdkSync{
			countersSync:   nil,
			gaugeSync:      nil,
			impressionSync: nil,
			latenciesSync:  nil,
			segmentSync:    nil,
			splitSync:      nil,
		},
	}

	var manager = SplitManager{
		logger:    logger,
		validator: inputValidation{logger: logger},
	}

	factory := SplitFactory{
		client:  &client2,
		manager: &manager,
	}

	client2.factory = &factory
	manager.factory = &factory

	client2.Destroy()

	result := client2.Treatment("key", "  feature   ", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Client has already been destroyed - no calls possible"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""

	resultManager := manager.Split("feature")

	if resultManager != nil {
		t.Error("Should be nil")
	}

	expected = "Client has already been destroyed - no calls possible"
	if !strings.Contains(strMsg, expected) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}
