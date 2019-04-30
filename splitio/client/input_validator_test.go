package client

import (
	"math"
	"strings"
	"testing"

	"github.com/splitio/go-client/splitio/engine/evaluator"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-toolkit/logging"
)

var strMsg string

type MockWriter struct {
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	msg := string(p[:])
	strMsg = msg[strings.Index(msg, "[")+1 : strings.Index(msg, "]")]
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

var logger = logging.NewLogger(options)
var cfg = conf.Default()
var client = SplitClient{
	cfg:         cfg,
	evaluator:   &mockEvaluator{},
	impressions: mutexmap.NewMMImpressionStorage(),
	metrics:     mutexmap.NewMMMetricsStorage(),
	logger:      logger,
	validator:   inputValidation{logger: logger},
	events:      &mockEvents{},
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

	expected := "Factory instantiation: you passed and empty apikey, apikey must be a non-empty string"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTreatmentValidatorWithNilKey(t *testing.T) {
	result := client.Treatment(nil, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed a nil key, key must be a non-empty string"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTreatmentValidatorWithBooleanKey(t *testing.T) {
	result := client.Treatment(true, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an invalid key, key must be a non-empty string"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTreatmentValidatorWitWhitespaceKey(t *testing.T) {
	result := client.Treatment("     ", "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an empty key, key must be a non-empty string"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTreatmentValidatorWithStringKey(t *testing.T) {
	result := client.Treatment("key", "feature", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := ""
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
}

func TestTreatmentValidatorWithIntKey(t *testing.T) {
	result := client.Treatment(123, "feature", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := "Treatment: key %!s(int=123) is not of type string, converting"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
}

func TestTreatmentValidatorWithInt32Key(t *testing.T) {
	result := client.Treatment(int32(123), "feature", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := "Treatment: key %!s(int32=123) is not of type string, converting"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
}

func TestTreatmentValidatorWithInt64Key(t *testing.T) {
	result := client.Treatment(int64(123), "feature", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := "Treatment: key %!s(int64=123) is not of type string, converting"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
}

func TestTreatmentValidatorWitFloatKey(t *testing.T) {
	result := client.Treatment(1.3, "feature", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := "Treatment: key %!s(float64=1.3) is not of type string, converting"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
}

func TestTreatmentValidatorWitFloatNaNKey(t *testing.T) {
	result := client.Treatment(math.NaN, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an invalid key, key must be a non-empty string"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTreatmentValidatorWitFloatInfKey(t *testing.T) {
	result := client.Treatment(math.Inf, "feature", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an invalid key, key must be a non-empty string"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
}

func TestTreatmentValidatorEmptyFeatureName(t *testing.T) {
	result := client.Treatment("key", "", nil)

	if result != "control" {
		t.Error("Should be control")
	}

	expected := "Treatment: you passed an empty featureName, featureName must be a non-empty string"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTreatmentValidatorWhitespacesFeatureName(t *testing.T) {
	result := client.Treatment("key", "  feature   ", nil)

	if result != "TreatmentA" {
		t.Error("Should be TreatmentA")
	}

	expected := "Treatment: split name '  feature   ' has extra whitespace, trimming"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTreatmentClientDestroyed(t *testing.T) {

	var client2 = SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexmap.NewMMImpressionStorage(),
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
}

func TestTreatmentsWhitespaceFeatures(t *testing.T) {
	features := []string{" some_feature  "}

	result := client.Treatments("key", features, nil)

	if result["some_feature"] != "control" {
		t.Error("Wrong result")
	}

	expected := "Treatments: split name ' some_feature  ' has extra whitespace, trimming"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTreatmentsClientDestroyed(t *testing.T) {

	var client2 = SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexmap.NewMMImpressionStorage(),
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTrackValidatorWithEmptyKey(t *testing.T) {
	err := client.Track("", "trafficType", "eventType", nil)

	expected := "Track: you passed an empty key, key must be a non-empty string"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTrackValidatorWithLengthKey(t *testing.T) {
	m := ""
	for n := 0; n <= 256; n++ {
		m += "m"
	}

	err := client.Track(m, "trafficType", "eventType", nil)

	expected := "Track: key too long - must be 250 characters or less"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTrackValidatorWithEmptyEventName(t *testing.T) {
	err := client.Track("key", "trafficType", "", nil)

	expected := "Track: you passed an empty event type, event type must be a non-empty string"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTrackValidatorWithNotConformEventName(t *testing.T) {
	err := client.Track("key", "trafficType", "//", nil)

	expected := "Track: you passed //, event name must adhere to " +
		"the regular expression ^[a-zA-Z0-9][-_.:a-zA-Z0-9]{0,79}$. This means an event " +
		"name must be alphanumeric, cannot be more than 80 characters long, and can " +
		"only include a dash, underscore, period, or colon as separators of " +
		"alphanumeric characters"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if !strings.Contains(expected, strMsg) {
		t.Error("Error is distinct from the expected one")
	}
	strMsg = ""
}

func TestTrackValidatorWithEmptyTrafficType(t *testing.T) {
	err := client.Track("key", "", "eventType", nil)

	expected := "Track: you passed an empty traffic type, traffic type must be a non-empty string"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTrackValidatorWithUpperCaseTrafficType(t *testing.T) {
	err := client.Track("key", "traficTYPE", "eventType", nil)

	expected := "Track: traffic type should be all lowercase - converting string to lowercase"
	if err != nil {
		t.Error("Should not be error")
	}

	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTrackValidatorWitWrongTypeValue(t *testing.T) {
	err := client.Track("key", "traffic", "eventType", true)

	expected := "Track: value must be a number"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestTrackValidator(t *testing.T) {
	err := client.Track("key", "traffic", "eventType", 1)

	if err != nil {
		t.Error("Shoueld not return error")
	}
}

func TestTrackClientDestroyed(t *testing.T) {

	var client2 = SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexmap.NewMMImpressionStorage(),
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

	err := client2.Track("key", "trafficType", "eventType", 0)

	expected := "Client has already been destroyed - no calls possible"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
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

	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}

func TestDestroy(t *testing.T) {

	var client2 = SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexmap.NewMMImpressionStorage(),
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
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""

	resultManager := manager.Split("feature")

	if resultManager != nil {
		t.Error("Should be nil")
	}

	expected = "Client has already been destroyed - no calls possible"
	if strMsg != expected {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", strMsg)
		t.Error("Expected -> ", expected)
	}
	strMsg = ""
}
