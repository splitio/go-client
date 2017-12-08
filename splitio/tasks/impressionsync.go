package tasks

import (
	"github.com/splitio/go-client/splitio/service"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/util/impressionlistener"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"
)

func listenerWrapper(
	impressions []dtos.ImpressionsDTO,
	listener impressionlistener.ListenerInterface,
	logger logging.LoggerInterface,
) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Impression listener is packing with the following message: ", r)
		}
	}()
	listener.Notify(impressions)
}

func submitImpressions(
	impressionStorage storage.ImpressionStorageConsumer,
	impressionRecorder service.ImpressionsRecorder,
	sdkVersion string,
	machineIP string,
	machineName string,
	listener impressionlistener.ListenerInterface,
	logger logging.LoggerInterface,
) error {
	impressions := impressionStorage.PopAll()
	if len(impressions) > 0 {
		err := impressionRecorder.Record(impressions, sdkVersion, machineIP, machineName)
		if listener != nil {
			go listenerWrapper(impressions, listener, logger)
		}
		return err
	}
	return nil
}

// NewRecordImpressionsTask creates a new splits fetching and storing task
func NewRecordImpressionsTask(
	impressionStorage storage.ImpressionStorageConsumer,
	impressionRecorder service.ImpressionsRecorder,
	period int,
	sdkVersion,
	machineIP string,
	machineName string,
	listener impressionlistener.ListenerInterface,
	logger logging.LoggerInterface,
) *asynctask.AsyncTask {
	record := func(logger logging.LoggerInterface) error {
		return submitImpressions(
			impressionStorage,
			impressionRecorder,
			sdkVersion,
			machineIP,
			machineName,
			listener,
			logger,
		)
	}

	onStop := func(logger logging.LoggerInterface) {
		// All this function does is flush impressions which will clear the storage
		record(logger)
	}

	return asynctask.NewAsyncTask("SubmitImpressions", record, period, nil, onStop, logger)
}
