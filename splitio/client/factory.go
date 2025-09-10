// Package client contains implementations of the Split SDK client and the factory used
// to instantiate it.
package client

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/splitio/go-client/v6/splitio"
	"github.com/splitio/go-client/v6/splitio/conf"
	impressionlistener "github.com/splitio/go-client/v6/splitio/impressionListener"
	"github.com/splitio/go-client/v6/splitio/impressions"

	config "github.com/splitio/go-split-commons/v6/conf"
	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/go-split-commons/v6/engine"
	"github.com/splitio/go-split-commons/v6/engine/evaluator"
	"github.com/splitio/go-split-commons/v6/flagsets"
	"github.com/splitio/go-split-commons/v6/healthcheck/application"
	"github.com/splitio/go-split-commons/v6/provisional"
	"github.com/splitio/go-split-commons/v6/provisional/strategy"
	"github.com/splitio/go-split-commons/v6/service/api"
	"github.com/splitio/go-split-commons/v6/service/api/specs"
	"github.com/splitio/go-split-commons/v6/service/local"
	"github.com/splitio/go-split-commons/v6/storage"
	"github.com/splitio/go-split-commons/v6/storage/inmemory"
	"github.com/splitio/go-split-commons/v6/storage/inmemory/mutexmap"
	"github.com/splitio/go-split-commons/v6/storage/inmemory/mutexqueue"
	"github.com/splitio/go-split-commons/v6/storage/mocks"
	"github.com/splitio/go-split-commons/v6/storage/redis"
	"github.com/splitio/go-split-commons/v6/synchronizer"
	"github.com/splitio/go-split-commons/v6/synchronizer/worker/event"
	"github.com/splitio/go-split-commons/v6/synchronizer/worker/segment"
	"github.com/splitio/go-split-commons/v6/synchronizer/worker/split"
	"github.com/splitio/go-split-commons/v6/tasks"
	"github.com/splitio/go-split-commons/v6/telemetry"
	"github.com/splitio/go-toolkit/v5/logging"
)

const (
	sdkStatusDestroyed = iota
	sdkStatusInitializing
	sdkStatusReady

	sdkInitializationFailed = -1
)

type sdkStorages struct {
	splits              storage.SplitStorageConsumer
	segments            storage.SegmentStorageConsumer
	ruleBasedSegments   storage.RuleBasedSegmentStorageConsumer
	impressionsConsumer storage.ImpressionStorageConsumer
	impressions         storage.ImpressionStorageProducer
	events              storage.EventStorageProducer
	initTelemetry       storage.TelemetryConfigProducer
	runtimeTelemetry    storage.TelemetryRuntimeProducer
	evaluationTelemetry storage.TelemetryEvaluationProducer
	impressionsCount    storage.ImpressionsCountProducer
}

// SplitFactory struct is responsible for instantiating and storing instances of client and manager.
type SplitFactory struct {
	startTime             time.Time // Tracking startTime
	metadata              dtos.Metadata
	storages              sdkStorages
	apikey                string
	status                atomic.Value
	readinessSubscriptors map[int]chan int
	operationMode         string
	mutex                 sync.Mutex
	cfg                   *conf.SplitSdkConfig
	impressionListener    *impressionlistener.WrapperImpressionListener
	logger                logging.LoggerInterface
	syncManager           synchronizer.Manager
	telemetrySync         telemetry.TelemetrySynchronizer // To execute SynchronizeInit
	impressionManager     provisional.ImpressionManager
}

// Client returns the split client instantiated by the factory
func (f *SplitFactory) Client() *SplitClient {
	return &SplitClient{
		logger:      f.logger,
		evaluator:   evaluator.NewEvaluator(f.storages.splits, f.storages.segments, f.storages.ruleBasedSegments, engine.NewEngine(f.logger), f.logger),
		impressions: f.storages.impressions,
		events:      f.storages.events,
		validator: inputValidation{
			logger:       f.logger,
			splitStorage: f.storages.splits,
		},
		factory:             f,
		impressionListener:  f.impressionListener,
		impressionManager:   f.impressionManager,
		initTelemetry:       f.storages.initTelemetry,       // For capturing NonReadyUsages
		runtimeTelemetry:    f.storages.runtimeTelemetry,    // For capturing runtime stats
		evaluationTelemetry: f.storages.evaluationTelemetry, // For capturing treatment stats
	}
}

// Manager returns the split manager instantiated by the factory
func (f *SplitFactory) Manager() *SplitManager {
	return &SplitManager{
		splitStorage:  f.storages.splits,
		validator:     inputValidation{logger: f.logger},
		logger:        f.logger,
		factory:       f,
		initTelemetry: f.storages.initTelemetry, // For capturing NonReadyUsages
	}
}

// IsDestroyed returns true if the client has been destroyed
func (f *SplitFactory) IsDestroyed() bool {
	return f.status.Load() == sdkStatusDestroyed
}

// IsReady returns true if the factory is ready
func (f *SplitFactory) IsReady() bool {
	return f.status.Load() == sdkStatusReady
}

// initializates tasks for in-memory mode
func (f *SplitFactory) initializationManager(readyChannel chan int, flagSetsInvalid int64) {
	go f.syncManager.Start()
	msg := <-readyChannel
	switch msg {
	case synchronizer.Ready:
		// Broadcast ready status for SDK
		f.broadcastReadiness(sdkStatusReady, make([]string, 0), flagSetsInvalid)
	default:
		f.broadcastReadiness(sdkInitializationFailed, make([]string, 0), flagSetsInvalid)
	}
}

func (f *SplitFactory) initializationRedis() {
	go f.syncManager.Start()
	f.broadcastReadiness(sdkStatusReady, make([]string, 0), 0)
}

// recordInitTelemetry In charge of recording init stats from redis and memory
func (f *SplitFactory) recordInitTelemetry(tags []string, currentFactories map[string]int64, flagSetsInvalid int64) {
	f.logger.Debug("Sending init telemetry")
	f.telemetrySync.SynchronizeConfig(
		telemetry.InitConfig{
			AdvancedConfig: config.AdvancedConfig{
				HTTPTimeout:          f.cfg.Advanced.HTTPTimeout,
				SegmentQueueSize:     f.cfg.Advanced.SegmentQueueSize,
				SegmentWorkers:       f.cfg.Advanced.SegmentWorkers,
				SdkURL:               f.cfg.Advanced.SdkURL,
				EventsURL:            f.cfg.Advanced.EventsURL,
				TelemetryServiceURL:  f.cfg.Advanced.TelemetryServiceURL,
				EventsBulkSize:       f.cfg.Advanced.EventsBulkSize,
				EventsQueueSize:      f.cfg.Advanced.EventsQueueSize,
				ImpressionsQueueSize: f.cfg.Advanced.ImpressionsQueueSize,
				ImpressionsBulkSize:  f.cfg.Advanced.ImpressionsBulkSize,
				StreamingEnabled:     f.cfg.Advanced.StreamingEnabled,
				AuthServiceURL:       f.cfg.Advanced.AuthServiceURL,
				StreamingServiceURL:  f.cfg.Advanced.StreamingServiceURL,
			},
			TaskPeriods:     config.TaskPeriods(f.cfg.TaskPeriods),
			ImpressionsMode: f.cfg.ImpressionsMode,
			ListenerEnabled: f.cfg.Advanced.ImpressionListener != nil,
			FlagSetsTotal:   int64(len(f.cfg.Advanced.FlagSetsFilter)),
			FlagSetsInvalid: flagSetsInvalid,
		},
		time.Now().UTC().Sub(f.startTime).Milliseconds(),
		currentFactories,
		tags,
	)
}

// broadcastReadiness broadcasts message to all the subscriptors
func (f *SplitFactory) broadcastReadiness(status int, tags []string, flagSetsInvalid int64) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.status.Load() == sdkStatusInitializing && status == sdkStatusReady {
		f.status.Store(sdkStatusReady)
	}
	for _, subscriptor := range f.readinessSubscriptors {
		subscriptor <- status
	}
	// At this point the SDK is ready for sending telemetry
	go f.recordInitTelemetry(tags, getFactories(), flagSetsInvalid)
}

// subscribes listener
func (f *SplitFactory) subscribe(name int, subscriptor chan int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.readinessSubscriptors[name] = subscriptor
}

// removes a particular subscriptor from the list
func (f *SplitFactory) unsubscribe(name int) {
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
			f.unsubscribe(subscriptorName)
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
		f.storages.initTelemetry.RecordBURTimeout() // Records BURTimeout
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
	if f.storages.runtimeTelemetry != nil {
		f.storages.runtimeTelemetry.RecordSessionLength(int64(time.Since(f.startTime) * time.Millisecond))
	}
	f.syncManager.Stop()
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
	metadata dtos.Metadata,
) (*SplitFactory, error) {
	advanced, warnings := conf.NormalizeSDKConf(cfg.Advanced)
	advanced.AuthSpecVersion = specs.FLAG_V1_3
	advanced.FlagsSpecVersion = specs.FLAG_V1_3
	printWarnings(logger, warnings)
	flagSetsInvalid := int64(len(cfg.Advanced.FlagSetsFilter) - len(advanced.FlagSetsFilter))
	if strings.TrimSpace(cfg.SplitSyncProxyURL) != "" {
		advanced.StreamingEnabled = false
	}

	inMememoryFullQueue := make(chan string, 2) // Size 2: So that it's able to accept one event from each resource simultaneously.

	flagSetFilter := flagsets.NewFlagSetFilter(advanced.FlagSetsFilter)
	splitsStorage := mutexmap.NewMMSplitStorage(flagSetFilter)
	segmentsStorage := mutexmap.NewMMSegmentStorage()
	ruleBasedSegmentStorage := mutexmap.NewRuleBasedSegmentsStorage()
	telemetryStorage, err := inmemory.NewTelemetryStorage()
	impressionsStorage := mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, inMememoryFullQueue, logger, telemetryStorage)
	eventsStorage := mutexqueue.NewMQEventsStorage(cfg.Advanced.EventsQueueSize, inMememoryFullQueue, logger, telemetryStorage)
	if err != nil {
		return nil, err
	}

	var dummyHC = &application.Dummy{}

	splitAPI := api.NewSplitAPI(apikey, advanced, logger, metadata)
	workers := synchronizer.Workers{
		SplitUpdater:      split.NewSplitUpdater(splitsStorage, ruleBasedSegmentStorage, splitAPI.SplitFetcher, logger, telemetryStorage, dummyHC, flagSetFilter),
		SegmentUpdater:    segment.NewSegmentUpdater(splitsStorage, segmentsStorage, splitAPI.SegmentFetcher, logger, telemetryStorage, dummyHC),
		EventRecorder:     event.NewEventRecorderSingle(eventsStorage, splitAPI.EventRecorder, logger, metadata, telemetryStorage),
		TelemetryRecorder: telemetry.NewTelemetrySynchronizer(telemetryStorage, splitAPI.TelemetryRecorder, splitsStorage, segmentsStorage, logger, metadata, telemetryStorage),
	}
	splitTasks := synchronizer.SplitTasks{
		SplitSyncTask:     tasks.NewFetchSplitsTask(workers.SplitUpdater, cfg.TaskPeriods.SplitSync, logger),
		SegmentSyncTask:   tasks.NewFetchSegmentsTask(workers.SegmentUpdater, cfg.TaskPeriods.SegmentSync, advanced.SegmentWorkers, advanced.SegmentQueueSize, logger, dummyHC),
		EventSyncTask:     tasks.NewRecordEventsTask(workers.EventRecorder, advanced.EventsBulkSize, cfg.TaskPeriods.EventsSync, logger),
		TelemetrySyncTask: tasks.NewRecordTelemetryTask(workers.TelemetryRecorder, cfg.TaskPeriods.TelemetrySync, logger),
	}

	storages := sdkStorages{
		splits:              splitsStorage,
		events:              eventsStorage,
		impressionsConsumer: impressionsStorage,
		impressions:         impressionsStorage,
		segments:            segmentsStorage,
		ruleBasedSegments:   ruleBasedSegmentStorage,
		initTelemetry:       telemetryStorage,
		evaluationTelemetry: telemetryStorage,
		runtimeTelemetry:    telemetryStorage,
	}

	if cfg.ImpressionsMode == "" {
		cfg.ImpressionsMode = config.ImpressionsModeOptimized
	}

	impressionManager, err := impressions.BuildInMemoryManager(cfg, advanced, logger, &splitTasks, &workers, metadata, splitAPI, storages.runtimeTelemetry, storages.impressionsConsumer)
	if err != nil {
		return nil, err
	}

	syncImpl := synchronizer.NewSynchronizer(
		advanced,
		splitTasks,
		workers,
		logger,
		inMememoryFullQueue,
	)

	readyChannel := make(chan int, 1)
	clientKey := apikey[len(apikey)-4:]
	syncManager, err := synchronizer.NewSynchronizerManager(
		syncImpl,
		logger,
		advanced,
		splitAPI.AuthClient,
		splitsStorage,
		readyChannel,
		telemetryStorage,
		metadata,
		&clientKey,
		dummyHC,
	)
	if err != nil {
		return nil, err
	}

	splitFactory := SplitFactory{
		startTime:             time.Now().UTC(),
		apikey:                apikey,
		cfg:                   cfg,
		metadata:              metadata,
		logger:                logger,
		operationMode:         conf.InMemoryStandAlone,
		storages:              storages,
		readinessSubscriptors: make(map[int]chan int),
		syncManager:           syncManager,
		telemetrySync:         workers.TelemetryRecorder,
		impressionManager:     impressionManager,
	}
	splitFactory.status.Store(sdkStatusInitializing)
	setFactory(splitFactory.apikey, splitFactory.logger)

	go splitFactory.initializationManager(readyChannel, flagSetsInvalid)

	return &splitFactory, nil
}

func setupRedisFactory(apikey string, cfg *conf.SplitSdkConfig, logger logging.LoggerInterface, metadata dtos.Metadata) (*SplitFactory, error) {
	redisClient, err := redis.NewRedisClient(&cfg.Redis, logger)
	if err != nil {
		logger.Error("Failed to instantiate redis client.")
		return nil, err
	}

	telemetryStorage := redis.NewTelemetryStorage(redisClient, logger, metadata)
	runtimeTelemetry := mocks.MockTelemetryStorage{
		RecordSyncLatencyCall:      func(resource int, latency time.Duration) {},
		RecordImpressionsStatsCall: func(dataType int, count int64) {},
		RecordSessionLengthCall:    func(session int64) {},
	}
	inMememoryFullQueue := make(chan string, 2) // Size 2: So that it's able to accept one event from each resource simultaneously.
	impressionStorage := redis.NewImpressionStorage(redisClient, metadata, logger)

	if len(cfg.Advanced.FlagSetsFilter) != 0 {
		cfg.Advanced.FlagSetsFilter = []string{}
		logger.Warning("FlagSets filter is not applicable for Consumer modes where the SDK does not keep rollout data in sync. FlagSet filter was discarded")
	}
	flagSetFilter := flagsets.NewFlagSetFilter([]string{})

	storages := sdkStorages{
		splits:              redis.NewSplitStorage(redisClient, logger, flagSetFilter),
		segments:            redis.NewSegmentStorage(redisClient, logger),
		impressionsConsumer: impressionStorage,
		impressions:         impressionStorage,
		events:              redis.NewEventsStorage(redisClient, metadata, logger),
		initTelemetry:       telemetryStorage,
		evaluationTelemetry: telemetryStorage,
		impressionsCount:    redis.NewImpressionsCountStorage(redisClient, logger),
		runtimeTelemetry:    runtimeTelemetry,
	}

	splitTasks := synchronizer.SplitTasks{}
	workers := synchronizer.Workers{}
	advanced := config.AdvancedConfig{}

	if cfg.ImpressionsMode == "" {
		cfg.ImpressionsMode = config.ImpressionsModeDebug
	}

	impressionManager, err := impressions.BuildRedisManager(cfg, logger, &splitTasks, storages.initTelemetry, storages.impressionsCount, storages.runtimeTelemetry)
	if err != nil {
		return nil, err
	}

	syncImpl := synchronizer.NewSynchronizer(
		advanced,
		splitTasks,
		workers,
		logger,
		inMememoryFullQueue,
	)

	syncManager := synchronizer.NewSynchronizerManagerRedis(syncImpl, logger)

	factory := &SplitFactory{
		startTime:             time.Now().UTC(),
		apikey:                apikey,
		cfg:                   cfg,
		metadata:              metadata,
		logger:                logger,
		operationMode:         conf.RedisConsumer,
		storages:              storages,
		readinessSubscriptors: make(map[int]chan int),
		telemetrySync:         telemetry.NewSynchronizerRedis(telemetryStorage, logger),
		impressionManager:     impressionManager,
		syncManager:           syncManager,
	}
	factory.status.Store(sdkStatusInitializing)
	setFactory(factory.apikey, factory.logger)

	factory.initializationRedis()

	return factory, nil
}

func setupLocalhostFactory(
	apikey string,
	cfg *conf.SplitSdkConfig,
	logger logging.LoggerInterface,
	metadata dtos.Metadata,
) (*SplitFactory, error) {
	flagSets, errs := flagsets.SanitizeMany(cfg.Advanced.FlagSetsFilter)
	flagSetsInvalid := int64(len(cfg.Advanced.FlagSetsFilter) - len(flagSets))
	printWarnings(logger, errs)
	flagSetFilter := flagsets.NewFlagSetFilter(flagSets)
	splitStorage := mutexmap.NewMMSplitStorage(flagSetFilter)
	segmentStorage := mutexmap.NewMMSegmentStorage()
	ruleBasedSegmentStorage := mutexmap.NewRuleBasedSegmentsStorage()
	telemetryStorage, err := inmemory.NewTelemetryStorage()
	if err != nil {
		return nil, err
	}
	readyChannel := make(chan int, 1)
	fileFormat := local.DefineFormat(cfg.SplitFile, logger)
	splitAPI := &api.SplitAPI{SplitFetcher: local.NewFileSplitFetcher(cfg.SplitFile, logger, fileFormat)}

	if cfg.SegmentDirectory != "" {
		splitAPI.SegmentFetcher = local.NewFileSegmentFetcher(cfg.SegmentDirectory, logger)
	}

	var dummyHC = &application.Dummy{}

	localConfig := &synchronizer.LocalConfig{
		SplitPeriod:      cfg.TaskPeriods.SplitSync,
		SegmentPeriod:    cfg.TaskPeriods.SegmentSync,
		SegmentWorkers:   cfg.Advanced.SegmentWorkers,
		QueueSize:        cfg.Advanced.SegmentQueueSize,
		SegmentDirectory: cfg.SegmentDirectory,
		RefreshEnabled:   cfg.LocalhostRefreshEnabled,
	}

	syncManager, err := synchronizer.NewSynchronizerManager(
		synchronizer.NewLocal(localConfig, splitAPI, splitStorage, segmentStorage, ruleBasedSegmentStorage, logger, telemetryStorage, dummyHC),
		logger,
		config.AdvancedConfig{StreamingEnabled: false},
		nil,
		splitStorage,
		readyChannel,
		telemetryStorage,
		metadata,
		nil,
		dummyHC,
	)

	if err != nil {
		return nil, err
	}

	splitFactory := &SplitFactory{
		startTime: time.Now().UTC(),
		apikey:    apikey,
		cfg:       cfg,
		metadata:  metadata,
		logger:    logger,
		storages: sdkStorages{
			splits:              splitStorage,
			impressions:         mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 1), logger, telemetryStorage),
			events:              mutexqueue.NewMQEventsStorage(cfg.Advanced.EventsQueueSize, make(chan string, 1), logger, telemetryStorage),
			ruleBasedSegments:   ruleBasedSegmentStorage,
			segments:            segmentStorage,
			initTelemetry:       telemetryStorage,
			evaluationTelemetry: telemetryStorage,
			runtimeTelemetry:    telemetryStorage,
		},
		readinessSubscriptors: make(map[int]chan int),
		syncManager:           syncManager,
		telemetrySync:         &telemetry.NoOp{},
	}
	splitFactory.status.Store(sdkStatusInitializing)

	impressionObserver, err := strategy.NewImpressionObserver(500)
	if err != nil {
		return nil, err
	}
	impressionsStrategy := strategy.NewDebugImpl(impressionObserver, cfg.Advanced.ImpressionListener != nil)
	splitFactory.impressionManager = provisional.NewImpressionManager(impressionsStrategy).(*provisional.ImpressionManagerImpl)
	setFactory(splitFactory.apikey, splitFactory.logger)

	// Call fetching tasks as goroutine
	go splitFactory.initializationManager(readyChannel, flagSetsInvalid)

	return splitFactory, nil
}

// newFactory instantiates a new SplitFactory object. Accepts a SplitSdkConfig struct as an argument,
// which will be used to instantiate both the client and the manager
func newFactory(apikey string, cfg conf.SplitSdkConfig, logger logging.LoggerInterface) (*SplitFactory, error) {
	metadata := dtos.Metadata{
		SDKVersion:  "go-" + splitio.Version,
		MachineIP:   cfg.IPAddress,
		MachineName: cfg.InstanceName,
	}

	var splitFactory *SplitFactory
	var err error

	switch cfg.OperationMode {
	case conf.InMemoryStandAlone:
		splitFactory, err = setupInMemoryFactory(apikey, &cfg, logger, metadata)
	case conf.RedisConsumer:
		splitFactory, err = setupRedisFactory(apikey, &cfg, logger, metadata)
	case conf.Localhost:
		splitFactory, err = setupLocalhostFactory(apikey, &cfg, logger, metadata)
	default:
		err = fmt.Errorf("Invalid operation mode \"%s\"", cfg.OperationMode)
	}

	if err != nil {
		return nil, err
	}

	if cfg.Advanced.ImpressionListener != nil {
		splitFactory.impressionListener = impressionlistener.NewImpressionListenerWrapper(
			cfg.Advanced.ImpressionListener,
			metadata,
		)
	}

	return splitFactory, nil
}

func printWarnings(logger logging.LoggerInterface, errs []error) {
	if len(errs) != 0 {
		for _, err := range errs {
			if errType, ok := err.(dtos.FlagSetValidatonError); ok {
				logger.Warning(errType.Message)
			}
		}
	}
}
