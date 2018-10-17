package client

import (
	"testing"

	"github.com/splitio/go-toolkit/logging"
)

func TestTreatmentValidatorWithNilKey(t *testing.T) {

	logger := logging.NewLogger(nil)
	validator := inputValidation{logger: logger}
	_, _, err := validator.ValidateTreatmentKey(nil)

	if err == nil {
		t.Error("Should be invalid key")
	}

	if err.Error() != "Treatment: key cannot be nil" {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", err.Error())
		t.Error("Expected -> ", "Treatment: key cannot be nil")
	}
}

func TestTreatmentValidatorWithBooleanKey(t *testing.T) {

	logger := logging.NewLogger(nil)
	validator := inputValidation{logger: logger}
	_, _, err := validator.ValidateTreatmentKey(true)

	if err == nil {
		t.Error("Should be invalid key")
	}

	if err.Error() != "Treatment: supplied key is neither a string or a Key struct" {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", err.Error())
		t.Error("Expected -> ", "Treatment: supplied key is neither a string or a Key struct")
	}
}

func TestTreatmentValidatorWithStringKey(t *testing.T) {

	logger := logging.NewLogger(nil)
	validator := inputValidation{logger: logger}
	matchingKey, _, err := validator.ValidateTreatmentKey("test")

	if err != nil {
		t.Error("Should be valid key")
	}

	if matchingKey != "test" {
		t.Error("matchingKey should be == test")
	}
}

func TestTreatmentValidatorWithIntKey(t *testing.T) {

	logger := logging.NewLogger(nil)
	validator := inputValidation{logger: logger}
	matchingKey, _, err := validator.ValidateTreatmentKey(123)

	if err != nil {
		t.Error("Should be valid key")
	}

	if matchingKey != "123" {
		t.Error("matchingKey should be == 123")
	}
}

func TestTreatmentValidatorWithMatchingKey(t *testing.T) {

	var key = &Key{
		MatchingKey:  "test",
		BucketingKey: "",
	}

	logger := logging.NewLogger(nil)
	validator := inputValidation{logger: logger}
	matchingKey, _, err := validator.ValidateTreatmentKey(key)

	if err != nil {
		t.Error("Should be valid key")
	}

	if matchingKey != "test" {
		t.Error("matchingKey should be == test")
	}
}

func TestTreatmentValidatorWithKeyObject(t *testing.T) {

	var key = &Key{
		MatchingKey:  "test",
		BucketingKey: "test-bucketing",
	}

	logger := logging.NewLogger(nil)
	validator := inputValidation{logger: logger}
	matchingKey, bucketingKey, err := validator.ValidateTreatmentKey(key)

	if err != nil {
		t.Error("Should be valid key")
	}

	if matchingKey != "test" {
		t.Error("matchingKey should be == test")
	}

	if *bucketingKey != "test-bucketing" {
		t.Error("bucketingKey should be == test-bucketing")
	}
}

func TestTrackValidatorWithEmptyTrafficType(t *testing.T) {

	key, _, eventType, value, err := ValidateTrackInputs("key", "", "eventType", 123)

	if err == nil {
		t.Error("Should be errors")
	}

	if err.Error() != "Track: trafficType must not be an empty String" {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", err.Error())
		t.Error("Expected -> ", "Track: trafficType must not be an empty String")
	}

	if key != "key" || eventType != "eventType" || value != 123 {
		t.Error("Inputs should not change")
	}
}

func TestTrackValidatorWithWrongEventType(t *testing.T) {

	key, trafficType, _, value, err := ValidateTrackInputs("key", "trafficType", "@@", 123)

	if err == nil {
		t.Error("Should be errors")
	}

	if err.Error() != "Track: eventName must adhere to the regular expression [a-zA-Z0-9][-_\\.a-zA-Z0-9]{0,62}" {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", err.Error())
		t.Error("Expected -> ", "Track: eventName must adhere to the regular expression [a-zA-Z0-9][-_\\.a-zA-Z0-9]{0,62}")
	}

	if key != "key" || trafficType != "trafficType" || value != 123 {
		t.Error("Inputs should not change")
	}
}

func TestTrackValidatorWithNilValue(t *testing.T) {

	key, trafficType, eventType, value, err := ValidateTrackInputs("key", "trafficType", "eventType", nil)

	if err != nil {
		t.Error("Should be valid")
	}

	if key != "key" || trafficType != "trafficType" || eventType != "eventType" || value != nil {
		t.Error("Inputs should not change")
	}
}

func TestTrackValidatorWithStringValue(t *testing.T) {

	key, trafficType, eventType, _, err := ValidateTrackInputs("key", "trafficType", "eventType", "invalid")

	if err == nil {
		t.Error("Should be errors")
	}

	if err.Error() != "Track: value must be a number" {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", err.Error())
		t.Error("Expected -> ", "TTrack: value must be a number")
	}

	if key != "key" || trafficType != "trafficType" || eventType != "eventType" {
		t.Error("Inputs should not change")
	}
}

func TestTrackValidatorWithBoolValue(t *testing.T) {

	key, trafficType, eventType, _, err := ValidateTrackInputs("key", "trafficType", "eventType", true)

	if err == nil {
		t.Error("Should be errors")
	}

	if err.Error() != "Track: value must be a number" {
		t.Error("Error is distinct from the expected one")
		t.Error("Actual -> ", err.Error())
		t.Error("Expected -> ", "TTrack: value must be a number")
	}

	if key != "key" || trafficType != "trafficType" || eventType != "eventType" {
		t.Error("Inputs should not change")
	}
}

func TestTrackValidatorWithValidInputs(t *testing.T) {

	key, trafficType, eventType, value, err := ValidateTrackInputs("key", "trafficType", "eventType", 123)

	if err != nil {
		t.Error("Should be valid")
	}

	if key != "key" || trafficType != "trafficType" || eventType != "eventType" || value != 123 {
		t.Error("Inputs should not change")
	}
}

func TestTrackValidatorWithValidInputs2(t *testing.T) {

	key, trafficType, eventType, value, err := ValidateTrackInputs("key", "trafficType", "eventType", 1.6)

	if err != nil {
		t.Error("Should be valid")
	}

	if key != "key" || trafficType != "trafficType" || eventType != "eventType" || value != 1.6 {
		t.Error("Inputs should not change")
	}
}
