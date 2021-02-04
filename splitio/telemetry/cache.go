package telemetry

import "sync/atomic"

// CacheTelemetryFacade keeps track of cache-related metrics
type CacheTelemetryFacade struct {
	splits      int64
	segments    int64
	segmentKeys int64
}

// NewCacheTelemetryFacade builds new facade
func NewCacheTelemetryFacade() CacheTelemetry {
	return &CacheTelemetryFacade{
		splits:      0,
		segments:    0,
		segmentKeys: 0,
	}
}

// RecordSplitsCount stores splits data
func (c *CacheTelemetryFacade) RecordSplitsCount(count int64) {
	atomic.AddInt64(&c.splits, count)
}

// RecordSegmentsCount stores segments data
func (c *CacheTelemetryFacade) RecordSegmentsCount(count int64) {
	atomic.AddInt64(&c.segments, count)

}

// RecordSegmentKeysCount stores segmentKeys data
func (c *CacheTelemetryFacade) RecordSegmentKeysCount(count int64) {
	atomic.AddInt64(&c.segmentKeys, count)
}

// PopSplitsCount gets total splits
func (c *CacheTelemetryFacade) PopSplitsCount() int64 { return atomic.SwapInt64(&c.splits, 0) }

// PopSegmentCount gets total segments
func (c *CacheTelemetryFacade) PopSegmentCount() int64 { return atomic.SwapInt64(&c.segments, 0) }

// PopSegmentKeyCount gets total segmentKeys
func (c *CacheTelemetryFacade) PopSegmentKeyCount() int64 { return atomic.SwapInt64(&c.segmentKeys, 0) }
