package telemetry

import (
	"github.com/splitio/go-split-commons/storage"
)

// CacheTelemetryFacade keeps track of cache-related metrics
type CacheTelemetryFacade struct {
	splitsStorage  storage.SplitStorageConsumer
	segmentStorage storage.SegmentStorageConsumer
}

// NewCacheTelemetryFacade builds new facade
func NewCacheTelemetryFacade(splitStorage storage.SplitStorageConsumer, segmentStorage storage.SegmentStorageConsumer) CacheTelemetry {
	return &CacheTelemetryFacade{
		splitsStorage:  splitStorage,
		segmentStorage: segmentStorage,
	}
}

// GetSplitsCount gets total splits
func (c *CacheTelemetryFacade) GetSplitsCount() int64 {
	return int64(len(c.splitsStorage.SplitNames()))
}

// GetSegmentsCount gets total segments
func (c *CacheTelemetryFacade) GetSegmentsCount() int64 {
	return int64(c.splitsStorage.SegmentNames().Size())
}

// GetSegmentKeysCount gets total segmentKeys
func (c *CacheTelemetryFacade) GetSegmentKeysCount() int64 {
	toReturn := 0
	segmentNames := c.splitsStorage.SegmentNames().List()
	for _, segmentName := range segmentNames {
		toReturn += c.segmentStorage.Keys(segmentName.(string)).Size()
	}
	return int64(toReturn)
}
