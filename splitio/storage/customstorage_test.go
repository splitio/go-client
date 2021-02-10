package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/splitio/go-client/splitio/constants"
	"github.com/splitio/go-client/splitio/dto"
	"github.com/splitio/go-split-commons/util"
)

type TestWrapper struct {
	data map[string]interface{}
}

func (t TestWrapper) GetByPrefix(prefix string) []interface{} {
	toReturn := make([]interface{}, 0)
	for key := range t.data {
		if strings.HasPrefix(key, prefix) {
			toReturn = append(toReturn, key)
		}
	}
	return toReturn
}

func (t TestWrapper) GetItem(key string) interface{} {
	asInterface, ok := t.data[key]
	if !ok {
		return 0
	}
	return convertToInt64(asInterface)
}

func (t TestWrapper) Increment(key string, value int64) {
	asInterface, ok := t.data[key]
	if !ok {
		t.data[key] = value
		return
	}
	t.data[key] = convertToInt64(asInterface) + value
}

func (t TestWrapper) PopItem(key string) interface{} {
	asInterface, ok := t.data[key]
	if !ok {
		return 0
	}
	delete(t.data, key)
	return convertToInt64(asInterface)
}

func (t TestWrapper) PopItems(key string) interface{} {
	items, ok := t.data[key]
	if !ok {
		return nil
	}
	delete(t.data, key)
	return items
}

func (t TestWrapper) PushItem(key string, item interface{}) {
	asInterface, ok := t.data[key]
	if !ok {
		t.data[key] = []interface{}{item}
		return
	}
	slice, ok := asInterface.([]interface{})
	slice = append(slice, item)
	t.data[key] = slice
}

func (t TestWrapper) Set(key string, value interface{}) {
	t.data[key] = value
}

func TestCustomStorage(t *testing.T) {
	customStorage := UserCustomTelemetryAdapter{
		wrapper: TestWrapper{
			data: make(map[string]interface{}),
		},
	}

	customStorage.RecordException(constants.Treatment)
	customStorage.RecordException(constants.Treatments)
	customStorage.RecordException(constants.Treatment)
	exceptions := customStorage.PopExceptions()
	if exceptions.Treatment != 2 || exceptions.Treatments != 1 || exceptions.TreatmentWithConfig != 0 || exceptions.TreatmentWithConfigs != 0 || exceptions.Track != 0 {
		t.Error("Wrong result")
	}
	exceptions = customStorage.PopExceptions()
	if exceptions.Treatment != 0 || exceptions.Treatments != 0 || exceptions.TreatmentWithConfig != 0 || exceptions.TreatmentWithConfigs != 0 || exceptions.Track != 0 {
		t.Error("Wrong result")
	}

	customStorage.RecordLatency(constants.Treatment, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordLatency(constants.Treatment, util.Bucket((2000 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordLatency(constants.Treatments, util.Bucket((3000 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordLatency(constants.Treatments, util.Bucket((500 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordLatency(constants.TreatmentWithConfig, util.Bucket((800 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordLatency(constants.TreatmentsWithConfig, util.Bucket((1000 * time.Nanosecond).Nanoseconds()))
	latencies := customStorage.PopLatencies()
	if latencies.Treatment[1] != 1 || latencies.Treatment[2] != 1 {
		t.Error("Wrong result")
	}
	if latencies.Treatments[0] != 1 || latencies.Treatments[3] != 1 {
		t.Error("Wrong result")
	}
	if latencies.TreatmentWithConfig[0] != 1 {
		t.Error("Wrong result")
	}
	if latencies.TreatmentWithConfigs[0] != 1 {
		t.Error("Wrong result")
	}
	latencies = customStorage.PopLatencies()
	if latencies.Treatment[1] != 0 {
		t.Error("Wrong result")
	}

	customStorage.RecordImpressionsStats(constants.ImpressionsQueued, 200)
	customStorage.RecordImpressionsStats(constants.ImpressionsDeduped, 100)
	customStorage.RecordImpressionsStats(constants.ImpressionsDropped, 50)
	customStorage.RecordImpressionsStats(constants.ImpressionsQueued, 200)
	if customStorage.GetImpressionsStats(constants.ImpressionsDeduped) != 100 {
		t.Error("Wrong result")
	}
	if customStorage.GetImpressionsStats(constants.ImpressionsQueued) != 400 {
		t.Error("Wrong result")
	}
	if customStorage.GetImpressionsStats(constants.ImpressionsDropped) != 50 {
		t.Error("Wrong result")
	}

	customStorage.RecordEventsStats(constants.EventsDropped, 100)
	customStorage.RecordEventsStats(constants.EventsQueued, 10)
	customStorage.RecordEventsStats(constants.EventsDropped, 100)
	customStorage.RecordEventsStats(constants.EventsQueued, 10)
	if customStorage.GetEventsStats(constants.EventsDropped) != 200 {
		t.Error("Wrong result")
	}
	if customStorage.GetEventsStats(constants.EventsQueued) != 20 {
		t.Error("Wrong result")
	}

	customStorage.RecordSuccessfulSync(constants.SplitSync, time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	customStorage.RecordSuccessfulSync(constants.SegmentSync, time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	customStorage.RecordSuccessfulSync(constants.ImpressionSync, time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	customStorage.RecordSuccessfulSync(constants.EventSync, time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	customStorage.RecordSuccessfulSync(constants.TelemetrySync, time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	customStorage.RecordSuccessfulSync(constants.TokenSync, time.Now().UnixNano())
	lastSynchronization := customStorage.GetLastSynchronization()
	if lastSynchronization.Splits == 0 || lastSynchronization.Segments == 0 || lastSynchronization.Impressions == 0 || lastSynchronization.Events == 0 || lastSynchronization.Telemetry == 0 {
		t.Error("Wrong result")
	}
	customStorage.RecordSyncError(constants.SplitSync, 500)
	customStorage.RecordSyncError(constants.SplitSync, 500)
	customStorage.RecordSyncError(constants.SplitSync, 500)
	customStorage.RecordSyncError(constants.SplitSync, 500)
	customStorage.RecordSyncError(constants.SplitSync, 500)
	customStorage.RecordSyncError(constants.SegmentSync, 401)
	customStorage.RecordSyncError(constants.SegmentSync, 401)
	customStorage.RecordSyncError(constants.SegmentSync, 401)
	customStorage.RecordSyncError(constants.SegmentSync, 404)
	customStorage.RecordSyncError(constants.ImpressionSync, 402)
	customStorage.RecordSyncError(constants.ImpressionSync, 402)
	customStorage.RecordSyncError(constants.ImpressionSync, 402)
	customStorage.RecordSyncError(constants.ImpressionSync, 402)
	customStorage.RecordSyncError(constants.EventSync, 400)
	customStorage.RecordSyncError(constants.TelemetrySync, 401)
	customStorage.RecordSyncError(constants.TokenSync, 400)
	httpErrors := customStorage.PopHTTPErrors()
	if httpErrors.Splits[500] != 5 || httpErrors.Segments[401] != 3 || httpErrors.Segments[404] != 1 || httpErrors.Impressions[402] != 4 || httpErrors.Events[400] != 1 || httpErrors.Telemetry[401] != 1 || httpErrors.Token[400] != 1 {
		t.Error("Wrong result")
	}
	httpErrors = customStorage.PopHTTPErrors()
	if len(httpErrors.Splits) != 0 || len(httpErrors.Segments) != 0 { // and so on
		t.Error("Wrong result")
	}

	customStorage.RecordSyncLatency(constants.SplitSync, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordSyncLatency(constants.SplitSync, util.Bucket((3000 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordSyncLatency(constants.SplitSync, util.Bucket((4000 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordSyncLatency(constants.SegmentSync, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordSyncLatency(constants.SegmentSync, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))
	httpLatencies := customStorage.PopHTTPLatencies()
	if httpLatencies.Splits[1] != 1 { // and so on
		t.Error("Wrong result")
	}
	httpLatencies = customStorage.PopHTTPLatencies()
	if httpLatencies.Splits[1] != 0 { // and so on
		t.Error("Wrong result")
	}

	customStorage.RecordAuthRejections()
	customStorage.RecordAuthRejections()
	customStorage.RecordTokenRefreshes()
	customStorage.RecordAuthRejections()
	if customStorage.PopAuthRejections() != 3 {
		t.Error("Wrong result")
	}
	if customStorage.PopAuthRejections() != 0 {
		t.Error("Wrong result")
	}
	if customStorage.PopTokenRefreshes() != 1 {
		t.Error("Wrong result")
	}
	if customStorage.PopTokenRefreshes() != 0 {
		t.Error("Wrong result")
	}

	customStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})
	customStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})
	customStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})
	customStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})
	customStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})
	customStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})
	if len(customStorage.PopStreamingEvents()) != 6 {
		t.Error("Wrong result")
	}
	if len(customStorage.PopStreamingEvents()) != 0 {
		t.Error("Wrong result")
	}

	customStorage.RecordSessionLength(123456789)
	if customStorage.GetSessionLength() != 123456789 {
		t.Error("Wrong result")
	}

	customStorage.AddTag("redo")
	customStorage.AddTag("doc")
	tags := customStorage.PopTags()
	if len(tags) != 2 {
		t.Error("Wrong result")
	}
	if len(customStorage.PopTags()) != 0 {
		t.Error("Wrong result")
	}

	customStorage.RecordBURTimeout()
	customStorage.RecordBURTimeout()
	customStorage.RecordBURTimeout()
	customStorage.RecordNonReadyUsage()
	customStorage.RecordNonReadyUsage()
	customStorage.RecordNonReadyUsage()
	customStorage.RecordNonReadyUsage()
	customStorage.RecordNonReadyUsage()
	if customStorage.GetBURTimeouts() != 3 {
		t.Error("Wrong result")
	}
	if customStorage.GetNonReadyUsages() != 5 {
		t.Error("Wrong result")
	}
}
