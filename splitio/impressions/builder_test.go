package impressions

import (
	"testing"

	"github.com/splitio/go-client/v6/splitio/conf"
	config "github.com/splitio/go-split-commons/v9/conf"
	"github.com/splitio/go-split-commons/v9/dtos"
	"github.com/splitio/go-split-commons/v9/service/api"
	"github.com/splitio/go-split-commons/v9/storage/inmemory"
	"github.com/splitio/go-split-commons/v9/storage/inmemory/mutexqueue"
	"github.com/splitio/go-split-commons/v9/storage/mocks"
	"github.com/splitio/go-split-commons/v9/synchronizer"
	"github.com/splitio/go-toolkit/v5/logging"
)

func TestBuildInMemoryWithNone(t *testing.T) {
	cfg := conf.SplitSdkConfig{
		ImpressionsMode: config.ImpressionsModeNone,
	}
	advanced, _ := conf.NormalizeSDKConf(cfg.Advanced)
	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitTasks := synchronizer.SplitTasks{}
	workers := synchronizer.Workers{}
	metadata := dtos.Metadata{}
	splitAPI := api.NewSplitAPI("apikey", advanced, logger, metadata)
	telemetryStorage, _ := inmemory.NewTelemetryStorage()
	impressionsStorage := mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 2), logger, telemetryStorage)

	impManager, err := BuildInMemoryManager(&cfg, advanced, logger, &splitTasks, &workers, metadata, splitAPI, telemetryStorage, impressionsStorage)
	if err != nil {
		t.Error("err should be nil. ", err.Error())
	}
	if impManager == nil {
		t.Error("impManager should not be nil. ")
	}
	if workers.ImpressionsCountRecorder == nil {
		t.Error("ImpressionRecorder should not be nil. ")
	}
	if splitTasks.ImpressionsCountSyncTask == nil {
		t.Error("ImpressionsCountSyncTask should not be nil. ")
	}
	if splitTasks.UniqueKeysTask == nil {
		t.Error("UniqueKeysTask should not be nil. ")
	}
	if splitTasks.CleanFilterTask == nil {
		t.Error("CleanFilterTask should not be nil. ")
	}

	if workers.ImpressionRecorder != nil {
		t.Error("ImpressionRecorder should be nil. ")
	}
	if splitTasks.ImpressionSyncTask != nil {
		t.Error("ImpressionSyncTask should be nil. ")
	}
}

func TestBuildInMemoryWithDebug(t *testing.T) {
	cfg := conf.SplitSdkConfig{
		ImpressionsMode: config.ImpressionsModeDebug,
	}
	advanced, _ := conf.NormalizeSDKConf(cfg.Advanced)
	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitTasks := synchronizer.SplitTasks{}
	workers := synchronizer.Workers{}
	metadata := dtos.Metadata{}
	splitAPI := api.NewSplitAPI("apikey", advanced, logger, metadata)
	telemetryStorage, _ := inmemory.NewTelemetryStorage()
	impressionsStorage := mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 2), logger, telemetryStorage)

	impManager, err := BuildInMemoryManager(&cfg, advanced, logger, &splitTasks, &workers, metadata, splitAPI, telemetryStorage, impressionsStorage)
	if err != nil {
		t.Error("err should be nil. ", err.Error())
	}
	if impManager == nil {
		t.Error("impManager should not be nil. ")
	}
	if workers.ImpressionsCountRecorder == nil {
		t.Error("ImpressionRecorder should not be nil. ")
	}
	if splitTasks.ImpressionsCountSyncTask == nil {
		t.Error("ImpressionsCountSyncTask should not be nil. ")
	}
	if splitTasks.UniqueKeysTask == nil {
		t.Error("UniqueKeysTask should not be nil. ")
	}
	if splitTasks.CleanFilterTask == nil {
		t.Error("CleanFilterTask should not be nil. ")
	}
	if workers.ImpressionRecorder == nil {
		t.Error("ImpressionRecorder should be nil. ")
	}
	if splitTasks.ImpressionSyncTask == nil {
		t.Error("ImpressionSyncTask should be nil. ")
	}
}

func TestBuildInMemoryWithOptimized(t *testing.T) {
	cfg := conf.SplitSdkConfig{
		ImpressionsMode: config.ImpressionsModeOptimized,
	}
	advanced, _ := conf.NormalizeSDKConf(cfg.Advanced)
	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitTasks := synchronizer.SplitTasks{}
	workers := synchronizer.Workers{}
	metadata := dtos.Metadata{}
	splitAPI := api.NewSplitAPI("apikey", advanced, logger, metadata)
	telemetryStorage, _ := inmemory.NewTelemetryStorage()
	impressionsStorage := mutexqueue.NewMQImpressionsStorage(cfg.Advanced.ImpressionsQueueSize, make(chan string, 2), logger, telemetryStorage)

	impManager, err := BuildInMemoryManager(&cfg, advanced, logger, &splitTasks, &workers, metadata, splitAPI, telemetryStorage, impressionsStorage)
	if err != nil {
		t.Error("err should be nil. ", err.Error())
	}
	if impManager == nil {
		t.Error("impManager should not be nil. ")
	}
	if workers.ImpressionsCountRecorder == nil {
		t.Error("ImpressionRecorder should not be nil. ")
	}
	if splitTasks.ImpressionsCountSyncTask == nil {
		t.Error("ImpressionsCountSyncTask should not be nil. ")
	}
	if splitTasks.UniqueKeysTask == nil {
		t.Error("UniqueKeysTask should not be nil. ")
	}
	if splitTasks.CleanFilterTask == nil {
		t.Error("CleanFilterTask should not be nil. ")
	}
	if workers.ImpressionRecorder == nil {
		t.Error("ImpressionRecorder should be nil. ")
	}
	if splitTasks.ImpressionSyncTask == nil {
		t.Error("ImpressionSyncTask should be nil. ")
	}
}

func TestBuildRedisWithNone(t *testing.T) {
	cfg := conf.SplitSdkConfig{
		ImpressionsMode: config.ImpressionsModeNone,
	}
	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitTasks := synchronizer.SplitTasks{}
	runtimeTelemetry := mocks.MockTelemetryStorage{}
	impressionCountStorage := mocks.MockImpressionsCountStorage{}

	impManager, err := BuildRedisManager(&cfg, logger, &splitTasks, runtimeTelemetry, impressionCountStorage, runtimeTelemetry)
	if err != nil {
		t.Error("err should be nil. ", err.Error())
	}
	if impManager == nil {
		t.Error("impManager should not be nil. ")
	}
}
