package tasks

import (
	"errors"
	"fmt"
	"github.com/splitio/go-client/splitio/service"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/tasks/workerpool"
	"github.com/splitio/go-client/splitio/util/logging"
	"github.com/splitio/go-toolkit/datastructures/set"
)

// SegmentWorker struct contains resources and functions for fetching segments and storing them
type SegmentWorker struct {
	name           string
	failureTime    int64
	segmentStorage storage.SegmentStorage
	segmentFetcher service.SegmentFetcher
}

// Name Returns the name of the worker
func (w *SegmentWorker) Name() string {
	return w.name
}

// FailureTime Returns how much time should be waited after an error, before the worker resumes execution
func (w *SegmentWorker) FailureTime() int64 {
	return w.failureTime
}

// DoWork performs the actual work and returns an error if something goes wrong
func (w *SegmentWorker) DoWork(msg interface{}) error {
	segmentName, ok := msg.(string)
	if !ok {
		return errors.New("segment name popped from queue is not a string")
	}

	till := w.segmentStorage.Till(segmentName)
	segmentChanges, err := w.segmentFetcher.Fetch(segmentName, till)
	if err != nil {
		return err
	}

	oldSegment := w.segmentStorage.Get(segmentName)
	if oldSegment == nil {
		s := set.NewSet()
		for _, key := range segmentChanges.Added {
			s.Add(key)
		}
		w.segmentStorage.Put(segmentName, s, segmentChanges.Till)
	} else {
		// Segment exists, must add new members and remove old ones
		for _, key := range segmentChanges.Added {
			oldSegment.Add(key)
		}
		for _, key := range segmentChanges.Removed {
			oldSegment.Remove(key)
		}
		w.segmentStorage.Put(segmentName, oldSegment, segmentChanges.Till)
	}

	return nil
}

// OnError callback does nothing
func (w *SegmentWorker) OnError(e error) {}

// Cleanup callback does nothing
func (w *SegmentWorker) Cleanup() error { return nil }

func updateSegments(splitStorage storage.SplitStorage, admin *workerpool.WorkerAdmin) error {
	for _, name := range splitStorage.SegmentNames() {
		admin.QueueMessage(name)
	}
	return nil
}

// NewFetchSegmentsTask creates a new segment fetching and storing task
func NewFetchSegmentsTask(
	splitStorage storage.SplitStorage,
	segmentStorage storage.SegmentStorage,
	segmentFetcher service.SegmentFetcher,
	period int64,
	workerCount int,
	logger logging.LoggerInterface,
) *AsyncTask {
	admin := workerpool.NewWorkerAdmin(10, logger)
	for i := 0; i < workerCount; i++ {
		admin.AddWorker(&SegmentWorker{
			name:           fmt.Sprintf("SegmentWorker_%d", i),
			failureTime:    0,
			segmentFetcher: segmentFetcher,
			segmentStorage: segmentStorage,
		})
	}

	update := func(logger logging.LoggerInterface) error {
		return updateSegments(splitStorage, admin)
	}

	cleanup := func(logger logging.LoggerInterface) {
		admin.StopAll()
	}

	return NewAsyncTask("UpdateSegments", update, period, cleanup)
}
