package telemetry

// EvaluationTelemetry as used by the client
type EvaluationTelemetry interface { // Client
	EvaluationTelemetryConsumer
	EvaluationTelemetryProducer
}

// EvaluationTelemetryConsumer reader
type EvaluationTelemetryConsumer interface { // Client
	GetLatencies() map[string][]int64
	GetExceptions() map[string]int64
}

// EvaluationTelemetryProducer writer
type EvaluationTelemetryProducer interface { // Client
	RecordLatency(method string, latency int64)
	RecordException(method string)
}

// ImpressionTelemetry includes the subset of telemetry operations triggered form the impressions manager
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

// HTTPErrorTelemetry is the interface used by all HTTP-related classes to track request's outcome
type HTTPErrorTelemetry interface { // Synchronizer
	HTTPErrorTelemetryConsumer
	HTTPErrorTelemetryProducer
}

// HTTPErrorTelemetryConsumer reader
type HTTPErrorTelemetryConsumer interface {
	GetHTTPErrors() HTTPErrors
}

// HTTPErrorTelemetryProducer writer
type HTTPErrorTelemetryProducer interface {
	RecordSplitSyncErr(status int)
	RecordSegmentSyncErr(status int)
	RecordImpressionSyncErr(status int)
	RecordEventSyncErr(status int)
	RecordTelemetrySyncErr(status int)
	RecordTokenGetErr(status int)
}

// CacheTelemetry is the interface for cached data
type CacheTelemetry interface {
	CacheTelemetryConsumer
	CacheTelemetryProducer
}

// CacheTelemetryConsumer reader
type CacheTelemetryConsumer interface {
	GetSplitsCount() int64
	GetSegmentCount() int64
	GetSegmentKeyCount() int64
}

// CacheTelemetryProducer writer
type CacheTelemetryProducer interface {
	RecordSplitCount() int64
	RecordSegmentCount() int64
	RecordSegmentKeyCount() int64
}

// PushTelemetry is the interface for push
type PushTelemetry interface {
	PushTelemetryConsumer
	PushTelemetryProducer
}

// PushTelemetryConsumer reader
type PushTelemetryConsumer interface {
	GetAuthRejections() int64
	GetTokenRefreshes() int64
}

// PushTelemetryProducer writer
type PushTelemetryProducer interface {
	RecordAuthRejections() int64
	RecordTokenRefreshes() int64
}

// StreamingTelemetry is referenced by several components of the streaming subsystem
type StreamingTelemetry interface {
	StreamingTelemetryConsumer
	StreamingTelemetryProducer
}

// StreamingTelemetryConsumer reader
type StreamingTelemetryConsumer interface {
	GetStreamingEvents() []StreamingEvent
}

// StreamingTelemetryProducer writer
type StreamingTelemetryProducer interface {
	RecordPrimaryOccupancyChange(newPublisherCount int64)   // NotificationManagerKeeper / PushStatusKeeper
	RecordSecondaryOccupancyChange(newPublisherCount int64) // NotificationManagerKeeper / PushStatusKeeper
	RecordConnectionSuccess()                               // PushStatusHandler (SyncManager)
	RecordStreamingServiceStatus(newStatus int)             // NotificationManagerKeeper / PushStatusKeeper
	RecordTokenRefresh(tokenExpirationUtcTs int64)          // Authenticator / AuthApiClient
	RecordAblyError(statusCode int64)                       // NotificationManagerKeeper / PushStatusKeeper
	RecordNonRequestedConnectionClose()                     // SSEClient
	RecordAuthRejection(statusCode int64)                   // NotificationManagerKeeper / PushStatusKeeper
	RecordSyncModeUpdate(newSyncMode int64)                 // NotificationManagerKeeper / PushStatusKeeper
}

// MiscTelemetry interface por misc data
type MiscTelemetry interface {
	MiscTelemetryConsumer
	MiscTelemetryProducer
}

// MiscTelemetryConsumer reader
type MiscTelemetryConsumer interface {
	GetTags() []string
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
	GetSessionLength() int64
}

// SDKInfoTelemetryProducer writer
type SDKInfoTelemetryProducer interface {
	RecordSessionLength(session int64)
}
