package tasks

import (
	"github.com/splitio/go-client/splitio/service"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/util/logging"
	"github.com/splitio/go-toolkit/asynctask"
)

func submitImpressions(
	impressionStorage storage.ImpressionStorage,
	impressionRecorder service.ImpressionsRecorder,
	sdkVersion string,
	machineIP string,
	machineName string,
) error {
	impressions := impressionStorage.PopAll()
	if len(impressions) > 0 {
		err := impressionRecorder.Record(impressions, sdkVersion, machineIP, machineName)
		return err
	}
	return nil
}

// NewRecordImpressionsTask creates a new splits fetching and storing task
func NewRecordImpressionsTask(
	impressionStorage storage.ImpressionStorage,
	impressionRecorder service.ImpressionsRecorder,
	period int64,
	sdkVersion,
	machineIP string,
	machineName string,
	logger logging.LoggerInterface,
) *asynctask.AsyncTask {
	record := func(logger logging.LoggerInterface) error {
		return submitImpressions(
			impressionStorage,
			impressionRecorder,
			sdkVersion,
			machineIP,
			machineName,
		)
	}
	return asynctask.NewAsyncTask("SubmitImpressions", record, period, nil, logger)
}
