package telemetry

// TODO: Move this to microservices toolkit

// Method constants
const (
	MethodTreatment = iota
	MethodTreatments
	MethodTreatmentWithConfig
	MethodTreatmentsWithConfig
	MethodTrack

	TotalMethodCount   = 5
	LatencyBucketCount = 20
)

// EvaluationTelemetry as used by the client
type EvaluationTelemetry interface { // Client
	RecordLatency(method string, latency int64)
	RecordException(method string)
}

// ImpressionTelemetry includes the subset of telemetry operations triggered form the impressions manager
type ImpressionTelemetry interface { // ImpressionManager
	RecordDroppedImpressions(count int64)
	RecordDedupedImpressions(count int64)
}

// HTTPTelemetry is the interface used by all HTTP-related classes to track request's outcome
type HTTPTelemetry interface { // Synchronizer
	RecordRequestOutcome(endpoint string, status int, latency int)
}

// SynchronizationTelemetry is referenced by the synchronizer to record
type SynchronizationTelemetry interface { // Individual Synchronizers/Fetcher/Recorders
	RecordSuccessfulSplitSync()
	RecordSuccessfulSegmentSync()
	RecordSuccessfulImpressionSync()
	RecordSuccessfulEventsSync()
	RecordSuccessfulTelemetrySync()
}

// StreamingTelemetry is referenced by several components of the streaming subsystem
type StreamingTelemetry interface {
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
