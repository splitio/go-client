package storage

import (
	"github.com/splitio/go-client/splitio/dto"
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
	GetImpressionsStats(dataType int) int64
	GetEventsStats(dataType int) int64
	GetLastSynchronization() dto.LastSynchronization
	PopHTTPErrors() dto.HTTPErrors
	PopHTTPLatencies() dto.HTTPLatencies
	PopAuthRejections() int64
	PopTokenRefreshes() int64
	PopStreamingEvents() []dto.StreamingEvent
	PopTags() []string
	GetSessionLength() int64
	GetNonReadyUsages() int64
	GetBURTimeouts() int64
}

// TelemetryStorageProducer producer interface
type TelemetryStorageProducer interface {
	RecordLatency(method int, bucket int)
	RecordException(method int)
	RecordImpressionsStats(dataType int, count int64)
	RecordEventsStats(dataType int, count int64)
	RecordSuccessfulSync(resource int, time int64)
	RecordSyncError(resource int, status int)
	RecordSyncLatency(resource int, bucket int)
	RecordAuthRejections()
	RecordTokenRefreshes()
	RecordStreamingEvent(streamingEvent dto.StreamingEvent)
	AddTag(tag string)
	RecordSessionLength(session int64)
	RecordNonReadyUsage()
	RecordBURTimeout()
}

// CustomTelemetryWrapper interface for pluggable storages
type CustomTelemetryWrapper interface {
	Increment(key string, value int64)
	Set(key string, value interface{})
	PushItem(key string, item interface{})

	PopItem(key string) interface{}
	GetItem(key string) interface{}
	PopItems(key string) interface{}

	GetByPrefix(prefix string) []interface{}
}
