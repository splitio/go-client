package telemetry

import (
	"time"

	"github.com/splitio/go-client/splitio/dto"
	"github.com/splitio/go-client/splitio/storage"
	commonStorage "github.com/splitio/go-split-commons/storage"
	"github.com/splitio/go-split-commons/util"
)

const (
	requested = iota
	nonRequested
)

const (
	eventTypeSSEConnectionEstablished = iota * 10
	eventTypeOccupancyPri
	eventTypeOccupancySec
	eventTypeStreamingStatus
	eventTypeConnectionError
	eventTypeTokenRefresh
	eventTypeAblyError
	eventTypeSyncMode
)

// FacadeImpl keeps track metrics
type FacadeImpl struct {
	storage        storage.TelemetryStorage
	splitsStorage  commonStorage.SplitStorageConsumer
	segmentStorage commonStorage.SegmentStorageConsumer
}

// NewTelemetry builds FacadeImpl
func NewTelemetry(telemetryStorage storage.TelemetryStorage, splitStorage commonStorage.SplitStorageConsumer, segmentStorage commonStorage.SegmentStorageConsumer) Facade {
	return &FacadeImpl{
		storage:        telemetryStorage,
		splitsStorage:  splitStorage,
		segmentStorage: segmentStorage,
	}
}

// EVALUATION

// RecordException records exceptions
func (t *FacadeImpl) RecordException(method string) {
	t.storage.RecordException(method)
}

// RecordLatency records latencies for method
func (t *FacadeImpl) RecordLatency(method string, latency int64) {
	t.storage.RecordLatency(method, util.Bucket(latency))
}

// PopLatencies returns latencies
func (t *FacadeImpl) PopLatencies() dto.MethodLatencies {
	return t.storage.PopLatencies()
}

// PopExceptions returns exceptions
func (t *FacadeImpl) PopExceptions() dto.MethodExceptions {
	return t.storage.PopExceptions()
}

// IMPRESSIONS

// RecordQueuedImpressions increments queued impressions
func (t *FacadeImpl) RecordQueuedImpressions(count int64) {
	t.storage.RecordQueuedImpressions(count)
}

// RecordDedupedImpressions increments dedupped impressions
func (t *FacadeImpl) RecordDedupedImpressions(count int64) {
	t.storage.RecordDedupedImpressions(count)
}

// RecordDroppedImpressions increments dropped impressions
func (t *FacadeImpl) RecordDroppedImpressions(count int64) {
	t.storage.RecordDroppedImpressions(count)
}

// GetDroppedImpressions returns dropped impressions
func (t *FacadeImpl) GetDroppedImpressions() int64 {
	return t.storage.GetDroppedImpressions()
}

// GetDedupedImpressions returns deduped impressions
func (t *FacadeImpl) GetDedupedImpressions() int64 {
	return t.storage.GetDedupedImpressions()
}

// GetQueuedmpressions returns queued impressions
func (t *FacadeImpl) GetQueuedmpressions() int64 {
	return t.storage.GetQueuedmpressions()
}

// EVENTS

// RecordDroppedEvents increments dropped events
func (t *FacadeImpl) RecordDroppedEvents(count int64) {
	t.storage.RecordDroppedEvents(count)
}

// RecordQueuedEvents increments queued events
func (t *FacadeImpl) RecordQueuedEvents(count int64) {
	t.storage.RecordQueuedEvents(count)
}

// GetDroppedEvents returns dropped events
func (t *FacadeImpl) GetDroppedEvents() int64 {
	return t.storage.GetDroppedEvents()
}

// GetQueuedEvents returns queued events
func (t *FacadeImpl) GetQueuedEvents() int64 {
	return t.storage.GetQueuedEvents()
}

// LAST SYNCHRONIZATION

// RecordSuccessfulSplitSync records split sync
func (t *FacadeImpl) RecordSuccessfulSplitSync() {
	t.storage.RecordSuccessfulSplitSync(time.Now().UTC().UnixNano() / 1000000)
}

// RecordSuccessfulSegmentSync records segment sync
func (t *FacadeImpl) RecordSuccessfulSegmentSync() {
	t.storage.RecordSuccessfulSegmentSync(time.Now().UTC().UnixNano() / 1000000)
}

// RecordSuccessfulImpressionSync records impression sync
func (t *FacadeImpl) RecordSuccessfulImpressionSync() {
	t.storage.RecordSuccessfulImpressionSync(time.Now().UTC().UnixNano() / 1000000)
}

// RecordSuccessfulEventsSync records event sync
func (t *FacadeImpl) RecordSuccessfulEventsSync() {
	t.storage.RecordSuccessfulEventsSync(time.Now().UTC().UnixNano() / 1000000)
}

// RecordSuccessfulTelemetrySync records telemetry sync
func (t *FacadeImpl) RecordSuccessfulTelemetrySync() {
	t.storage.RecordSuccessfulTelemetrySync(time.Now().UTC().UnixNano() / 1000000)
}

// RecordSuccessfulTokenGet records token sync
func (t *FacadeImpl) RecordSuccessfulTokenGet() {
	t.storage.RecordSuccessfulTokenGet(time.Now().UTC().UnixNano() / 1000000)
}

// GetLastSynchronization gets last sync records
func (t *FacadeImpl) GetLastSynchronization() dto.LastSynchronization {
	return t.storage.GetLastSynchronization()
}

// HTTP

// PopHTTPErrors returns errors stored
func (t *FacadeImpl) PopHTTPErrors() dto.HTTPErrors {
	return t.storage.PopHTTPErrors()
}

// PopHTTPLatencies returns latencies stored
func (t *FacadeImpl) PopHTTPLatencies() dto.HTTPLatencies {
	return t.storage.PopHTTPLatencies()
}

// RecordSyncError records error
func (t *FacadeImpl) RecordSyncError(path string, status int) {
	t.storage.RecordSyncError(path, status)
}

// RecordSyncLatency records latencies
func (t *FacadeImpl) RecordSyncLatency(path string, latency int64) {
	t.storage.RecordSyncLatency(path, util.Bucket(latency))
}

// CACHE

// GetSplitsCount gets total splits
func (t *FacadeImpl) GetSplitsCount() int64 {
	return int64(len(t.splitsStorage.SplitNames()))
}

// GetSegmentsCount gets total segments
func (t *FacadeImpl) GetSegmentsCount() int64 {
	return int64(t.splitsStorage.SegmentNames().Size())
}

// GetSegmentKeysCount gets total segmentKeys
func (t *FacadeImpl) GetSegmentKeysCount() int64 {
	toReturn := 0
	segmentNames := t.splitsStorage.SegmentNames().List()
	for _, segmentName := range segmentNames {
		toReturn += t.segmentStorage.Keys(segmentName.(string)).Size()
	}
	return int64(toReturn)
}

// PUSH

// RecordAuthRejections records auth rejection
func (t *FacadeImpl) RecordAuthRejections() {
	t.storage.RecordAuthRejections()
}

// RecordTokenRefreshes records token refresh
func (t *FacadeImpl) RecordTokenRefreshes() {
	t.storage.RecordTokenRefreshes()
}

// PopAuthRejections returns all the rejections
func (t *FacadeImpl) PopAuthRejections() int64 {
	return t.storage.PopAuthRejections()
}

// PopTokenRefreshes returns all the refreshes made
func (t *FacadeImpl) PopTokenRefreshes() int64 {
	return t.storage.PopTokenRefreshes()
}

// STREAMING

// RecordPrimaryOccupancyChange records occupancy on primary
func (t *FacadeImpl) RecordPrimaryOccupancyChange(newPublisherCount int64) {
	t.storage.RecordStreamingEvent(dto.StreamingEvent{
		Type:      eventTypeOccupancyPri,
		Data:      newPublisherCount,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordSecondaryOccupancyChange records occupancy on secondary
func (t *FacadeImpl) RecordSecondaryOccupancyChange(newPublisherCount int64) {
	t.storage.RecordStreamingEvent(dto.StreamingEvent{
		Type:      eventTypeOccupancySec,
		Data:      newPublisherCount,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordConnectionSuccess records success on SSE
func (t *FacadeImpl) RecordConnectionSuccess() {
	t.storage.RecordStreamingEvent(dto.StreamingEvent{
		Type:      eventTypeSSEConnectionEstablished,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordStreamingServiceStatus records new status on streaming
func (t *FacadeImpl) RecordStreamingServiceStatus(newStatus int) {
	t.storage.RecordStreamingEvent(dto.StreamingEvent{
		Type:      eventTypeStreamingStatus,
		Data:      int64(newStatus),
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordTokenRefresh records next expiration event
func (t *FacadeImpl) RecordTokenRefresh(tokenExpirationUtcTs int64) {
	t.storage.RecordStreamingEvent(dto.StreamingEvent{
		Type:      eventTypeTokenRefresh,
		Data:      tokenExpirationUtcTs,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordAblyError records erros in ably SSE
func (t *FacadeImpl) RecordAblyError(statusCode int64) {
	t.storage.RecordStreamingEvent(dto.StreamingEvent{
		Type:      eventTypeAblyError,
		Data:      statusCode,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordConnectionClose records connections close
func (t *FacadeImpl) RecordConnectionClose(wasRequested bool) {
	data := requested
	if !wasRequested {
		data = nonRequested
	}
	t.storage.RecordStreamingEvent(dto.StreamingEvent{
		Type:      eventTypeConnectionError,
		Data:      int64(data),
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordSyncModeUpdate records updates in streaming
func (t *FacadeImpl) RecordSyncModeUpdate(newSyncMode int64) {
	t.storage.RecordStreamingEvent(dto.StreamingEvent{
		Type:      eventTypeSyncMode,
		Data:      newSyncMode,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// PopStreamingEvents returns all the stored StreamingEvents
func (t *FacadeImpl) PopStreamingEvents() []dto.StreamingEvent {
	return t.storage.PopStreamingEvents()
}

// TAGS

// AddTag add tags to misc data
func (t *FacadeImpl) AddTag(tag string) {
	t.storage.AddTag(tag)
}

// PopTags returns all the tags stored
func (t *FacadeImpl) PopTags() []string {
	return t.storage.PopTags()
}

// SDK

// RecordSessionLength stores session duration
func (t *FacadeImpl) RecordSessionLength(session int64) {
	t.storage.RecordSessionLength(session)
}

// GetSessionLength returns stored session
func (t *FacadeImpl) GetSessionLength() int64 {
	return t.storage.GetSessionLength()
}

// FACTORY

// AddIntegration adds integration to factory data
func (t *FacadeImpl) AddIntegration(integration string) {
	t.storage.AddIntegration(integration)
}

// RecordFactory stores factory
func (t *FacadeImpl) RecordFactory(apikey string) {
	t.storage.RecordFactory(apikey)
}

// RecordNonReadyUsage stores non ready usage
func (t *FacadeImpl) RecordNonReadyUsage() {
	t.storage.RecordNonReadyUsage()
}

// RecordBURTimeout stores timeouts
func (t *FacadeImpl) RecordBURTimeout() {
	t.storage.RecordBURTimeout()
}

// RecordTimeUntilReady stores time duration
func (t *FacadeImpl) RecordTimeUntilReady(time int64) {
	t.storage.RecordTimeUntilReady(time)
}

// GetIntegrations returns all the integrations stored
func (t *FacadeImpl) GetIntegrations() []string {
	return t.storage.GetIntegrations()
}

// GetActiveFactories gets active factories
func (t *FacadeImpl) GetActiveFactories() int64 {
	return t.storage.GetActiveFactories()
}

// GetRedundantActiveFactories gets redundant factories
func (t *FacadeImpl) GetRedundantActiveFactories() int64 {
	return t.storage.GetRedundantActiveFactories()
}

// GetNonReadyUsages gets non ready usages
func (t *FacadeImpl) GetNonReadyUsages() int64 {
	return t.storage.GetNonReadyUsages()
}

// GetBURTimeouts gets bur timeots
func (t *FacadeImpl) GetBURTimeouts() int64 {
	return t.storage.GetBURTimeouts()
}

// GetTimeUntilReady returns stored session
func (t *FacadeImpl) GetTimeUntilReady() int64 {
	return t.storage.GetTimeUntilReady()
}
