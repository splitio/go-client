package impressions

import (
	"github.com/splitio/go-client/v6/splitio/conf"
	config "github.com/splitio/go-split-commons/v6/conf"
	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/go-split-commons/v6/provisional"
	"github.com/splitio/go-split-commons/v6/provisional/strategy"
	"github.com/splitio/go-split-commons/v6/service/api"
	"github.com/splitio/go-split-commons/v6/storage"
	"github.com/splitio/go-split-commons/v6/storage/filter"
	"github.com/splitio/go-split-commons/v6/synchronizer"
	"github.com/splitio/go-split-commons/v6/synchronizer/worker/impression"
	"github.com/splitio/go-split-commons/v6/synchronizer/worker/impressionscount"
	"github.com/splitio/go-split-commons/v6/tasks"
	"github.com/splitio/go-split-commons/v6/telemetry"
	"github.com/splitio/go-toolkit/v5/logging"
)

const (
	bfExpectedElemenets                = 10000000
	bfFalsePositiveProbability         = 0.01
	bfCleaningPeriod                   = 86400 // 24 hours
	uniqueKeysPeriodTaskInMemory       = 900   // 15 min
	uniqueKeysPeriodTaskRedis          = 300   // 5 min
	impressionsCountPeriodTaskInMemory = 1800  // 30 min
	impressionsCountPeriodTaskRedis    = 300   // 5 min
	impressionsBulkSizeRedis           = 100
)

func BuildInMemoryManager(
	cfg *conf.SplitSdkConfig,
	advanced config.AdvancedConfig,
	logger logging.LoggerInterface,
	splitTasks *synchronizer.SplitTasks,
	workers *synchronizer.Workers,
	metadata dtos.Metadata,
	splitAPI *api.SplitAPI,
	telemetryStorage storage.TelemetryRuntimeProducer,
	impressionStorage storage.ImpressionStorageConsumer,
) (provisional.ImpressionManager, error) {
	listenerEnabled := cfg.Advanced.ImpressionListener != nil
	impressionsCounter := strategy.NewImpressionsCounter()
	filter := filter.NewBloomFilter(bfExpectedElemenets, bfFalsePositiveProbability)
	uniqueKeysTracker := strategy.NewUniqueKeysTracker(filter)

	workers.ImpressionsCountRecorder = impressionscount.NewRecorderSingle(impressionsCounter, splitAPI.ImpressionRecorder, metadata, logger, telemetryStorage)

	splitTasks.ImpressionsCountSyncTask = tasks.NewRecordImpressionsCountTask(workers.ImpressionsCountRecorder, logger, impressionsCountPeriodTaskInMemory)
	splitTasks.UniqueKeysTask = tasks.NewRecordUniqueKeysTask(workers.TelemetryRecorder, uniqueKeysTracker, uniqueKeysPeriodTaskInMemory, logger)
	splitTasks.CleanFilterTask = tasks.NewCleanFilterTask(filter, logger, bfCleaningPeriod)

	noneStrategy := strategy.NewNoneImpl(impressionsCounter, uniqueKeysTracker, listenerEnabled)

	if cfg.ImpressionsMode == config.ImpressionsModeNone {
		impManager := provisional.NewImpressionManagerImp(noneStrategy, noneStrategy)
		return impManager, nil
	}

	workers.ImpressionRecorder = impression.NewRecorderSingle(impressionStorage, splitAPI.ImpressionRecorder, logger, metadata, cfg.ImpressionsMode, telemetryStorage)
	splitTasks.ImpressionSyncTask = tasks.NewRecordImpressionsTask(workers.ImpressionRecorder, cfg.TaskPeriods.ImpressionSync, logger, advanced.ImpressionsBulkSize)

	impressionObserver, err := strategy.NewImpressionObserver(500)
	if err != nil {
		return nil, err
	}

	var impressionsStrategy strategy.ProcessStrategyInterface
	switch cfg.ImpressionsMode {
	case config.ImpressionsModeDebug:
		impressionsStrategy = strategy.NewDebugImpl(impressionObserver, listenerEnabled)
	default:
		impressionsStrategy = strategy.NewOptimizedImpl(impressionObserver, impressionsCounter, telemetryStorage, listenerEnabled)
	}

	manager := provisional.NewImpressionManagerImp(noneStrategy, impressionsStrategy)

	return manager, nil
}

func BuildRedisManager(
	cfg *conf.SplitSdkConfig,
	logger logging.LoggerInterface,
	splitTasks *synchronizer.SplitTasks,
	telemetryConfigStorage storage.TelemetryConfigProducer,
	impressionsCountStorage storage.ImpressionsCountProducer,
	telemetryRuntimeStorage storage.TelemetryRuntimeProducer,
) (provisional.ImpressionManager, error) {
	listenerEnabled := cfg.Advanced.ImpressionListener != nil

	impressionsCounter := strategy.NewImpressionsCounter()
	filter := filter.NewBloomFilter(bfExpectedElemenets, bfFalsePositiveProbability)
	uniqueKeysTracker := strategy.NewUniqueKeysTracker(filter)

	telemetryRecorder := telemetry.NewSynchronizerRedis(telemetryConfigStorage, logger)
	impressionsCountRecorder := impressionscount.NewRecorderRedis(impressionsCounter, impressionsCountStorage, logger)

	splitTasks.ImpressionsCountSyncTask = tasks.NewRecordImpressionsCountTask(impressionsCountRecorder, logger, impressionsCountPeriodTaskRedis)
	splitTasks.UniqueKeysTask = tasks.NewRecordUniqueKeysTask(telemetryRecorder, uniqueKeysTracker, uniqueKeysPeriodTaskRedis, logger)
	splitTasks.CleanFilterTask = tasks.NewCleanFilterTask(filter, logger, bfCleaningPeriod)

	noneStrategy := strategy.NewNoneImpl(impressionsCounter, uniqueKeysTracker, listenerEnabled)

	if cfg.ImpressionsMode == config.ImpressionsModeNone {
		impManager := provisional.NewImpressionManagerImp(noneStrategy, noneStrategy)
		return impManager, nil
	}

	impressionObserver, err := strategy.NewImpressionObserver(500)
	if err != nil {
		return nil, err
	}

	var impressionsStrategy strategy.ProcessStrategyInterface
	switch cfg.ImpressionsMode {
	case config.ImpressionsModeDebug:
		impressionsStrategy = strategy.NewDebugImpl(impressionObserver, listenerEnabled)
	default:
		impressionsStrategy = strategy.NewOptimizedImpl(impressionObserver, impressionsCounter, telemetryRuntimeStorage, listenerEnabled)
	}

	manager := provisional.NewImpressionManagerImp(noneStrategy, impressionsStrategy)

	return manager, nil
}
