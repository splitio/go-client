package telemetry

import (
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/dto"
)

// Manager interface for building regular data
type Manager interface {
	BuildInitData(cfg *conf.SplitSdkConfig, timedUntilReady int64, factoryInstances map[string]int64) dto.InitData
	BuildStatsData() dto.StatsData
}

// Facade adapter
type Facade interface {
	FacadeConsumer
	FacadeProducer
}

// FacadeProducer producer
type FacadeProducer interface {
	FactoryTelemetryProducer
	EvaluationTelemetryProducer
	ImpressionTelemetryProducer
	EventTelemetryProducer
	SynchronizationTelemetryProducer
	HTTPTelemetryProducer
	PushTelemetryProducer
	StreamingTelemetryProducer
	MiscTelemetryProducer
	SDKInfoTelemetryProducer
}

// FacadeConsumer adapter
type FacadeConsumer interface {
	FactoryTelemetryConsumer
	EvaluationTelemetryConsumer
	ImpressionTelemetryConsumer
	EventTelemetryConsumer
	SynchronizationTelemetryConsumer
	HTTPTelemetryConsumer
	CacheTelemetryConsumer
	PushTelemetryConsumer
	StreamingTelemetryConsumer
	MiscTelemetryConsumer
	SDKInfoTelemetryConsumer
}

// EvaluationTelemetry as used by the client
type EvaluationTelemetry interface { // Client
	EvaluationTelemetryConsumer
	EvaluationTelemetryProducer
}

// EvaluationTelemetryConsumer reader
type EvaluationTelemetryConsumer interface { // Client
	PopLatencies() dto.MethodLatencies
	PopExceptions() dto.MethodExceptions
}

// EvaluationTelemetryProducer writer
type EvaluationTelemetryProducer interface { // Client
	RecordLatency(method int, latency int64)
	RecordException(method int)
}

// ImpressionTelemetry includes the subset of telemetry operations triggered from the impressions manager
type ImpressionTelemetry interface { // ImpressionManager
	ImpressionTelemetryConsumer
	ImpressionTelemetryProducer
}

// ImpressionTelemetryConsumer reader
type ImpressionTelemetryConsumer interface {
	GetImpressionsStats(dataType int) int64
}

// ImpressionTelemetryProducer writer
type ImpressionTelemetryProducer interface { // ImpressionManager
	RecordImpressionsStats(dataType int, count int64)
}

// EventTelemetry includes the subset of telemetry operations
type EventTelemetry interface {
	EventTelemetryConsumer
	EventTelemetryProducer
}

// EventTelemetryConsumer reader
type EventTelemetryConsumer interface {
	GetEventsStats(dataType int) int64
}

// EventTelemetryProducer writer
type EventTelemetryProducer interface {
	RecordEventsStats(dataType int, count int64)
}

// SynchronizationTelemetry is referenced by the synchronizer to record
type SynchronizationTelemetry interface { // Individual Synchronizers/Fetcher/Recorders
	SynchronizationTelemetryConsumer
	SynchronizationTelemetryProducer
}

// SynchronizationTelemetryConsumer reader
type SynchronizationTelemetryConsumer interface {
	GetLastSynchronization() dto.LastSynchronization
}

// SynchronizationTelemetryProducer writer
type SynchronizationTelemetryProducer interface {
	RecordSuccessfulSync(resource int)
}

// HTTPTelemetry is the interface used by all HTTP-related classes to track request's outcome
type HTTPTelemetry interface { // Synchronizer
	HTTPTelemetryConsumer
	HTTPTelemetryProducer
}

// HTTPTelemetryConsumer reader
type HTTPTelemetryConsumer interface {
	PopHTTPErrors() dto.HTTPErrors
	PopHTTPLatencies() dto.HTTPLatencies
}

// HTTPTelemetryProducer writer
type HTTPTelemetryProducer interface {
	RecordSyncError(resource int, status int)
	RecordSyncLatency(resource int, latency int64)
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
	PopStreamingEvents() []dto.StreamingEvent
}

// StreamingTelemetryProducer writer
type StreamingTelemetryProducer interface {
	RecordStreamingEvent(eventType int, data int64)
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
	GetSessionLength() int64
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
	GetNonReadyUsages() int64
	GetBURTimeouts() int64
}

// FactoryTelemetryProducer writer
type FactoryTelemetryProducer interface {
	RecordNonReadyUsage()
	RecordBURTimeout()
}
