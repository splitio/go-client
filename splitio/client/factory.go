// Package client contains implementations of the Split SDK client and the factory used
// to instantiate it.
package client

import (
	"errors"
	"fmt"
	"sync/atomic"
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

// SdkError flag
const SdkError = 0

// SdkReady flag
const SdkReady = 1

var inMemoryFullQueueChan = make(chan bool, 1)

// SplitFactory struct is responsible for instantiating and storing instances of client and manager.
type SplitFactory struct {
	client                *SplitClient
	manager               *SplitManager
	destroyed             atomic.Value
	readinessSubscriptors map[int]chan int
	ready                 atomic.Value
	operationMode         string
}

// Client returns the split client instantiated by the factory
func (f *SplitFactory) Client() *SplitClient {
	f.client.factory = f
	f.manager.factory = f
	return f.client
}

// Manager returns the split manager instantiated by the factory
func (f *SplitFactory) Manager() *SplitManager {
	f.client.factory = f
	f.manager.factory = f
	return f.manager
}

// Destroy blocks all the operations
func (f *SplitFactory) Destroy() {
	f.destroyed.Store(true)
}

// IsDestroyed returns true if tbe client has been destroyed
func (f *SplitFactory) IsDestroyed() bool {
	return f.destroyed.Load().(bool)
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

// initializates task for localhost mode
func (f *SplitFactory) initializationLocalhost(readyChannel chan string, syncTasks *sdkSync) {
	syncTasks.splitSync.Start()

	<-readyChannel
	f.broadcastReadiness(SdkReady)
}

// initializates tasks for in-memory mode
func (f *SplitFactory) initializationInMemory(readyChannel chan string, syncTasks *sdkSync) {
	// Start split fetching task
	syncTasks.splitSync.Start()

	msg := <-readyChannel
	switch msg {
	case "SPLITS_READY":
		// Once splits are ready, start segment fetching task
		syncTasks.segmentSync.Start()
		break
	case "SPLITS_ERROR":
		// Broadcast on error
		f.broadcastReadiness(SdkError)
		return
	}

	msg = <-readyChannel
	switch msg {
	case "SEGMENTS_READY":
		// Once segments are ready, start impressions and metrics recording tasks
		syncTasks.impressionSync.Start()
		syncTasks.latenciesSync.Start()
		syncTasks.countersSync.Start()
		syncTasks.gaugeSync.Start()
		syncTasks.eventsSync.Start()
		// Broadcast ready status for SDK
		f.broadcastReadiness(SdkReady)
	}
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
		logger.Error(err.Error())
		return nil, err
	}

	// Set up storages
	var splitStorage storage.SplitStorage
	var segmentStorage storage.SegmentStorage
	var impressionStorage storage.ImpressionStorage
	var metricsStorage storage.MetricsStorage
	var eventsStorage storage.EventsStorage

	if cfg.OperationMode == "inmemory-standalone" {
		err = api.ValidateApikey(apikey, cfg.Advanced)
		if err != nil {
			return nil, err
		}
	}

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

	// Create new Split Factory
	sp := &SplitFactory{
		client:                client,
		manager:               manager,
		readinessSubscriptors: make(map[int]chan int),
		operationMode:         cfg.OperationMode,
	}
	sp.ready.Store(false)

	switch cfg.OperationMode {
	case "localhost":
		splitFetcher := local.NewFileSplitFetcher(cfg.SplitFile, local.SplitFileFormatClassic)
		splitPeriod := cfg.TaskPeriods.SplitSync
		readyChannel := make(chan string)
		syncTasks = &sdkSync{
			splitSync: tasks.NewFetchSplitsTask(splitStorage, splitFetcher, splitPeriod, logger, readyChannel),
		}
		// Call fetching tasks as goroutine
		go sp.initializationLocalhost(readyChannel, syncTasks)
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
		// Call fetching tasks as goroutine
		go sp.initializationInMemory(readyChannel, syncTasks)

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
		sp.ready.Store(true)
	default:
		return nil, fmt.Errorf("Invalid operation mode \"%s\"", cfg.OperationMode)
	}

	logger.Info("Sdk initialization complete!")

	sp.destroyed.Store(false)

	return sp, nil
}

// broadcastReadiness broadcasts message to all the subscriptors
func (f *SplitFactory) broadcastReadiness(status int) {
	if status == SdkReady {
		f.ready.Store(true)
	}
	for _, subscriptor := range f.readinessSubscriptors {
		subscriptor <- status
	}
}

// subscribes listener
func (f *SplitFactory) subscribe(name int, subscriptor chan int) {
	f.readinessSubscriptors[name] = subscriptor
}

// removes a particular subscriptor from the list
func (f *SplitFactory) unsubscribe(name int, subscriptor chan int) {
	_, ok := f.readinessSubscriptors[name]
	if ok {
		delete(f.readinessSubscriptors, name)
	}
}

// IsReady returns true if the factory is ready
func (f *SplitFactory) IsReady() bool {
	return f.ready.Load().(bool)
}

// BlockUntilReady blocks client or manager until the SDK is ready, error occurs or times out
func (f *SplitFactory) BlockUntilReady(timer int) error {
	if f.IsReady() {
		return nil
	}
	if timer <= 0 {
		return errors.New("SDK Initialization: timer must be positive number")
	}
	block := make(chan int, 1)

	subscriptorName := len(f.readinessSubscriptors)

	defer func() {
		// Unsubscription will happen only if a block channel has been created
		if block != nil {
			f.unsubscribe(subscriptorName, block)
			close(block)
		}
	}()

	f.subscribe(subscriptorName, block)

	select {
	case status := <-block:
		switch status {
		case SdkReady:
			break
		case SdkError:
			return errors.New("SDK Initialization failed")
		}
	case <-time.After(time.Second * time.Duration(timer)):
		return fmt.Errorf("SDK Initialization: time of %d exceeded", timer)
	}

	return nil
}
