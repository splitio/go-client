package tasks

import (
	"github.com/splitio/go-client/splitio/service"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"
)

func updateSplits(splitStorage storage.SplitStorageProducer, splitFetcher service.SplitFetcher, trafficTypeStorage storage.TrafficTypeStorageProducer) (bool, error) {
	till := splitStorage.Till()
	if till == 0 {
		till = -1
	}

	splits, err := splitFetcher.Fetch(till)
	if err != nil {
		return false, err
	}

	inactiveSplits := make([]dtos.SplitDTO, 0)
	activeSplits := make([]dtos.SplitDTO, 0)
	trafficTypesToAdd := make([]string, 0)
	trafficTypesToRemove := make([]string, 0)
	for _, split := range splits.Splits {
		if split.Status == "ACTIVE" {
			activeSplits = append(activeSplits, split)
			trafficTypesToAdd = append(trafficTypesToAdd, split.TrafficTypeName)
		} else {
			inactiveSplits = append(inactiveSplits, split)
			if till != -1 {
				trafficTypesToRemove = append(trafficTypesToRemove, split.TrafficTypeName)
			}
		}
	}

	// Add/Update active splits
	splitStorage.PutMany(activeSplits, splits.Till)

	// Remove inactive splits
	for _, split := range inactiveSplits {
		splitStorage.Remove(split.Name)
	}

	for _, trafficType := range trafficTypesToAdd {
		trafficTypeStorage.Increase(trafficType)
	}

	for _, trafficType := range trafficTypesToRemove {
		trafficTypeStorage.Decrease(trafficType)
	}

	if splits.Since == splits.Till {
		return true, nil
	}
	return false, nil
}

// NewFetchSplitsTask creates a new splits fetching and storing task
func NewFetchSplitsTask(
	splitStorage storage.SplitStorageProducer,
	splitFetcher service.SplitFetcher,
	period int,
	logger logging.LoggerInterface,
	readyChannel chan string,
	trafficTypeStorage storage.TrafficTypeStorageProducer,
) *asynctask.AsyncTask {
	init := func(logger logging.LoggerInterface) error {
		ready := false
		var err error
		for !ready {
			ready, err = updateSplits(splitStorage, splitFetcher, trafficTypeStorage)
			if err != nil {
				readyChannel <- "SPLITS_ERROR"
				return err
			}
		}
		readyChannel <- "SPLITS_READY"
		return nil
	}

	update := func(logger logging.LoggerInterface) error {
		_, err := updateSplits(splitStorage, splitFetcher, trafficTypeStorage)
		return err
	}

	return asynctask.NewAsyncTask("UpdateSplits", update, period, init, nil, logger)
}
