package tasks

import (
	"github.com/splitio/go-client/splitio/service"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"
)

func updateSplits(splitStorage storage.SplitStorage, splitFetcher service.SplitFetcher) (bool, error) {
	till := splitStorage.Till()
	if till == 0 {
		till = -1
	}

	splits, err := splitFetcher.Fetch(till)
	if err != nil {
		return false, err
	}
	splitStorage.PutMany(splits.Splits, splits.Till)

	if splits.Since == splits.Till {
		return true, nil
	}
	return false, nil
}

// NewFetchSplitsTask creates a new splits fetching and storing task
func NewFetchSplitsTask(
	splitStorage storage.SplitStorage,
	splitFetcher service.SplitFetcher,
	period int,
	logger logging.LoggerInterface,
	readyChannel chan string,
) *asynctask.AsyncTask {
	init := func(logger logging.LoggerInterface) error {
		ready := false
		var err error
		for !ready {
			ready, err = updateSplits(splitStorage, splitFetcher)
			if err != nil {
				return err
			}
		}
		readyChannel <- "SPLITS_READY"
		return nil
	}

	update := func(logger logging.LoggerInterface) error {
		_, err := updateSplits(splitStorage, splitFetcher)
		return err
	}

	return asynctask.NewAsyncTask("UpdateSplits", update, period, init, nil, logger)
}
