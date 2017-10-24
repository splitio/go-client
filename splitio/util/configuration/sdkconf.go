// Package configuration ...
// Contains configuration structures used to setup the SDK
package configuration

import (
	"github.com/splitio/go-client/splitio/util/logging"
)

// SplitSdkConfig struct ...
// struct used to setup a Split.io SDK client.
//
// Parameters:
// - Apikey: API-KEY used to authenticate user requests
type SplitSdkConfig struct {
	Apikey       string
	Logger       logging.LoggerInterface
	LoggerConfig *logging.LoggerOptions
	HTTPTimeout  int
	Advanced     *AdvancedConfig
}

// AdvancedConfig exposes more configurable parameters that can be used to further tailor the sdk to the user's needs
type AdvancedConfig struct {
	SdkURL           string
	EventsURL        string
	segmentQueueSize int
	segmentWorkers   int
}
