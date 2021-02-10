package constants

const (
	// Treatment getTreatment
	Treatment = iota
	// Treatments getTreatments
	Treatments
	// TreatmentWithConfig getTreatmentWithConfig
	TreatmentWithConfig
	// TreatmentsWithConfig getTreatmentsWithConfig
	TreatmentsWithConfig
	// Track track
	Track
)

const (
	// SplitSync splitChanges
	SplitSync = iota
	// SegmentSync segmentChanges
	SegmentSync
	// ImpressionSync impressions
	ImpressionSync
	// EventSync events
	EventSync
	// TelemetrySync telemetry
	TelemetrySync
	// TokenSync auth
	TokenSync
)

const (
	// ImpressionsDropped dropped
	ImpressionsDropped = iota
	// ImpressionsDeduped deduped
	ImpressionsDeduped
	// ImpressionsQueued queued
	ImpressionsQueued
)

const (
	// EventsDropped dropped
	EventsDropped = iota
	// EventsQueued queued
	EventsQueued
)

const (
	// LatencyBucketCount Max buckets
	LatencyBucketCount = 23
	// MaxStreamingEvents Max streaming events allowed
	MaxStreamingEvents = 20
	// MaxTags Max tags
	MaxTags = 10
)
