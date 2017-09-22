// Package client ...
// This package contains implementations of the Split SDK client and the factory used
// to instantiate it.
package client

import (
	//"github.com/splitio/go-client/splitio/service/api"
	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-client/splitio/util/logging"
)

// SplitFactory responsible for instantiating and storing instances of client and manager.
type SplitFactory struct {
	Client *SplitClient
}

func setupLogger(cfg *configuration.SplitSdkConfig) logging.LoggerInterface {
	var logger logging.LoggerInterface
	if cfg.Logger != nil {
		// If a custom logger is supplied, use it.
		logger = cfg.Logger
	} else {
		var loggerCfg *logging.LoggerOptions
		if cfg.LoggerConfig != nil {
			// If no custom logger is supplied but common & error writers are, use them
			loggerCfg = cfg.LoggerConfig
		} else {
			// No custom logger nor writers provided. Use logging package defaults
			loggerCfg = &logging.LoggerOptions{}
		}
		logger = logging.NewLogger(loggerCfg)
	}
	return logger
}

// NewFactory ...
// Factory constructor. Accepts a SplitSdkConfig struct as an argument, which will be used
// to instantiate both the client and the manager
func NewFactory(cfg *configuration.SplitSdkConfig) *SplitFactory {

	logger := setupLogger(cfg)
	//	splitFetcher := api.NewHTTPSplitFetcher(cfg, logger)
	//	segmentFetcher := api.NewHTTPSegmentFetcher(cfg, logger)

	client := &SplitClient{
		Apikey: cfg.Apikey,
		Logger: logger,
	}

	return &SplitFactory{
		Client: client,
	}
}
