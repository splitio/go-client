package telemetry

import (
	"time"

	"github.com/splitio/go-client/splitio/constants"
	"github.com/splitio/go-client/splitio/dto"
	"github.com/splitio/go-client/splitio/storage"
	commonStorage "github.com/splitio/go-split-commons/storage"
	"github.com/splitio/go-split-commons/util"
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
func (t *FacadeImpl) RecordException(method int) {
	t.storage.RecordException(method)
}

// RecordLatency records latencies for method
func (t *FacadeImpl) RecordLatency(method int, latency int64) {
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

// RecordImpressionsStats increments impressions data
func (t *FacadeImpl) RecordImpressionsStats(dataType int, count int64) {
	t.storage.RecordImpressionsStats(dataType, count)
}

// GetImpressionsStats returns impressions stats
func (t *FacadeImpl) GetImpressionsStats(dataType int) int64 {
	return t.storage.GetImpressionsStats(dataType)
}

// EVENTS

// RecordEventsStats records events stats
func (t *FacadeImpl) RecordEventsStats(dataType int, count int64) {
	t.storage.RecordEventsStats(dataType, count)
}

// GetEventsStats returns events stats
func (t *FacadeImpl) GetEventsStats(dataType int) int64 {
	return t.storage.GetEventsStats(dataType)
}

// LAST SYNCHRONIZATION

// RecordSuccessfulSync records sync
func (t *FacadeImpl) RecordSuccessfulSync(resource int) {
	t.storage.RecordSuccessfulSync(resource, time.Now().UTC().UnixNano()/1000000)
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
func (t *FacadeImpl) RecordSyncError(resource int, status int) {
	t.storage.RecordSyncError(resource, status)
}

// RecordSyncLatency records latencies
func (t *FacadeImpl) RecordSyncLatency(resource int, latency int64) {
	t.storage.RecordSyncLatency(resource, util.Bucket(latency))
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

// RecordStreamingEvent records streaming event
func (t *FacadeImpl) RecordStreamingEvent(eventType int, data int64) {
	switch eventType {
	case constants.EventTypeOccupancySec:
		fallthrough
	case constants.EventTypeSSEConnectionEstablished:
		fallthrough
	case constants.EventTypeStreamingStatus:
		fallthrough
	case constants.EventTypeTokenRefresh:
		fallthrough
	case constants.EventTypeAblyError:
		fallthrough
	case constants.EventTypeConnectionError:
		fallthrough
	case constants.EventTypeSyncMode:
		fallthrough
	case constants.EventTypeOccupancyPri:
		t.storage.RecordStreamingEvent(dto.StreamingEvent{
			Type:      eventType,
			Data:      data,
			Timestamp: time.Now().UTC().Unix(),
		})
	}
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

// RecordNonReadyUsage stores non ready usage
func (t *FacadeImpl) RecordNonReadyUsage() {
	t.storage.RecordNonReadyUsage()
}

// RecordBURTimeout stores timeouts
func (t *FacadeImpl) RecordBURTimeout() {
	t.storage.RecordBURTimeout()
}

// GetNonReadyUsages gets non ready usages
func (t *FacadeImpl) GetNonReadyUsages() int64 {
	return t.storage.GetNonReadyUsages()
}

// GetBURTimeouts gets bur timeots
func (t *FacadeImpl) GetBURTimeouts() int64 {
	return t.storage.GetBURTimeouts()
}
