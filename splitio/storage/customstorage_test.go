package storage

import (
	"strings"
	"testing"
	"time"

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

func convertToInt64(item interface{}) int64 {
	switch t := item.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	default:
		return 0
	}
}

func (t TestWrapper) GetItem(key string) int64 {
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

func (t TestWrapper) PopItem(key string) int64 {
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

func (t TestWrapper) Set(key string, value int64) {
	t.data[key] = value
}

func TestCustomStorage(t *testing.T) {
	customStorage := UserCustomTelemetryAdapter{
		wrapper: TestWrapper{
			data: make(map[string]interface{}),
		},
	}

	customStorage.RecordException(treatment)
	customStorage.RecordException(treatments)
	customStorage.RecordException(treatment)
	exceptions := customStorage.PopExceptions()
	if exceptions.Treatment != 2 || exceptions.Treatments != 1 || exceptions.TreatmentWithConfig != 0 || exceptions.TreatmentWithConfigs != 0 || exceptions.Track != 0 {
		t.Error("Wrong result")
	}
	exceptions = customStorage.PopExceptions()
	if exceptions.Treatment != 0 || exceptions.Treatments != 0 || exceptions.TreatmentWithConfig != 0 || exceptions.TreatmentWithConfigs != 0 || exceptions.Track != 0 {
		t.Error("Wrong result")
	}

	customStorage.RecordLatency(treatment, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordLatency(treatment, util.Bucket((2000 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordLatency(treatments, util.Bucket((3000 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordLatency(treatments, util.Bucket((500 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordLatency(treatmentWithConfig, util.Bucket((800 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordLatency(treatmentsWithConfig, util.Bucket((1000 * time.Nanosecond).Nanoseconds()))
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

	customStorage.RecordQueuedImpressions(200)
	customStorage.RecordDedupedImpressions(100)
	customStorage.RecordDroppedImpressions(50)
	customStorage.RecordQueuedImpressions(200)
	if customStorage.GetDedupedImpressions() != 100 {
		t.Error("Wrong result")
	}
	if customStorage.GetQueuedmpressions() != 400 {
		t.Error("Wrong result")
	}
	if customStorage.GetDroppedImpressions() != 50 {
		t.Error("Wrong result")
	}

	customStorage.RecordDroppedEvents(100)
	customStorage.RecordQueuedEvents(10)
	customStorage.RecordDroppedEvents(100)
	customStorage.RecordQueuedEvents(10)
	if customStorage.GetDroppedEvents() != 200 {
		t.Error("Wrong result")
	}
	if customStorage.GetQueuedEvents() != 20 {
		t.Error("Wrong result")
	}

	customStorage.RecordSuccessfulSplitSync(time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	customStorage.RecordSuccessfulSegmentSync(time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	customStorage.RecordSuccessfulImpressionSync(time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	customStorage.RecordSuccessfulEventsSync(time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	customStorage.RecordSuccessfulTelemetrySync(time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	customStorage.RecordSuccessfulTokenGet(time.Now().UnixNano())
	lastSynchronization := customStorage.GetLastSynchronization()
	if lastSynchronization.Splits == 0 || lastSynchronization.Segments == 0 || lastSynchronization.Impressions == 0 || lastSynchronization.Events == 0 || lastSynchronization.Telemetry == 0 {
		t.Error("Wrong result")
	}

	customStorage.RecordSyncError(splitSync, 500)
	customStorage.RecordSyncError(splitSync, 500)
	customStorage.RecordSyncError(splitSync, 500)
	customStorage.RecordSyncError(splitSync, 500)
	customStorage.RecordSyncError(splitSync, 500)
	customStorage.RecordSyncError(segmentSync, 401)
	customStorage.RecordSyncError(segmentSync, 401)
	customStorage.RecordSyncError(segmentSync, 401)
	customStorage.RecordSyncError(segmentSync, 404)
	customStorage.RecordSyncError(impressionSync, 402)
	customStorage.RecordSyncError(impressionSync, 402)
	customStorage.RecordSyncError(impressionSync, 402)
	customStorage.RecordSyncError(impressionSync, 402)
	customStorage.RecordSyncError(eventSync, 400)
	customStorage.RecordSyncError(telemetrySync, 401)
	customStorage.RecordSyncError(tokenSync, 400)
	httpErrors := customStorage.PopHTTPErrors()
	if httpErrors.Splits[500] != 5 || httpErrors.Segments[401] != 3 || httpErrors.Segments[404] != 1 || httpErrors.Impressions[402] != 4 || httpErrors.Events[400] != 1 || httpErrors.Telemetry[401] != 1 || httpErrors.Token[400] != 1 {
		t.Error("Wrong result")
	}
	httpErrors = customStorage.PopHTTPErrors()
	if len(httpErrors.Splits) != 0 || len(httpErrors.Segments) != 0 { // and so on
		t.Error("Wrong result")
	}

	customStorage.RecordSyncLatency(splitSync, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordSyncLatency(splitSync, util.Bucket((3000 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordSyncLatency(splitSync, util.Bucket((4000 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordSyncLatency(segmentSync, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))
	customStorage.RecordSyncLatency(segmentSync, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))
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
	customStorage.RecordFactory("123456789")
	customStorage.RecordFactory("123456789")
	customStorage.RecordFactory("123456789")
	customStorage.RecordFactory("987654321")
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
	if customStorage.GetActiveFactories() != 4 {
		t.Error("Wrong result")
	}
	if customStorage.GetRedundantActiveFactories() != 2 {
		t.Error("Wrong result")
	}

	customStorage.AddIntegration("some")
	customStorage.AddIntegration("other")
	if len(customStorage.GetIntegrations()) != 2 {
		t.Error("Wrong result")
	}
}
