package tasks

import (
	"errors"
	"github.com/splitio/go-client/splitio/service"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"
)

func submitCounters(
	metricsStorage storage.MetricsStorage,
	metricsRecorder service.MetricsRecorder,
	sdkVersion string,
	machineIP string,
	machineName string,
) error {
	counters := metricsStorage.PopCounters()
	if len(counters) > 0 {
		err := metricsRecorder.RecordCounters(counters, sdkVersion, machineIP, machineName)
		return err
	}
	return nil
}

func submitGauges(
	metricsStorage storage.MetricsStorage,
	metricsRecorder service.MetricsRecorder,
	sdkVersion string,
	machineIP string,
	machineName string,
) error {
	var errs []error
	for _, gauge := range metricsStorage.PopGauges() {
		err := metricsRecorder.RecordGauge(gauge, sdkVersion, machineIP, machineName)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.New("Some gauges could not be posted")
	}
	return nil
}

func submitLatencies(
	metricsStorage storage.MetricsStorage,
	metricsRecorder service.MetricsRecorder,
	sdkVersion string,
	machineIP string,
	machineName string,
) error {
	latencies := metricsStorage.PopLatencies()
	if len(latencies) > 0 {
		err := metricsRecorder.RecordLatencies(latencies, sdkVersion, machineIP, machineName)
		return err
	}
	return nil
}

// NewRecordCountersTask creates a new splits fetching and storing task
func NewRecordCountersTask(
	metricsStorage storage.MetricsStorage,
	metricsRecorder service.MetricsRecorder,
	period int,
	sdkVersion,
	machineIP string,
	machineName string,
	logger logging.LoggerInterface,
) *asynctask.AsyncTask {
	record := func(logger logging.LoggerInterface) error {
		return submitCounters(
			metricsStorage,
			metricsRecorder,
			sdkVersion,
			machineIP,
			machineName,
		)
	}
	return asynctask.NewAsyncTask("SubmitCounters", record, period, nil, logger)
}

// NewRecordGaugesTask creates a new splits fetching and storing task
func NewRecordGaugesTask(
	metricsStorage storage.MetricsStorage,
	metricsRecorder service.MetricsRecorder,
	period int,
	sdkVersion,
	machineIP string,
	machineName string,
	logger logging.LoggerInterface,
) *asynctask.AsyncTask {
	record := func(logger logging.LoggerInterface) error {
		return submitGauges(
			metricsStorage,
			metricsRecorder,
			sdkVersion,
			machineIP,
			machineName,
		)
	}
	return asynctask.NewAsyncTask("SubmitGauges", record, period, nil, logger)
}

// NewRecordLatenciesTask creates a new splits fetching and storing task
func NewRecordLatenciesTask(
	metricsStorage storage.MetricsStorage,
	metricsRecorder service.MetricsRecorder,
	period int,
	sdkVersion,
	machineIP string,
	machineName string,
	logger logging.LoggerInterface,
) *asynctask.AsyncTask {
	record := func(logger logging.LoggerInterface) error {
		return submitLatencies(
			metricsStorage,
			metricsRecorder,
			sdkVersion,
			machineIP,
			machineName,
		)
	}
	return asynctask.NewAsyncTask("SubmitLatencies", record, period, nil, logger)
}
