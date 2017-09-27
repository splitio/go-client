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
}
