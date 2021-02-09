package telemetry

import "github.com/splitio/go-client/splitio/conf"

// TelemetryStorage interface
type TelemetryStorage interface {
	TelemetryStorageConsumer
	TelemetryStorageProducer
}

// TelemetryStorageConsumer consumer interface
type TelemetryStorageConsumer interface {
	GetFactories() map[string]int64
	GetRecord(name string) int64
	PopCounter(name string) int64
	PopHTTPErrors() HTTPErrors
	PopItems(name string) interface{}
	PopLatency(name string) []int64
}

// TelemetryStorageProducer producer interface
type TelemetryStorageProducer interface {
	AddLatency(name string, bucket int)
	Increment(name string, value int64)
	PushItem(name string, item interface{})
	Set(name string, value int64)
}

// TelemetryManager interface for building regular data
type TelemetryManager interface {
	BuildConfigData(cfg *conf.SplitSdkConfig) ConfigMetrics
	BuildUsageData() RegularMetrics
}

// Telemetry adapter
type Telemetry interface {
	PopLatencies() MethodLatencies
	PopExceptions() MethodExceptions
	RecordLatency(method string, latency int64)
	RecordException(method string)

	GetDroppedImpressions() int64
	GetDedupedImpressions() int64
	GetQueuedmpressions() int64
	RecordDroppedImpressions(count int64)
	RecordDedupedImpressions(count int64)
	RecordQueuedImpressions(count int64)

	GetDroppedEvents() int64
	GetQueuedEvents() int64
	RecordDroppedEvents(count int64)
	RecordQueuedEvents(count int64)

	GetLastSynchronization() LastSynchronization
	RecordSuccessfulSplitSync()
	RecordSuccessfulSegmentSync()
	RecordSuccessfulImpressionSync()
	RecordSuccessfulEventsSync()
	RecordSuccessfulTelemetrySync()
	RecordSuccessfulTokenGet()

	PopHTTPErrors() HTTPErrors
	PopHTTPLatencies() HTTPLatencies
	RecordSyncError(path string, status int)
	RecordSyncLatency(path string, latency int64)

	GetSplitsCount() int64
	GetSegmentsCount() int64
	GetSegmentKeysCount() int64

	PopAuthRejections() int64
	PopTokenRefreshes() int64
	RecordAuthRejections()
	RecordTokenRefreshes()

	PopStreamingEvents() []StreamingEvent
	RecordPrimaryOccupancyChange(newPublisherCount int64)   // NotificationManagerKeeper / PushStatusKeeper
	RecordSecondaryOccupancyChange(newPublisherCount int64) // NotificationManagerKeeper / PushStatusKeeper
	RecordConnectionSuccess()                               // PushStatusHandler (SyncManager)
	RecordStreamingServiceStatus(newStatus int)             // NotificationManagerKeeper / PushStatusKeeper
	RecordTokenRefresh(tokenExpirationUtcTs int64)          // Authenticator / AuthApiClient
	RecordAblyError(statusCode int64)                       // NotificationManagerKeeper / PushStatusKeeper
	RecordConnectionClose(wasRequested bool)                // SSEClient
	RecordSyncModeUpdate(newSyncMode int64)

	PopTags() []string
	AddTag(tag string)

	GetSessionLength() int64
	RecordSessionLength(session int64)

	GetActiveFactories() int64
	GetRedundantActiveFactories() int64
	GetNonReadyUsages() int64
	GetBURTimeouts() int64
	GetTimeUntilReady() int64
	RecordFactory(apikey string)
	RecordNonReadyUsage()
	RecordBURTimeout()
	RecordTimeUntilReady(time int64)
}

// EvaluationTelemetry as used by the client
type EvaluationTelemetry interface { // Client
	EvaluationTelemetryConsumer
	EvaluationTelemetryProducer
}

// EvaluationTelemetryConsumer reader
type EvaluationTelemetryConsumer interface { // Client
	PopLatencies() MethodLatencies
	PopExceptions() MethodExceptions
}

// EvaluationTelemetryProducer writer
type EvaluationTelemetryProducer interface { // Client
	RecordLatency(method string, latency int64)
	RecordException(method string)
}

// ImpressionTelemetry includes the subset of telemetry operations triggered from the impressions manager
type ImpressionTelemetry interface { // ImpressionManager
	ImpressionTelemetryConsumer
	ImpressionTelemetryProducer
}

// ImpressionTelemetryConsumer reader
type ImpressionTelemetryConsumer interface {
	GetDroppedImpressions() int64
	GetDedupedImpressions() int64
	GetQueuedmpressions() int64
}

// ImpressionTelemetryProducer writer
type ImpressionTelemetryProducer interface { // ImpressionManager
	RecordDroppedImpressions(count int64)
	RecordDedupedImpressions(count int64)
	RecordQueuedImpressions(count int64)
}

// EventTelemetry includes the subset of telemetry operations
type EventTelemetry interface {
	EventTelemetryConsumer
	EventTelemetryProducer
}

// EventTelemetryConsumer reader
type EventTelemetryConsumer interface {
	GetDroppedEvents() int64
	GetQueuedEvents() int64
}

// EventTelemetryProducer writer
type EventTelemetryProducer interface {
	RecordDroppedEvents(count int64)
	RecordQueuedEvents(count int64)
}

// SynchronizationTelemetry is referenced by the synchronizer to record
type SynchronizationTelemetry interface { // Individual Synchronizers/Fetcher/Recorders
	SynchronizationTelemetryConsumer
	SynchronizationTelemetryProducer
}

// SynchronizationTelemetryConsumer reader
type SynchronizationTelemetryConsumer interface {
	GetLastSynchronization() LastSynchronization
}

// SynchronizationTelemetryProducer writer
type SynchronizationTelemetryProducer interface {
	RecordSuccessfulSplitSync()
	RecordSuccessfulSegmentSync()
	RecordSuccessfulImpressionSync()
	RecordSuccessfulEventsSync()
	RecordSuccessfulTelemetrySync()
	RecordSuccessfulTokenGet()
}

// HTTPTelemetry is the interface used by all HTTP-related classes to track request's outcome
type HTTPTelemetry interface { // Synchronizer
	HTTPTelemetryConsumer
	HTTPTelemetryProducer
}

// HTTPTelemetryConsumer reader
type HTTPTelemetryConsumer interface {
	PopHTTPErrors() HTTPErrors
	PopHTTPLatencies() HTTPLatencies
}

// HTTPTelemetryProducer writer
type HTTPTelemetryProducer interface {
	RecordSyncError(path string, status int)
	RecordSyncLatency(path string, latency int64)
}

// CacheTelemetry is the interface for cached data
type CacheTelemetry interface {
	CacheTelemetryConsumer
}

// CacheTelemetryConsumer reader
type CacheTelemetryConsumer interface {
	GetSplitsCount() int64
	GetSegmentsCount() int64
	GetSegmentKeysCount() int64
}

// PushTelemetry is the interface for push
type PushTelemetry interface {
	PushTelemetryConsumer
	PushTelemetryProducer
}

// PushTelemetryConsumer reader
type PushTelemetryConsumer interface {
	PopAuthRejections() int64
	PopTokenRefreshes() int64
}

// PushTelemetryProducer writer
type PushTelemetryProducer interface {
	RecordAuthRejections()
	RecordTokenRefreshes()
}

// StreamingTelemetry is referenced by several components of the streaming subsystem
type StreamingTelemetry interface {
	StreamingTelemetryConsumer
	StreamingTelemetryProducer
}

// StreamingTelemetryConsumer reader
type StreamingTelemetryConsumer interface {
	PopStreamingEvents() []StreamingEvent
}

// StreamingTelemetryProducer writer
type StreamingTelemetryProducer interface {
	RecordPrimaryOccupancyChange(newPublisherCount int64)   // NotificationManagerKeeper / PushStatusKeeper
	RecordSecondaryOccupancyChange(newPublisherCount int64) // NotificationManagerKeeper / PushStatusKeeper
	RecordConnectionSuccess()                               // PushStatusHandler (SyncManager)
	RecordStreamingServiceStatus(newStatus int)             // NotificationManagerKeeper / PushStatusKeeper
	RecordTokenRefresh(tokenExpirationUtcTs int64)          // Authenticator / AuthApiClient
	RecordAblyError(statusCode int64)                       // NotificationManagerKeeper / PushStatusKeeper
	RecordConnectionClose(wasRequested bool)                // SSEClient
	RecordSyncModeUpdate(newSyncMode int64)                 // NotificationManagerKeeper / PushStatusKeeper
}

// MiscTelemetry interface por misc data
type MiscTelemetry interface {
	MiscTelemetryConsumer
	MiscTelemetryProducer
}

// MiscTelemetryConsumer reader
type MiscTelemetryConsumer interface {
	PopTags() []string
}

// MiscTelemetryProducer writer
type MiscTelemetryProducer interface {
	AddTag(tag string)
}

// SDKInfoTelemetry interface for sdk info metrics
type SDKInfoTelemetry interface {
	SDKInfoTelemetryConsumer
	SDKInfoTelemetryProducer
}

// SDKInfoTelemetryConsumer reader
type SDKInfoTelemetryConsumer interface {
	PopSessionLength() int64
}

// SDKInfoTelemetryProducer writer
type SDKInfoTelemetryProducer interface {
	RecordSessionLength(session int64)
}

// FactoryTelemetry interface for factory metrics
type FactoryTelemetry interface {
	FactoryTelemetryConsumer
	FactoryTelemetryProducer
}

// FactoryTelemetryConsumer reader
type FactoryTelemetryConsumer interface {
	GetActiveFactories() int64
	GetRedundantActiveFactories() int64
	GetNonReadyUsages() int64
	GetBURTimeouts() int64
	GetTimeUntilReady() int64
}

// FactoryTelemetryProducer writer
type FactoryTelemetryProducer interface {
	RecordFactory(apikey string)
	RecordNonReadyUsage()
	RecordBURTimeout()
	RecordTimeUntilReady(time int64)
}
