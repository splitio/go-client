package storage

import (
	"github.com/splitio/go-client/splitio/dto"
)

const (
	treatment            = "getTreatment"
	treatments           = "getTreatments"
	treatmentWithConfig  = "getTreatmentWithConfig"
	treatmentsWithConfig = "getTreatmentsWithConfig"
	track                = "track"

	splitSync      = "split"
	segmentSync    = "segment"
	impressionSync = "impression"
	eventSync      = "event"
	telemetrySync  = "telemetry"
	tokenSync      = "token"
)

const (
	maxStreamingEvents = 20
	maxTags            = 10
	latencyBucketCount = 23
	maxIntegrations    = 10
)

// TelemetryStorage interface
type TelemetryStorage interface {
	TelemetryStorageConsumer
	TelemetryStorageProducer
}

// TelemetryStorageConsumer consumer interface
type TelemetryStorageConsumer interface {
	PopLatencies() dto.MethodLatencies
	PopExceptions() dto.MethodExceptions
	GetDroppedImpressions() int64
	GetDedupedImpressions() int64
	GetQueuedmpressions() int64
	GetDroppedEvents() int64
	GetQueuedEvents() int64
	GetLastSynchronization() dto.LastSynchronization
	PopHTTPErrors() dto.HTTPErrors
	PopHTTPLatencies() dto.HTTPLatencies
	PopAuthRejections() int64
	PopTokenRefreshes() int64
	PopStreamingEvents() []dto.StreamingEvent
	PopTags() []string
	GetSessionLength() int64
	GetActiveFactories() int64
	GetRedundantActiveFactories() int64
	GetNonReadyUsages() int64
	GetBURTimeouts() int64
	GetTimeUntilReady() int64
	GetIntegrations() []string
}

// TelemetryStorageProducer producer interface
type TelemetryStorageProducer interface {
	RecordLatency(method string, bucket int)
	RecordException(method string)
	RecordDroppedImpressions(count int64)
	RecordDedupedImpressions(count int64)
	RecordQueuedImpressions(count int64)
	RecordDroppedEvents(count int64)
	RecordQueuedEvents(count int64)
	RecordSuccessfulSplitSync(timestamp int64)
	RecordSuccessfulSegmentSync(timestamp int64)
	RecordSuccessfulImpressionSync(timestamp int64)
	RecordSuccessfulEventsSync(timestamp int64)
	RecordSuccessfulTelemetrySync(timestamp int64)
	RecordSuccessfulTokenGet(timestamp int64)
	RecordSyncError(path string, status int)
	RecordSyncLatency(path string, bucket int)
	RecordAuthRejections()
	RecordTokenRefreshes()
	RecordStreamingEvent(streamingEvent dto.StreamingEvent)
	AddTag(tag string)
	RecordSessionLength(session int64)
	RecordFactory(apikey string)
	RecordNonReadyUsage()
	RecordBURTimeout()
	RecordTimeUntilReady(time int64)
	AddIntegration(integration string)
}
