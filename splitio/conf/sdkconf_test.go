package conf

import (
	"testing"

	"github.com/splitio/go-split-commons/v9/conf"
	"github.com/splitio/go-split-commons/v9/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
)

// MockLogger implements logging.LoggerInterface for testing
type MockLogger struct {
	logging.LoggerInterface
}

func (m *MockLogger) Error(msg ...interface{}) {}


func TestSdkConfNormalization(t *testing.T) {
	cfg := Default()
	cfg.OperationMode = "invalid_mode"
	err := Normalize("asd", cfg)

	if err == nil {
		t.Error("Should throw an error when setting an invalid operation mode")
	}

	cfg = Default()
	err = Normalize("", cfg)
	if err == nil {
		t.Error("Should throw an error if no apikey is passed and operation mode != \"localhost\"")
	}

	cfg.SplitSyncProxyURL = "http://some-proxy"
	err = Normalize("asd", cfg)
	if err != nil {
		t.Error("Should not return an error with proper parameters")
	}

	if cfg.Advanced.SdkURL != cfg.SplitSyncProxyURL || cfg.Advanced.EventsURL != cfg.SplitSyncProxyURL {
		t.Error("Sdk & Events URL should be updated when SplitSyncProxyURL is not empty")
	}

	cfg = Default()
	cfg.IPAddressesEnabled = false
	err = Normalize("asd", cfg)
	if err != nil || cfg.IPAddress != "NA" || cfg.InstanceName != "NA" {
		t.Error("Should be NA")
	}

	cfg = Default()
	err = Normalize("asd", cfg)
	if err != nil || cfg.IPAddress == "NA" || cfg.InstanceName == "NA" {
		t.Error("Should not be NA")
	}
}

func TestSanitizeGlobalFallbackTreatment(t *testing.T) {
	logger := &MockLogger{}

	// Test nil input
	result := SanitizeGlobalFallbackTreatment(nil, logger)
	if result != nil {
		t.Error("Expected nil result for nil input")
	}

	// Test valid treatment
	validTreatment := "on"
	global := &dtos.FallbackTreatment{Treatment: &validTreatment}
	result = SanitizeGlobalFallbackTreatment(global, logger)
	if result == nil || *result.Treatment != validTreatment {
		t.Error("Expected valid treatment to be returned as is")
	}

	// Test invalid treatment (too long)
	longTreatment := "thisIsAVeryLongTreatmentThatExceedsTheMaxLengthOfOneHundredCharactersAndShouldDefinitelyBeTruncatedBecauseItIsTooLongToBeValid"
	global = &dtos.FallbackTreatment{Treatment: &longTreatment}
	result = SanitizeGlobalFallbackTreatment(global, logger)
	if result != nil {
		t.Error("Expected nil result for invalid treatment")
	}

	// Test invalid treatment (invalid characters)
	invalidTreatment := "invalid treatment"
	global = &dtos.FallbackTreatment{Treatment: &invalidTreatment}
	result = SanitizeGlobalFallbackTreatment(global, logger)
	if result != nil {
		t.Error("Expected nil result for treatment with invalid characters")
	}
}

func TestSanitizeByFlagFallBackTreatment(t *testing.T) {
	logger := &MockLogger{}

	// Test empty map
	result := SanitizeByFlagFallBackTreatment(nil, logger)
	if len(result) != 0 {
		t.Error("Expected empty map for nil input")
	}

	// Test valid entries
	validTreatment := "on"
	validFlag := "feature1"
	byFlag := map[string]dtos.FallbackTreatment{
		validFlag: {Treatment: &validTreatment},
	}
	result = SanitizeByFlagFallBackTreatment(byFlag, logger)
	if len(result) != 1 || *result[validFlag].Treatment != validTreatment {
		t.Error("Expected valid flag and treatment to be included")
	}

	// Test invalid flag name (contains space)
	invalidFlag := "invalid flag"
	byFlag = map[string]dtos.FallbackTreatment{
		invalidFlag: {Treatment: &validTreatment},
	}
	result = SanitizeByFlagFallBackTreatment(byFlag, logger)
	if len(result) != 0 {
		t.Error("Expected invalid flag name to be excluded")
	}

	// Test invalid treatment
	invalidTreatment := "invalid treatment"
	byFlag = map[string]dtos.FallbackTreatment{
		validFlag: {Treatment: &invalidTreatment},
	}
	result = SanitizeByFlagFallBackTreatment(byFlag, logger)
	if len(result) != 0 {
		t.Error("Expected invalid treatment to be excluded")
	}

	// Test multiple entries with mix of valid and invalid
	validTreatment2 := "off"
	validFlag2 := "feature2"
	byFlag = map[string]dtos.FallbackTreatment{
		validFlag:    {Treatment: &validTreatment},
		validFlag2:   {Treatment: &validTreatment2},
		invalidFlag: {Treatment: &validTreatment},
	}
	result = SanitizeByFlagFallBackTreatment(byFlag, logger)
	if len(result) != 2 || *result[validFlag].Treatment != validTreatment || *result[validFlag2].Treatment != validTreatment2 {
		t.Error("Expected only valid entries to be included")
	}
}

func TestValidRates(t *testing.T) {
	cfg := Default()
	err := Normalize("asd", cfg)
	if err != nil {
		t.Error("It should not return err")
	}

	cfg.TaskPeriods.TelemetrySync = 0
	err = Normalize("asd", cfg)
	if err == nil || err.Error() != "TelemetrySync must be >= 30. Actual is: 0" {
		t.Error("It should return err")
	}

	cfg = Default()
	cfg.TaskPeriods.SplitSync = 4
	err = Normalize("asd", cfg)
	if err == nil || err.Error() != "SplitSync must be >= 5. Actual is: 4" {
		t.Error("It should return err")
	}

	cfg = Default()
	cfg.TaskPeriods.SegmentSync = 29
	err = Normalize("asd", cfg)
	if err == nil || err.Error() != "SegmentSync must be >= 30. Actual is: 29" {
		t.Error("It should return err")
	}

	cfg = Default() // Optimized by Default
	cfg.TaskPeriods.ImpressionSync = 59
	err = Normalize("asd", cfg)
	if err == nil || err.Error() != "ImpressionSync must be >= 60. Actual is: 59" {
		t.Error("It should return err")
	}

	cfg = Default() // Optimized by Default
	cfg.TaskPeriods.ImpressionSync = 75
	err = Normalize("asd", cfg)
	if err != nil || cfg.TaskPeriods.ImpressionSync != 75 {
		t.Error("It should match")
	}

	cfg = Default() // Debug
	cfg.TaskPeriods.ImpressionSync = -1
	cfg.ImpressionsMode = conf.ImpressionsModeDebug
	err = Normalize("asd", cfg)
	if err == nil || err.Error() != "ImpressionSync must be >= 1. Actual is: -1" {
		t.Error("It should return err")
	}

	cfg = Default()
	cfg.TaskPeriods.EventsSync = 0
	err = Normalize("asd", cfg)
	if err == nil || err.Error() != "EventsSync must be >= 1. Actual is: 0" {
		t.Error("It should return err")
	}

	cfg = Default()
	cfg.Advanced.SegmentWorkers = 0
	err = Normalize("asd", cfg)
	if err == nil || err.Error() != "number of workers for fetching segments MUST be greater than zero" {
		t.Error("It should return err")
	}

	cfg = Default()
	cfg.ImpressionsMode = "some"
	err = Normalize("asd", cfg)
	if err != nil || cfg.ImpressionsMode != conf.ImpressionsModeOptimized {
		t.Error("It should not return err")
	}
}
