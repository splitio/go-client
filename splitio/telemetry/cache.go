package telemetry

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
	c.splits = count
}

// RecordSegmentsCount stores segments data
func (c *CacheTelemetryFacade) RecordSegmentsCount(count int64) {
	c.segments = count
}

// RecordSegmentKeysCount stores segmentKeys data
func (c *CacheTelemetryFacade) RecordSegmentKeysCount(count int64) {
	c.segmentKeys = count
}

// GetSplitsCount gets total splits
func (c *CacheTelemetryFacade) GetSplitsCount() int64 { return c.splits }

// GetSegmentCount gets total segments
func (c *CacheTelemetryFacade) GetSegmentCount() int64 { return c.segments }

// GetSegmentKeyCount gets total segmentKeys
func (c *CacheTelemetryFacade) GetSegmentKeyCount() int64 { return c.segmentKeys }
