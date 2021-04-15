package conf

import (
	"testing"

	"github.com/splitio/go-split-commons/v3/conf"
)

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
	if err == nil || err.Error() != "Number of workers for fetching segments MUST be greater than zero" {
		t.Error("It should return err")
	}

	cfg = Default()
	cfg.ImpressionsMode = "some"
	err = Normalize("asd", cfg)
	if err != nil || cfg.ImpressionsMode != conf.ImpressionsModeOptimized {
		t.Error("It should not return err")
	}
}
