package tasks

import (
	"errors"

	"github.com/splitio/go-client/splitio/service"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"
)

func submitImpressions(
	impressionStorage storage.ImpressionStorageConsumer,
	impressionRecorder service.ImpressionsRecorder,
	logger logging.LoggerInterface,
	bulkSize int64,
) error {
	queuedImpressions, err := impressionStorage.PopN(bulkSize)
	if err != nil {
		logger.Error("Error reading impressions queue", err)
		return errors.New("Error reading impressions queue")
	}

	if len(queuedImpressions) == 0 {
		logger.Debug("No impressions fetched from queue. Nothing to send")
		return nil
	}

	return impressionRecorder.Record(queuedImpressions)
}

// NewRecordImpressionsTask creates a new splits fetching and storing task
func NewRecordImpressionsTask(
	impressionStorage storage.ImpressionStorageConsumer,
	impressionRecorder service.ImpressionsRecorder,
	period int,
	logger logging.LoggerInterface,
	bulkSize int64,
) *asynctask.AsyncTask {
	record := func(logger logging.LoggerInterface) error {
		return submitImpressions(
			impressionStorage,
			impressionRecorder,
			logger,
			bulkSize,
		)
	}

	onStop := func(logger logging.LoggerInterface) {
		// All this function does is flush impressions which will clear the storage
		record(logger)
	}

	return asynctask.NewAsyncTask("SubmitImpressions", record, period, nil, onStop, logger)
}
