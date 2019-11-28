package tasks

import (
	"errors"

	"github.com/splitio/go-client/splitio/service"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"
)

func submitEvents(
	eventStorage storage.EventStorageConsumer,
	eventRecorder service.EventsRecorder,
	bulkSize int64,
	logger logging.LoggerInterface,
) error {
	queuedEvents, err := eventStorage.PopN(bulkSize)
	if err != nil {
		logger.Error("Error reading events queue", err)
		return errors.New("Error reading events queue")
	}

	if len(queuedEvents) == 0 {
		logger.Debug("No events fetched from queue. Nothing to send")
		return nil
	}

	return eventRecorder.Record(queuedEvents)
}

func onStopAction(
	eventStorage storage.EventStorageConsumer,
	eventRecorder service.EventsRecorder,
	bulkSize int64,
	logger logging.LoggerInterface,
) {

	for !eventStorage.Empty() {
		submitEvents(
			eventStorage,
			eventRecorder,
			bulkSize,
			logger,
		)
	}

}

// NewRecordEventsTask creates a new events recording task
func NewRecordEventsTask(
	eventStorage storage.EventStorageConsumer,
	eventRecorder service.EventsRecorder,
	bulkSize int64,
	period int,
	logger logging.LoggerInterface,
) *asynctask.AsyncTask {
	record := func(logger logging.LoggerInterface) error {
		return submitEvents(eventStorage, eventRecorder, bulkSize, logger)
	}

	onStop := func(logger logging.LoggerInterface) {
		// All this function does is flush events which will clear the storage
		//record(logger)
		onStopAction(eventStorage, eventRecorder, bulkSize, logger)
	}

	return asynctask.NewAsyncTask("SubmitEvents", record, period, nil, onStop, logger)
}
