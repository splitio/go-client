package tasks

import (
	"github.com/splitio/go-client/splitio/service"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/util/logging"
)

func updateSplits(splitStorage storage.SplitStorage, splitFetcher service.SplitFetcher) error {
	till := splitStorage.Till()
	if till == 0 {
		till = -1
	}

	splits, err := splitFetcher.Fetch(till)
	if err != nil {
		return err
	}

	splitStorage.PutMany(splits.Splits, splits.Till)
	return nil
}

// NewFetchSplitsTask creates a new splits fetching and storing task
func NewFetchSplitsTask(
	splitStorage storage.SplitStorage,
	splitFetcher service.SplitFetcher,
	period int64,
	logger logging.LoggerInterface,
) *AsyncTask {
	update := func(logger logging.LoggerInterface) error {
		return updateSplits(splitStorage, splitFetcher)
	}

	return NewAsyncTask("UpdateSplits", update, period, nil, logger)
}
