// Package client contains implementations of the Split SDK client and the factory used
// to instantiate it.
package client

import (
	"fmt"
	"time"

	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/service/api"
	"github.com/splitio/go-client/splitio/service/local"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-client/splitio/storage/mutexqueue"
	"github.com/splitio/go-client/splitio/storage/redisdb"
	"github.com/splitio/go-client/splitio/tasks"

	"github.com/splitio/go-toolkit/logging"
)

var inMemoryFullQueueChan = make(chan bool, 1)

// SplitFactory struct is responsible for instantiating and storing instances of client and manager.
type SplitFactory struct {
	client  *SplitClient
	manager *SplitManager
}

// Client returns the split client instantiated by the factory
func (f *SplitFactory) Client() *SplitClient {
	return f.client
}

// Manager returns the split manager instantiated by the factory
func (f *SplitFactory) Manager() *SplitManager {
	return f.manager
}

// setupLogger sets up the logger according to the parameters submitted by the sdk user
func setupLogger(cfg *conf.SplitSdkConfig) logging.LoggerInterface {
	var logger logging.LoggerInterface
	if cfg.Logger != nil {
		// If a custom logger is supplied, use it.
		logger = cfg.Logger
	} else {
		logger = logging.NewLogger(&cfg.LoggerConfig)
	}
	return logger
}

// NewSplitFactory instntiates a new SplitFactory object. Accepts a SplitSdkConfig struct as an argument,
// which will be used to instantiate both the client and the manager
func NewSplitFactory(apikey string, cfg *conf.SplitSdkConfig) (*SplitFactory, error) {
	if cfg == nil {
		cfg = conf.Default()
	}

	logger := setupLogger(cfg)

	err := conf.Normalize(apikey, cfg)
	if err != nil {
		logger.Error("Error occurred when processing configuration")
		return nil, err
	}

	// Set up storages
	var splitStorage storage.SplitStorage
	var segmentStorage storage.SegmentStorage
	var impressionStorage storage.ImpressionStorage
	var metricsStorage storage.MetricsStorage
	var eventsStorage storage.EventsStorage

	switch cfg.OperationMode {
	case "inmemory-standalone", "localhost":
		splitStorage = mutexmap.NewMMSplitStorage()
		segmentStorage = mutexmap.NewMMSegmentStorage()
		impressionStorage = mutexmap.NewMMImpressionStorage()
		metricsStorage = mutexmap.NewMMMetricsStorage()
		eventsStorage = mutexqueue.NewMQEventsStorage(cfg.Advanced.EventsQueueSize, inMemoryFullQueueChan)
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
			cfg.IPAddress,
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
			cfg.IPAddress,
			splitio.Version,
			logger,
		)
		eventsStorage = redisdb.NewRedisEventsStorage(
			host,
			port,
			db,
			password,
			prefix,
			cfg.IPAddress,
			cfg.InstanceName,
			fmt.Sprintf("go-%s", splitio.Version),
			logger,
		)
	default:
		return nil, fmt.Errorf("Invalid operation mode \"%s\"", cfg.OperationMode)
	}

	version := splitio.Version
	ip := cfg.IPAddress
	instance := cfg.InstanceName
	// Setup synchronization structs and tasks
	var syncTasks *sdkSync
	switch cfg.OperationMode {
	case "localhost":
		splitFetcher := local.NewFileSplitFetcher(cfg.SplitFile, local.SplitFileFormatClassic)
		splitPeriod := cfg.TaskPeriods.SplitSync
		readyChannel := make(chan string)
		syncTasks = &sdkSync{
			splitSync: tasks.NewFetchSplitsTask(splitStorage, splitFetcher, splitPeriod, logger, readyChannel),
		}
		syncTasks.splitSync.Start()
		select {
		case <-readyChannel:
			break // Only SPLITS_READY should be sent, no need to check
		case <-time.After(time.Second * time.Duration(cfg.BlockUntilReady)):
			return nil, fmt.Errorf("SDK Initialization time of %d exceeded", cfg.BlockUntilReady)
		}
	case "inmemory-standalone", "redis-standalone":
		// Sync structs
		splitFetcher := api.NewHTTPSplitFetcher(apikey, cfg, logger)
		segmentFetcher := api.NewHTTPSegmentFetcher(apikey, cfg, logger)
		impressionRecorder := api.NewHTTPImpressionRecorder(apikey, cfg, logger)
		metricsRecorder := api.NewHTTPMetricsRecorder(apikey, cfg, logger)
		eventsRecorder := api.NewHTTPEventsRecorder(apikey, cfg, logger)

		// Task periods
		splitPeriod := cfg.TaskPeriods.SplitSync
		segmentPeriod := cfg.TaskPeriods.SegmentSync
		impressionPeriod := cfg.TaskPeriods.ImpressionSync
		eventsPeriod := cfg.TaskPeriods.EventsSync
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
				impressionStorage,
				impressionRecorder,
				impressionPeriod,
				version,
				ip,
				instance,
				cfg.Advanced.ImpressionListener,
				logger,
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
			eventsSync: tasks.NewRecordEventsTask(
				eventsStorage,
				eventsRecorder,
				cfg.Advanced.EventsBulkSize,
				eventsPeriod,
				version,
				ip,
				instance,
				logger,
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
			case "SPLITS_ERROR":
				return nil, fmt.Errorf("Split syncrhonization failed. Please check your apikey")
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
				syncTasks.eventsSync.Start()
				break
			}
		case <-time.After(time.Duration(remaining) * time.Second):
			return nil, fmt.Errorf("SDK Initialization time of %d exceeded", cfg.BlockUntilReady)
		}

		// Flushing storage queue signal only for inmemory-standalone
		if cfg.OperationMode == "inmemory-standalone" {
			go func() {
				for true {
					isFull := <-inMemoryFullQueueChan
					if isFull {
						logger.Debug("FLUSHING storage queue")
						errWakeUp := syncTasks.eventsSync.WakeUp()
						if errWakeUp != nil {
							logger.Error("Error flushing storage queue", errWakeUp)
						}
					}
				}
			}()
		}

	case "redis-consumer":
		// No synchronization tasks necessary in redis-consumer mode
	default:
		return nil, fmt.Errorf("Invalid operation mode \"%s\"", cfg.OperationMode)
	}

	logger.Info("Sdk initialization complete!")

	engine := engine.NewEngine(logger)
	client := &SplitClient{
		apikey:      apikey,
		logger:      logger,
		evaluator:   evaluator.NewEvaluator(splitStorage, segmentStorage, engine, logger),
		impressions: impressionStorage,
		metrics:     metricsStorage,
		sync:        syncTasks,
		cfg:         cfg,
		events:      eventsStorage,
		validator:   inputValidation{logger: logger},
	}

	manager := &SplitManager{
		splitStorage: splitStorage,
		validator:    inputValidation{logger: logger},
		logger:       logger,
	}

	return &SplitFactory{
		client:  client,
		manager: manager,
	}, nil
}
