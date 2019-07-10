// Package client contains implementations of the Split SDK client and the factory used
// to instantiate it.
package client

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/impressionListener"
	"github.com/splitio/go-client/splitio/service/api"
	"github.com/splitio/go-client/splitio/service/local"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-client/splitio/storage/mutexqueue"
	"github.com/splitio/go-client/splitio/storage/redisdb"
	"github.com/splitio/go-client/splitio/tasks"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"
)

const (
	sdkStatusDestroyed = iota
	sdkStatusInitializing
	sdkStatusReady

	sdkInitializationFailed = -1
)

type sdkStorages struct {
	splits      storage.SplitStorageConsumer
	segments    storage.SegmentStorageConsumer
	impressions storage.ImpressionStorageProducer
	events      storage.EventStorageProducer
	telemetry   storage.MetricsStorageProducer
}

type sdkSync struct {
	splits      *asynctask.AsyncTask
	segments    *asynctask.AsyncTask
	impressions *asynctask.AsyncTask
	gauges      *asynctask.AsyncTask
	counters    *asynctask.AsyncTask
	latencies   *asynctask.AsyncTask
	events      *asynctask.AsyncTask
}

// SplitFactory struct is responsible for instantiating and storing instances of client and manager.
type SplitFactory struct {
	metadata              splitio.SdkMetadata
	tasks                 sdkSync
	storages              sdkStorages
	apikey                string
	status                atomic.Value
	readinessSubscriptors map[int]chan int
	operationMode         string
	mutex                 sync.Mutex
	cfg                   *conf.SplitSdkConfig
	impressionListener    *impressionlistener.WrapperImpressionListener
	logger                logging.LoggerInterface
}

// Client returns the split client instantiated by the factory
func (f *SplitFactory) Client() *SplitClient {
	return &SplitClient{
		logger:      f.logger,
		evaluator:   evaluator.NewEvaluator(f.storages.splits, f.storages.segments, engine.NewEngine(logger), f.logger),
		impressions: f.storages.impressions,
		metrics:     f.storages.telemetry,
		events:      f.storages.events,
		validator: inputValidation{
			logger:       logger,
			splitStorage: f.storages.splits,
		},
		factory:            f,
		impressionListener: f.impressionListener,
	}
}

// Manager returns the split manager instantiated by the factory
func (f *SplitFactory) Manager() *SplitManager {
	return &SplitManager{
		splitStorage: f.storages.splits,
		validator:    inputValidation{logger: logger},
		logger:       logger,
		factory:      f,
	}
}

// IsDestroyed returns true if tbe client has been destroyed
func (f *SplitFactory) IsDestroyed() bool {
	return f.status.Load() == sdkStatusDestroyed
}

// IsReady returns true if the factory is ready
func (f *SplitFactory) IsReady() bool {
	return f.status.Load() == sdkStatusReady
}

// initializates task for localhost mode
func (f *SplitFactory) initializationLocalhost(readyChannel chan string, syncTasks *sdkSync) {
	syncTasks.splits.Start()

	<-readyChannel
	f.broadcastReadiness(sdkStatusReady)
}

// initializates tasks for in-memory mode
func (f *SplitFactory) initializationInMemory(readyChannel chan string, syncTasks *sdkSync) {
	// Start split fetching task
	syncTasks.splits.Start()

	msg := <-readyChannel
	switch msg {
	case "SPLITS_READY":
		// Once splits are ready, start segment fetching task
		syncTasks.segments.Start()
		break
	case "SPLITS_ERROR":
		// Broadcast on error
		f.broadcastReadiness(sdkInitializationFailed)
		return
	}

	msg = <-readyChannel
	switch msg {
	case "SEGMENTS_READY":
		// Once segments are ready, start impressions and metrics recording tasks
		syncTasks.impressions.Start()
		syncTasks.latencies.Start()
		syncTasks.counters.Start()
		syncTasks.gauges.Start()
		syncTasks.events.Start()
		// Broadcast ready status for SDK
		f.broadcastReadiness(sdkStatusReady)
	}
}

func dataFlusher(syncTasks *sdkSync, inMememoryFullQueue chan string, logger logging.LoggerInterface) {
	for true {
		msg := <-inMememoryFullQueue
		switch msg {
		case "EVENTS_FULL":
			logger.Debug("FLUSHING storage queue")
			errWakeUp := syncTasks.events.WakeUp()
			if errWakeUp != nil {
				logger.Error("Error flushing storage queue", errWakeUp)
			}
			break
		case "IMPRESSIONS_FULL":
			logger.Debug("FLUSHING storage queue")
			errWakeUp := syncTasks.impressions.WakeUp()
			if errWakeUp != nil {
				logger.Error("Error flushing storage queue", errWakeUp)
			}
		}
	}
}

// broadcastReadiness broadcasts message to all the subscriptors
func (f *SplitFactory) broadcastReadiness(status int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.status.Load() == sdkStatusInitializing && status == sdkStatusReady {
		f.status.Store(sdkStatusReady)
	}
	for _, subscriptor := range f.readinessSubscriptors {
		subscriptor <- status
	}
}

// subscribes listener
func (f *SplitFactory) subscribe(name int, subscriptor chan int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.readinessSubscriptors[name] = subscriptor
}

// removes a particular subscriptor from the list
func (f *SplitFactory) unsubscribe(name int, subscriptor chan int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	_, ok := f.readinessSubscriptors[name]
	if ok {
		delete(f.readinessSubscriptors, name)
	}
}

// BlockUntilReady blocks client or manager until the SDK is ready, error occurs or times out
func (f *SplitFactory) BlockUntilReady(timer int) error {
	if f.IsReady() {
		return nil
	}
	if timer <= 0 {
		return errors.New("SDK Initialization: timer must be positive number")
	}
	if f.IsDestroyed() {
		return errors.New("SDK Initialization: Client is destroyed")
	}
	block := make(chan int, 1)

	f.mutex.Lock()
	subscriptorName := len(f.readinessSubscriptors)
	f.mutex.Unlock()

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
		case sdkStatusReady:
			break
		case sdkInitializationFailed:
			return errors.New("SDK Initialization failed")
		}
	case <-time.After(time.Second * time.Duration(timer)):
		return fmt.Errorf("SDK Initialization: time of %d exceeded", timer)
	}

	return nil
}

// Destroy stops all async tasks and clears all storages
func (f *SplitFactory) Destroy() {
	if !f.IsDestroyed() {
		removeInstanceFromTracker(f.apikey)
	}
	f.status.Store(sdkStatusDestroyed)

	if f.cfg.OperationMode == "redis-consumer" {
		return
	}

	// Stop all tasks
	if f.tasks.splits != nil {
		f.tasks.splits.Stop()
	}
	if f.tasks.segments != nil {
		f.tasks.segments.Stop()
	}

	if f.tasks.impressions != nil {
		f.tasks.impressions.Stop()
	}
	if f.tasks.gauges != nil {
		f.tasks.gauges.Stop()
	}
	if f.tasks.counters != nil {
		f.tasks.counters.Stop()
	}
	if f.tasks.latencies != nil {
		f.tasks.latencies.Stop()
	}
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

func setupInMemoryFactory(
	apikey string,
	cfg *conf.SplitSdkConfig,
	logger logging.LoggerInterface,
	metadata *splitio.SdkMetadata,
) (*SplitFactory, error) {
	err := api.ValidateApikey(apikey, cfg.Advanced)
	if err != nil {
		return nil, err
	}

	inMememoryFullQueue := make(chan string, 2) // Size 2: So that it's able to accept one event from each resource simultaneously.

	storages := sdkStorages{
		splits:      mutexmap.NewMMSplitStorage(),
		segments:    mutexmap.NewMMSegmentStorage(),
		impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, inMememoryFullQueue, logger),
		telemetry:   mutexmap.NewMMMetricsStorage(),
		events:      mutexqueue.NewMQEventsStorage(cfg.Advanced.EventsQueueSize, inMememoryFullQueue, logger),
	}

	readyChannel := make(chan string, 1)

	syncTasks := sdkSync{
		splits: tasks.NewFetchSplitsTask(
			storages.splits.(storage.SplitStorage),
			api.NewHTTPSplitFetcher(apikey, cfg, logger),
			cfg.TaskPeriods.SplitSync,
			logger,
			readyChannel,
		),
		segments: tasks.NewFetchSegmentsTask(
			storages.splits.(storage.SplitStorage),
			storages.segments.(storage.SegmentStorage),
			api.NewHTTPSegmentFetcher(apikey, cfg, logger),
			cfg.TaskPeriods.SegmentSync,
			cfg.Advanced.SegmentWorkers,
			cfg.Advanced.SegmentQueueSize,
			logger,
			readyChannel,
		),
		impressions: tasks.NewRecordImpressionsTask(
			storages.impressions.(storage.ImpressionStorage),
			api.NewHTTPImpressionRecorder(apikey, cfg, metadata, logger),
			cfg.TaskPeriods.ImpressionSync,
			logger,
			cfg.Advanced.ImpressionsBulkSize,
		),
		counters: tasks.NewRecordCountersTask(
			storages.telemetry.(storage.MetricsStorage),
			api.NewHTTPMetricsRecorder(apikey, cfg, metadata, logger),
			cfg.TaskPeriods.CounterSync,
			logger,
		),
		gauges: tasks.NewRecordGaugesTask(
			storages.telemetry.(storage.MetricsStorage),
			api.NewHTTPMetricsRecorder(apikey, cfg, metadata, logger),
			cfg.TaskPeriods.GaugeSync,
			logger,
		),
		latencies: tasks.NewRecordLatenciesTask(
			storages.telemetry.(storage.MetricsStorage),
			api.NewHTTPMetricsRecorder(apikey, cfg, metadata, logger),
			cfg.TaskPeriods.LatencySync,
			logger,
		),
		events: tasks.NewRecordEventsTask(
			storages.events.(storage.EventsStorage),
			api.NewHTTPEventsRecorder(apikey, cfg, metadata, logger),
			cfg.Advanced.EventsBulkSize,
			cfg.TaskPeriods.EventsSync,
			logger,
		),
	}

	splitFactory := SplitFactory{
		apikey:                apikey,
		cfg:                   cfg,
		metadata:              *metadata,
		logger:                logger,
		operationMode:         "inmemory-standalone",
		storages:              storages,
		tasks:                 syncTasks,
		readinessSubscriptors: make(map[int]chan int),
	}
	splitFactory.status.Store(sdkStatusInitializing)

	go splitFactory.initializationInMemory(readyChannel, &syncTasks)
	go dataFlusher(&syncTasks, inMememoryFullQueue, logger)

	return &splitFactory, nil
}

func setupRedisFactory(
	apikey string,
	cfg *conf.SplitSdkConfig,
	logger logging.LoggerInterface,
	metadata *splitio.SdkMetadata,
) (*SplitFactory, error) {
	redisClient, err := redisdb.NewPrefixedRedisClient(&cfg.Redis)
	if err != nil {
		logger.Error("Failed to instantiate redis client.")
		return nil, err
	}

	storages := sdkStorages{
		splits:      redisdb.NewRedisSplitStorage(redisClient, logger),
		segments:    redisdb.NewRedisSegmentStorage(redisClient, logger),
		impressions: redisdb.NewRedisImpressionStorage(redisClient, metadata, logger),
		telemetry:   redisdb.NewRedisMetricsStorage(redisClient, metadata, logger),
		events:      redisdb.NewRedisEventsStorage(redisClient, metadata, logger),
	}

	factory := &SplitFactory{
		apikey:                apikey,
		cfg:                   cfg,
		metadata:              *metadata,
		logger:                logger,
		operationMode:         "redis-consumer",
		storages:              storages,
		readinessSubscriptors: make(map[int]chan int),
	}
	factory.status.Store(sdkStatusReady)
	return factory, nil
}

func setupLocalhostFactory(
	apikey string,
	cfg *conf.SplitSdkConfig,
	logger logging.LoggerInterface,
	metadata *splitio.SdkMetadata,
) (*SplitFactory, error) {
	splitStorage := mutexmap.NewMMSplitStorage()
	splitFetcher := local.NewFileSplitFetcher(cfg.SplitFile, logger)
	splitPeriod := cfg.TaskPeriods.SplitSync
	readyChannel := make(chan string, 1)

	splitFactory := &SplitFactory{
		apikey:   apikey,
		cfg:      cfg,
		metadata: *metadata,
		logger:   logger,
		storages: sdkStorages{
			splits:      splitStorage,
			impressions: mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger),
			telemetry:   mutexmap.NewMMMetricsStorage(),
			events:      mutexqueue.NewMQEventsStorage(cfg.Advanced.EventsQueueSize, make(chan string, 1), logger),
			segments:    mutexmap.NewMMSegmentStorage(),
		},
		tasks: sdkSync{
			splits: tasks.NewFetchSplitsTask(splitStorage, splitFetcher, splitPeriod, logger, readyChannel),
		},

		readinessSubscriptors: make(map[int]chan int),
	}
	splitFactory.status.Store(sdkStatusInitializing)

	// Call fetching tasks as goroutine
	go splitFactory.initializationLocalhost(readyChannel, &splitFactory.tasks)

	return splitFactory, nil
}

// newFactory instantiates a new SplitFactory object. Accepts a SplitSdkConfig struct as an argument,
// which will be used to instantiate both the client and the manager
func newFactory(apikey string, cfg *conf.SplitSdkConfig, logger logging.LoggerInterface) (*SplitFactory, error) {
	metadata := splitio.SdkMetadata{
		SDKVersion:  splitio.Version,
		MachineIP:   cfg.IPAddress,
		MachineName: cfg.InstanceName,
	}

	var splitFactory *SplitFactory
	var err error

	switch cfg.OperationMode {
	case "inmemory-standalone":
		splitFactory, err = setupInMemoryFactory(apikey, cfg, logger, &metadata)
	case "redis-consumer":
		splitFactory, err = setupRedisFactory(apikey, cfg, logger, &metadata)
	case "localhost":
		splitFactory, err = setupLocalhostFactory(apikey, cfg, logger, &metadata)
	default:
		err = fmt.Errorf("Invalid operation mode \"%s\"", cfg.OperationMode)
	}

	if err != nil {
		return nil, err
	}

	if cfg.Advanced.ImpressionListener != nil {
		splitFactory.impressionListener = impressionlistener.NewImpressionListenerWrapper(cfg.Advanced.ImpressionListener)
	}

	return splitFactory, nil
}
