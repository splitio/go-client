package telemetry

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
	i.impressionsDropped += count
}

// RecordDedupedImpressions increments dedupped impressions
func (i *ImpressionTelemetryFacade) RecordDedupedImpressions(count int64) {
	i.impressionsDeduped += count
}

// RecordQueuedImpressions increments queued impressions
func (i *ImpressionTelemetryFacade) RecordQueuedImpressions(count int64) {
	i.impressionsQueued += count
}

// GetDroppedImpressions returns dropped impressions
func (i *ImpressionTelemetryFacade) GetDroppedImpressions() int64 {
	return i.impressionsDropped
}

// GetDedupedImpressions returns deduped impressions
func (i *ImpressionTelemetryFacade) GetDedupedImpressions() int64 {
	return i.impressionsDeduped
}

// GetQueuedmpressions returns queued impressions
func (i *ImpressionTelemetryFacade) GetQueuedmpressions() int64 {
	return i.impressionsQueued
}
