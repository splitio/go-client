// Package client contains implementations of the Split SDK client and the factory used
// to instantiate it.
package client

import (
	"fmt"
	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/engine"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/service/api"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-client/splitio/storage/redisdb"
	"github.com/splitio/go-client/splitio/tasks"
	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-toolkit/logging"
	"time"
)

// SplitFactory struct is responsible for instantiating and storing instances of client and manager.
type SplitFactory struct {
	Client  *SplitClient
	Manager *SplitManager
}

// setupLogger sets up the logger according to the parameters submitted by the sdk user
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

// NewSplitFactory instntiates a new SplitFactory object. Accepts a SplitSdkConfig struct as an argument,
// which will be used to instantiate both the client and the manager
func NewSplitFactory(cfg *configuration.SplitSdkConfig) (*SplitFactory, error) {
	logger := setupLogger(cfg)

	err := cfg.Normalize()

	if err != nil {
		logger.Error("Error occurred when processing configuration")
		return nil, err
	}

	// Set up storages
	var splitStorage storage.SplitStorage
	var segmentStorage storage.SegmentStorage
	var impressionStorage storage.ImpressionStorage
	var metricsStorage storage.MetricsStorage
	switch cfg.OperationMode {
	case "inmemory-standalone":
		splitStorage = mutexmap.NewMMSplitStorage()
		segmentStorage = mutexmap.NewMMSegmentStorage()
		impressionStorage = mutexmap.NewMMImpressionStorage()
		metricsStorage = mutexmap.NewMMMetricsStorage()
	case "redis-consumer", "redis-standalone":
		host := cfg.Redis.Host
		port := cfg.Redis.Port
		db := cfg.Redis.Database
		password := cfg.Redis.Password
		prefix := cfg.Redis.Prefix
		splitStorage = redisdb.NewRedisSplitStorage(host, port, db, password, prefix, logger)
		segmentStorage = redisdb.NewRedisSegmentStorage(host, port, db, password, prefix, logger)
		impressionStorage = redisdb.NewRedisImpressionStorage(
			host,
			port,
			db,
			password,
			prefix,
			cfg.InstanceName,
			fmt.Sprintf("go-%s", splitio.Version),
			logger,
		)
		metricsStorage = redisdb.NewRedisMetricsStorage(
			host,
			port,
			db,
			password,
			prefix,
			cfg.InstanceName,
			splitio.Version,
			logger,
		)
	default:
		return nil, fmt.Errorf("Invalid operation mode \"%s\"", cfg.OperationMode)
	}

	version := splitio.Version
	ip := cfg.IpAddress
	instance := cfg.InstanceName
	// Setup synchronization structs and tasks
	var syncTasks *sdkSync
	switch cfg.OperationMode {
	case "inmemory-standalone", "redis-standalone":
		// Sync structs
		splitFetcher := api.NewHTTPSplitFetcher(cfg, logger)
		segmentFetcher := api.NewHTTPSegmentFetcher(cfg, logger)
		impressionRecorder := api.NewHTTPImpressionRecorder(cfg, logger)
		metricsRecorder := api.NewHTTPMetricsRecorder(cfg, logger)

		// Task periods
		splitPeriod := cfg.TaskPeriods.SplitSync
		segmentPeriod := cfg.TaskPeriods.SegmentSync
		impressionPeriod := cfg.TaskPeriods.ImpressionSync
		countersPeriod := cfg.TaskPeriods.CounterSync
		gaugePeriod := cfg.TaskPeriods.GaugeSync
		latencyPeriod := cfg.TaskPeriods.LatencySync
		workers := cfg.Advanced.SegmentWorkers
		qSize := cfg.Advanced.SegmentQueueSize

		readyChannel := make(chan string)
		// Sync tasks
		syncTasks = &sdkSync{
			splitSync: tasks.NewFetchSplitsTask(splitStorage, splitFetcher, splitPeriod, logger, readyChannel),
			segmentSync: tasks.NewFetchSegmentsTask(
				splitStorage,
				segmentStorage,
				segmentFetcher,
				segmentPeriod,
				workers,
				qSize,
				logger,
				readyChannel,
			),
			impressionSync: tasks.NewRecordImpressionsTask(
				impressionStorage, impressionRecorder, impressionPeriod, version, ip, instance, logger,
			),
			countersSync: tasks.NewRecordCountersTask(
				metricsStorage, metricsRecorder, countersPeriod, version, ip, instance, logger,
			),
			gaugeSync: tasks.NewRecordGaugesTask(
				metricsStorage, metricsRecorder, gaugePeriod, version, ip, instance, logger,
			),
			latenciesSync: tasks.NewRecordLatenciesTask(
				metricsStorage, metricsRecorder, latencyPeriod, version, ip, instance, logger,
			),
		}

		// Start split fetching task
		syncTasks.splitSync.Start()

		// Block until ready part 1: splits
		preSplitsTS := time.Now()
		select {
		case msg := <-readyChannel:
			switch msg {
			case "SPLITS_READY":
				// Once splits are ready, start segment fetching task
				syncTasks.segmentSync.Start()
				break
			}
		case <-time.After(time.Second * time.Duration(cfg.BlockUntilReady)):
			return nil, fmt.Errorf("SDK Initialization time of %d exceeded", cfg.BlockUntilReady)
		}

		// Block until ready part 2: segments
		remaining := cfg.BlockUntilReady - int(time.Now().Sub(preSplitsTS).Seconds())
		select {
		case msg := <-readyChannel:
			switch msg {
			case "SEGMENTS_READY":
				// Once segments are ready, start impressions and metrics recording tasks
				syncTasks.impressionSync.Start()
				syncTasks.latenciesSync.Start()
				syncTasks.countersSync.Start()
				syncTasks.gaugeSync.Start()
				break
			}
		case <-time.After(time.Duration(remaining) * time.Second):
			return nil, fmt.Errorf("SDK Initialization time of %d exceeded", cfg.BlockUntilReady)
		}

	case "redis-consumer":
		// No synchronization tasks necessary in redis-consumer mode
	default:
		return nil, fmt.Errorf("Invalid operation mode \"%s\"", cfg.OperationMode)
	}

	engine := engine.NewEngine(logger)
	client := &SplitClient{
		apikey:      cfg.Apikey,
		logger:      logger,
		evaluator:   evaluator.NewEvaluator(splitStorage, segmentStorage, engine, logger),
		impressions: impressionStorage,
		metrics:     metricsStorage,
		sync:        syncTasks,
	}

	manager := &SplitManager{splitStorage: splitStorage}

	return &SplitFactory{
		Client:  client,
		Manager: manager,
	}, nil
}
