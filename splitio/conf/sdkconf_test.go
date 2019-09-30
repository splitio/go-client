package conf

import (
	"testing"
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
