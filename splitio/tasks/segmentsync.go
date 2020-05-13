package tasks

import (
	"errors"
	"fmt"
	"sync"

	"github.com/splitio/go-client/splitio/service"
	"github.com/splitio/go-split-commons/storage"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
	"github.com/splitio/go-toolkit/workerpool"
)

func updateSegment(
	segmentFetcher service.SegmentFetcher,
	segmentStorage storage.SegmentStorage,
	name string,
) (bool, error) {
	till := segmentStorage.Till(name)
	segmentChanges, err := segmentFetcher.Fetch(name, till)
	if err != nil {
		return false, err
	}

	oldSegment := segmentStorage.Get(name)
	if oldSegment == nil {
		s := set.NewSet()
		for _, key := range segmentChanges.Added {
			s.Add(key)
		}
		segmentStorage.Put(name, s, segmentChanges.Till)
	} else {
		// Segment exists, must add new members and remove old ones
		for _, key := range segmentChanges.Added {
			oldSegment.Add(key)
		}
		for _, key := range segmentChanges.Removed {
			oldSegment.Remove(key)
		}
		segmentStorage.Put(name, oldSegment, segmentChanges.Till)
	}

	return segmentChanges.Since == segmentChanges.Till, nil
}

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

	_, err := updateSegment(w.segmentFetcher, w.segmentStorage, segmentName)
	return err
}

// OnError callback does nothing
func (w *SegmentWorker) OnError(e error) {}

// Cleanup callback does nothing
func (w *SegmentWorker) Cleanup() error { return nil }

func updateSegments(
	splitStorage storage.SplitStorageConsumer,
	admin *workerpool.WorkerAdmin,
	logger logging.LoggerInterface,
) error {
	segmentList := splitStorage.SegmentNames().List()
	for _, name := range segmentList {
		ok := admin.QueueMessage(name)
		if !ok {
			logger.Error(
				fmt.Sprintf("Segment %s could not be added because the job queue is full.\n", name),
				fmt.Sprintf(
					"You currently have %d segments and the queue size is %d.\n",
					len(segmentList),
					admin.QueueSize(),
				),
				"Please consider updating the segment queue size accordingly in the configuration options",
			)
		}
	}
	return nil
}

// NewFetchSegmentsTask creates a new segment fetching and storing task
func NewFetchSegmentsTask(
	splitStorage storage.SplitStorageConsumer,
	segmentStorage storage.SegmentStorage,
	segmentFetcher service.SegmentFetcher,
	period int,
	workerCount int,
	queueSize int,
	logger logging.LoggerInterface,
	readyChannel chan string,
) *asynctask.AsyncTask {
	admin := workerpool.NewWorkerAdmin(queueSize, logger)

	init := func(logger logging.LoggerInterface) error {
		segmentNames := splitStorage.SegmentNames().List()
		wg := sync.WaitGroup{}
		wg.Add(len(segmentNames))
		failedSegments := make([]string, 0)
		for _, name := range segmentNames {
			conv, ok := name.(string)
			if !ok {
				logger.Warning("Skipping non-string segment present in storage at initialization-time!")
				continue
			}
			go func(segmentName string) {
				defer wg.Done() // Make sure the "finished" signal is always sent
				ready := false
				var err error
				for !ready {
					ready, err = updateSegment(segmentFetcher, segmentStorage, segmentName)
					if err != nil {
						failedSegments = append(failedSegments, segmentName)
						return
					}
				}
			}(conv)
		}
		wg.Wait()

		if len(failedSegments) > 0 {
			return fmt.Errorf("The following segments failed to be fetched %v", failedSegments)
		}

		// After all segments are in sync, add workers to the pool that will keep them up to date
		// periodically
		for i := 0; i < workerCount; i++ {
			admin.AddWorker(&SegmentWorker{
				name:           fmt.Sprintf("SegmentWorker_%d", i),
				failureTime:    0,
				segmentFetcher: segmentFetcher,
				segmentStorage: segmentStorage,
			})
		}

		readyChannel <- "SEGMENTS_READY"
		return nil
	}

	update := func(logger logging.LoggerInterface) error {
		return updateSegments(splitStorage, admin, logger)
	}

	cleanup := func(logger logging.LoggerInterface) {
		admin.StopAll()
	}

	return asynctask.NewAsyncTask("UpdateSegments", update, period, init, cleanup, logger)
}
