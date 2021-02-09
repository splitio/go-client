package telemetry

import (
	"fmt"
	"time"

	"github.com/splitio/go-split-commons/storage"
	"github.com/splitio/go-split-commons/util"
)

// TelemetryImpl keeps track metrics
type TelemetryImpl struct {
	storage        TelemetryStorage
	splitsStorage  storage.SplitStorageConsumer
	segmentStorage storage.SegmentStorageConsumer
}

// NewTelemetry builds TelemetryImpl
func NewTelemetry(telemetryStorage TelemetryStorage, splitStorage storage.SplitStorageConsumer, segmentStorage storage.SegmentStorageConsumer) Telemetry {
	return &TelemetryImpl{
		storage:        telemetryStorage,
		splitsStorage:  splitStorage,
		segmentStorage: segmentStorage,
	}
}

// EVALUATION

// RecordException records exceptions
func (t *TelemetryImpl) RecordException(method string) {
	t.storage.Increment(method, 1)
}

// RecordLatency records latencies for method
func (t *TelemetryImpl) RecordLatency(method string, latency int64) {
	t.storage.AddLatency(method, util.Bucket(latency))
}

// PopLatencies returns latencies
func (t *TelemetryImpl) PopLatencies() MethodLatencies {
	return MethodLatencies{
		Treatment:            t.storage.PopLatency(treatment),
		Treatments:           t.storage.PopLatency(treatments),
		TreatmentWithConfig:  t.storage.PopLatency(treatmentWithConfig),
		TreatmentWithConfigs: t.storage.PopLatency(treatmentsWithConfig),
		Track:                t.storage.PopLatency(track),
	}
}

// PopExceptions returns exceptions
func (t *TelemetryImpl) PopExceptions() MethodExceptions {
	return MethodExceptions{
		Treatment:            t.storage.PopCounter(treatment),
		Treatments:           t.storage.PopCounter(treatments),
		TreatmentWithConfig:  t.storage.PopCounter(treatmentWithConfig),
		TreatmentWithConfigs: t.storage.PopCounter(treatmentsWithConfig),
		Track:                t.storage.PopCounter(track),
	}
}

// IMPRESSIONS

// RecordQueuedImpressions increments queued impressions
func (t *TelemetryImpl) RecordQueuedImpressions(count int64) {
	t.storage.Increment(impressionsQueued, count)
}

// RecordDedupedImpressions increments dedupped impressions
func (t *TelemetryImpl) RecordDedupedImpressions(count int64) {
	t.storage.Increment(impressionsDeduped, count)
}

// RecordDroppedImpressions increments dropped impressions
func (t *TelemetryImpl) RecordDroppedImpressions(count int64) {
	t.storage.Increment(impressionsDropped, count)
}

// GetDroppedImpressions returns dropped impressions
func (t *TelemetryImpl) GetDroppedImpressions() int64 {
	return t.storage.GetRecord(impressionsDropped)
}

// GetDedupedImpressions returns deduped impressions
func (t *TelemetryImpl) GetDedupedImpressions() int64 {
	return t.storage.GetRecord(impressionsDeduped)
}

// GetQueuedmpressions returns queued impressions
func (t *TelemetryImpl) GetQueuedmpressions() int64 {
	return t.storage.GetRecord(impressionsQueued)
}

// EVENTS

// RecordDroppedEvents increments dropped events
func (t *TelemetryImpl) RecordDroppedEvents(count int64) {
	t.storage.Increment(eventsDropped, count)
}

// RecordQueuedEvents increments queued events
func (t *TelemetryImpl) RecordQueuedEvents(count int64) {
	t.storage.Increment(eventsQueued, count)
}

// GetDroppedEvents returns dropped events
func (t *TelemetryImpl) GetDroppedEvents() int64 {
	return t.storage.GetRecord(eventsDropped)
}

// GetQueuedEvents returns queued events
func (t *TelemetryImpl) GetQueuedEvents() int64 {
	return t.storage.GetRecord(eventsQueued)
}

// LAST SYNCHRONIZATION

// RecordSuccessfulSplitSync records split sync
func (t *TelemetryImpl) RecordSuccessfulSplitSync() {
	t.storage.Set(split, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulSegmentSync records segment sync
func (t *TelemetryImpl) RecordSuccessfulSegmentSync() {
	t.storage.Set(segment, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulImpressionSync records impression sync
func (t *TelemetryImpl) RecordSuccessfulImpressionSync() {
	t.storage.Set(impression, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulEventsSync records event sync
func (t *TelemetryImpl) RecordSuccessfulEventsSync() {
	t.storage.Set(event, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulTelemetrySync records telemetry sync
func (t *TelemetryImpl) RecordSuccessfulTelemetrySync() {
	t.storage.Set(telemetry, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulTokenGet records token sync
func (t *TelemetryImpl) RecordSuccessfulTokenGet() {
	t.storage.Set(token, time.Now().UTC().UnixNano()/1000000)
}

// GetLastSynchronization gets last sync records
func (t *TelemetryImpl) GetLastSynchronization() LastSynchronization {
	return LastSynchronization{
		Splits:      t.storage.GetRecord(split),
		Segments:    t.storage.GetRecord(segment),
		Impressions: t.storage.GetRecord(impression),
		Events:      t.storage.GetRecord(event),
		Telemetry:   t.storage.GetRecord(telemetry),
		Token:       t.storage.GetRecord(token),
	}
}

// HTTP

// PopHTTPErrors returns errors stored
func (t *TelemetryImpl) PopHTTPErrors() HTTPErrors {
	return t.storage.PopHTTPErrors()
}

// PopHTTPLatencies returns latencies stored
func (t *TelemetryImpl) PopHTTPLatencies() HTTPLatencies {
	return HTTPLatencies{
		Splits:      t.storage.PopLatency(split),
		Segments:    t.storage.PopLatency(segment),
		Impressions: t.storage.PopLatency(impression),
		Events:      t.storage.PopLatency(event),
		Token:       t.storage.PopLatency(token),
		Telemetry:   t.storage.PopLatency(telemetry),
	}
}

// RecordSyncError records error
func (t *TelemetryImpl) RecordSyncError(path string, status int) {
	t.storage.Increment(fmt.Sprintf("%s::%d", path, status), 1)
}

// RecordSyncLatency records latencies
func (t *TelemetryImpl) RecordSyncLatency(path string, latency int64) {
	bucket := util.Bucket(latency)
	t.storage.AddLatency(path, bucket)
}

// CACHE

// GetSplitsCount gets total splits
func (t *TelemetryImpl) GetSplitsCount() int64 {
	return int64(len(t.splitsStorage.SplitNames()))
}

// GetSegmentsCount gets total segments
func (t *TelemetryImpl) GetSegmentsCount() int64 {
	return int64(t.splitsStorage.SegmentNames().Size())
}

// GetSegmentKeysCount gets total segmentKeys
func (t *TelemetryImpl) GetSegmentKeysCount() int64 {
	toReturn := 0
	segmentNames := t.splitsStorage.SegmentNames().List()
	for _, segmentName := range segmentNames {
		toReturn += t.segmentStorage.Keys(segmentName.(string)).Size()
	}
	return int64(toReturn)
}

// PUSH

// RecordAuthRejections records auth rejection
func (t *TelemetryImpl) RecordAuthRejections() {
	t.storage.Increment(authRejections, 1)
}

// RecordTokenRefreshes records token refresh
func (t *TelemetryImpl) RecordTokenRefreshes() {
	t.storage.Increment(tokenRefreshes, 1)
}

// PopAuthRejections returns all the rejections
func (t *TelemetryImpl) PopAuthRejections() int64 {
	return t.storage.PopCounter(authRejections)
}

// PopTokenRefreshes returns all the refreshes made
func (t *TelemetryImpl) PopTokenRefreshes() int64 {
	return t.storage.PopCounter(tokenRefreshes)
}

// STREAMING

// RecordPrimaryOccupancyChange records occupancy on primary
func (t *TelemetryImpl) RecordPrimaryOccupancyChange(newPublisherCount int64) {
	t.storage.PushItem(streamingType, StreamingEvent{
		Type:      eventTypeOccupancyPri,
		Data:      newPublisherCount,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordSecondaryOccupancyChange records occupancy on secondary
func (t *TelemetryImpl) RecordSecondaryOccupancyChange(newPublisherCount int64) {
	t.storage.PushItem(streamingType, StreamingEvent{
		Type:      eventTypeOccupancySec,
		Data:      newPublisherCount,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordConnectionSuccess records success on SSE
func (t *TelemetryImpl) RecordConnectionSuccess() {
	t.storage.PushItem(streamingType, StreamingEvent{
		Type:      eventTypeSSEConnectionEstablished,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordStreamingServiceStatus records new status on streaming
func (t *TelemetryImpl) RecordStreamingServiceStatus(newStatus int) {
	t.storage.PushItem(streamingType, StreamingEvent{
		Type:      eventTypeStreamingStatus,
		Data:      int64(newStatus),
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordTokenRefresh records next expiration event
func (t *TelemetryImpl) RecordTokenRefresh(tokenExpirationUtcTs int64) {
	t.storage.PushItem(streamingType, StreamingEvent{
		Type:      eventTypeTokenRefresh,
		Data:      tokenExpirationUtcTs,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordAblyError records erros in ably SSE
func (t *TelemetryImpl) RecordAblyError(statusCode int64) {
	t.storage.PushItem(streamingType, StreamingEvent{
		Type:      eventTypeAblyError,
		Data:      statusCode,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordConnectionClose records connections close
func (t *TelemetryImpl) RecordConnectionClose(wasRequested bool) {
	data := requested
	if !wasRequested {
		data = nonRequested
	}
	t.storage.PushItem(streamingType, StreamingEvent{
		Type:      eventTypeConnectionError,
		Data:      int64(data),
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordSyncModeUpdate records updates in streaming
func (t *TelemetryImpl) RecordSyncModeUpdate(newSyncMode int64) {
	t.storage.PushItem(streamingType, StreamingEvent{
		Type:      eventTypeSyncMode,
		Data:      newSyncMode,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// PopStreamingEvents returns all the stored StreamingEvents
func (t *TelemetryImpl) PopStreamingEvents() []StreamingEvent {
	events := t.storage.PopItems(streamingType)
	toReturn, ok := events.([]StreamingEvent)
	if !ok {
		return []StreamingEvent{}
	}
	return toReturn
}

// TAGS

// AddTag add tags to misc data
func (t *TelemetryImpl) AddTag(tag string) {
	t.storage.PushItem(tagsType, tag)
}

// PopTags returns all the tags stored
func (t *TelemetryImpl) PopTags() []string {
	data := t.storage.PopItems(tagsType)
	toReturn, ok := data.([]string)
	if !ok {
		return []string{}
	}
	return toReturn
}

// SDK

// RecordSessionLength stores session duration
func (t *TelemetryImpl) RecordSessionLength(session int64) {
	t.storage.Set(sessionType, session)
}

// GetSessionLength returns stored session
func (t *TelemetryImpl) GetSessionLength() int64 {
	return t.storage.GetRecord(sessionType)
}

// FACTORY

// RecordFactory stores factory
func (t *TelemetryImpl) RecordFactory(apikey string) {
	t.storage.Increment(fmt.Sprintf("%s::%s", factoryType, apikey), 1)
}

// RecordNonReadyUsage stores non ready usage
func (t *TelemetryImpl) RecordNonReadyUsage() {
	t.storage.Increment(nonReadyUsages, 1)
}

// RecordBURTimeout stores timeouts
func (t *TelemetryImpl) RecordBURTimeout() {
	t.storage.Increment(burTimeouts, 1)
}

// RecordTimeUntilReady stores time duration
func (t *TelemetryImpl) RecordTimeUntilReady(time int64) {
	t.storage.Set(timeUntilReady, time)
}

// GetActiveFactories gets active factories
func (t *TelemetryImpl) GetActiveFactories() int64 {
	factoriesData := t.storage.GetFactories()
	var toReturn int64 = 0
	for _, activeFactories := range factoriesData {
		toReturn += activeFactories
	}
	return toReturn
}

// GetRedundantActiveFactories gets redundant factories
func (t *TelemetryImpl) GetRedundantActiveFactories() int64 {
	factoriesData := t.storage.GetFactories()
	var toReturn int64 = 0
	for _, activeFactory := range factoriesData {
		if activeFactory > 1 {
			toReturn += activeFactory - 1
		}
	}
	return toReturn
}

// GetNonReadyUsages gets non ready usages
func (t *TelemetryImpl) GetNonReadyUsages() int64 {
	return t.storage.PopCounter(nonReadyUsages)
}

// GetBURTimeouts gets bur timeots
func (t *TelemetryImpl) GetBURTimeouts() int64 {
	return t.storage.PopCounter(burTimeouts)
}

// GetTimeUntilReady returns stored session
func (t *TelemetryImpl) GetTimeUntilReady() int64 {
	return t.storage.GetRecord(timeUntilReady)
}
