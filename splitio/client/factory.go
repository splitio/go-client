// Package client contains implementations of the Split SDK client and the factory used
// to instantiate it.
package client

import (
	"github.com/splitio/go-client/splitio/engine"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/service/api"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/tasks"
	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-client/splitio/util/logging"
)

// SplitFactory struct is responsible for instantiating and storing instances of client and manager.
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

// NewSplitFactory instntiates a new SplitFactory object. Accepts a SplitSdkConfig struct as an argument, which will be used
// to instantiate both the client and the manager
func NewSplitFactory(cfg *configuration.SplitSdkConfig) *SplitFactory {

	logger := setupLogger(cfg)

	// Setup fetchers
	splitFetcher := api.NewHTTPSplitFetcher(cfg, logger)
	segmentFetcher := api.NewHTTPSegmentFetcher(cfg, logger)
	impressionRecorder := api.NewHTTPImpressionRecorder(cfg, logger)
	metricsRecorder := api.NewHTTPMetricsRecorder(cfg, logger)

	// Setup Storage
	splitStorage := storage.NewMMSplitStorage()
	segmentStorage := storage.NewMMSegmentStorage()
	impressionStorage := storage.NewMMImpressionStorage()
	metricsStorage := storage.NewMMMetricsStorage()

	// Setup tasks
	// TODO: PARAMETRIZE!
	splitSyncTask := tasks.NewFetchSplitsTask(splitStorage, splitFetcher, 30, logger)
	segmentSyncTask := tasks.NewFetchSegmentsTask(splitStorage, segmentStorage, segmentFetcher, 30, 10, 1000, logger)
	impressionSyncTask := tasks.NewRecordImpressionsTask(impressionStorage, impressionRecorder, 30, "go-0.1", "1.2.3.4", "m1", logger)
	countersSyncTask := tasks.NewRecordCountersTask(metricsStorage, metricsRecorder, 30, "go-0.1", "1.2.3.4", "m1", logger)
	gaugeSyncTask := tasks.NewRecordGaugesTask(metricsStorage, metricsRecorder, 30, "go-0.1", "1.2.3.4", "m1", logger)
	latenciesSyncTask := tasks.NewRecordLatenciesTask(metricsStorage, metricsRecorder, 30, "go-0.1", "1.2.3.4", "m1", logger)

	client := &SplitClient{
		Apikey:    cfg.Apikey,
		Logger:    logger,
		Evaluator: evaluator.NewEvaluator(splitStorage, segmentStorage, engine.Engine{Logger: logger}),
		sync: sdkSync{
			splitSync:      splitSyncTask,
			segmentSync:    segmentSyncTask,
			impressionSync: impressionSyncTask,
			countersSync:   countersSyncTask,
			gaugeSync:      gaugeSyncTask,
			latenciesSync:  latenciesSyncTask,
		},
	}

	return &SplitFactory{
		Client: client,
	}
}
