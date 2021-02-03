package telemetry

import "sync/atomic"

// ImpressionTelemetryFacade keeps track of impression-related metrics
type ImpressionTelemetryFacade struct {
	impressionsQueued  int64
	impressionsDropped int64
	impressionsDeduped int64
}

// NewImpressionTelemetryFacade facade for ImpressionTelemetry
func NewImpressionTelemetryFacade() ImpressionTelemetry {
	return &ImpressionTelemetryFacade{
		impressionsQueued:  0,
		impressionsDropped: 0,
		impressionsDeduped: 0,
	}
}

// RecordDroppedImpressions increments dropped impressions
func (i *ImpressionTelemetryFacade) RecordDroppedImpressions(count int64) {
	atomic.AddInt64(&i.impressionsDropped, atomic.LoadInt64(&i.impressionsDropped)+count)
}

// RecordDedupedImpressions increments dedupped impressions
func (i *ImpressionTelemetryFacade) RecordDedupedImpressions(count int64) {
	atomic.AddInt64(&i.impressionsDeduped, atomic.LoadInt64(&i.impressionsDeduped)+count)
}

// RecordQueuedImpressions increments queued impressions
func (i *ImpressionTelemetryFacade) RecordQueuedImpressions(count int64) {
	atomic.AddInt64(&i.impressionsQueued, atomic.LoadInt64(&i.impressionsQueued)+count)
}

// GetDroppedImpressions returns dropped impressions
func (i *ImpressionTelemetryFacade) GetDroppedImpressions() int64 {
	return atomic.SwapInt64(&i.impressionsDropped, 0)
}

// GetDedupedImpressions returns deduped impressions
func (i *ImpressionTelemetryFacade) GetDedupedImpressions() int64 {
	return atomic.SwapInt64(&i.impressionsDeduped, 0)
}

// GetQueuedmpressions returns queued impressions
func (i *ImpressionTelemetryFacade) GetQueuedmpressions() int64 {
	return atomic.SwapInt64(&i.impressionsQueued, 0)
}
