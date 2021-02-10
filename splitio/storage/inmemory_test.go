package storage

import (
	"testing"
	"time"

	"github.com/splitio/go-client/splitio/dto"
	"github.com/splitio/go-split-commons/util"
)

func TestTelemetryStorage(t *testing.T) {
	telemetryStorage := NewIMTelemetryStorage()

	telemetryStorage.RecordException(treatment)
	telemetryStorage.RecordException(treatments)
	telemetryStorage.RecordException(treatment)
	telemetryStorage.RecordLatency(treatment, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))
	telemetryStorage.RecordLatency(treatment, util.Bucket((2000 * time.Nanosecond).Nanoseconds()))
	telemetryStorage.RecordLatency(treatments, util.Bucket((3000 * time.Nanosecond).Nanoseconds()))
	telemetryStorage.RecordLatency(treatments, util.Bucket((500 * time.Nanosecond).Nanoseconds()))
	telemetryStorage.RecordLatency(treatmentWithConfig, util.Bucket((800 * time.Nanosecond).Nanoseconds()))
	telemetryStorage.RecordLatency(treatmentsWithConfig, util.Bucket((1000 * time.Nanosecond).Nanoseconds()))

	exceptions := telemetryStorage.PopExceptions()
	if exceptions.Treatment != 2 || exceptions.Treatments != 1 || exceptions.TreatmentWithConfig != 0 || exceptions.TreatmentWithConfigs != 0 || exceptions.Track != 0 {
		t.Error("Wrong result")
	}
	exceptions = telemetryStorage.PopExceptions()
	if exceptions.Treatment != 0 || exceptions.Treatments != 0 || exceptions.TreatmentWithConfig != 0 || exceptions.TreatmentWithConfigs != 0 || exceptions.Track != 0 {
		t.Error("Wrong result")
	}
	latencies := telemetryStorage.PopLatencies()
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
	latencies = telemetryStorage.PopLatencies()
	if latencies.Treatment[1] != 0 {
		t.Error("Wrong result")
	}

	telemetryStorage.RecordQueuedImpressions(200)
	telemetryStorage.RecordDedupedImpressions(100)
	telemetryStorage.RecordDroppedImpressions(50)
	telemetryStorage.RecordQueuedImpressions(200)
	if telemetryStorage.GetDedupedImpressions() != 100 {
		t.Error("Wrong result")
	}
	if telemetryStorage.GetQueuedmpressions() != 400 {
		t.Error("Wrong result")
	}
	if telemetryStorage.GetDroppedImpressions() != 50 {
		t.Error("Wrong result")
	}

	telemetryStorage.RecordDroppedEvents(100)
	telemetryStorage.RecordQueuedEvents(10)
	telemetryStorage.RecordDroppedEvents(100)
	telemetryStorage.RecordQueuedEvents(10)
	if telemetryStorage.GetDroppedEvents() != 200 {
		t.Error("Wrong result")
	}
	if telemetryStorage.GetQueuedEvents() != 20 {
		t.Error("Wrong result")
	}

	telemetryStorage.RecordSuccessfulSplitSync(time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	telemetryStorage.RecordSuccessfulSegmentSync(time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	telemetryStorage.RecordSuccessfulImpressionSync(time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	telemetryStorage.RecordSuccessfulEventsSync(time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	telemetryStorage.RecordSuccessfulTelemetrySync(time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	telemetryStorage.RecordSuccessfulTokenGet(time.Now().UnixNano())

	lastSynchronization := telemetryStorage.GetLastSynchronization()
	if lastSynchronization.Splits == 0 || lastSynchronization.Segments == 0 || lastSynchronization.Impressions == 0 || lastSynchronization.Events == 0 || lastSynchronization.Telemetry == 0 {
		t.Error("Wrong result")
	}

	telemetryStorage.RecordSyncError(splitSync, 500)
	telemetryStorage.RecordSyncError(splitSync, 500)
	telemetryStorage.RecordSyncError(splitSync, 500)
	telemetryStorage.RecordSyncError(splitSync, 500)
	telemetryStorage.RecordSyncError(splitSync, 500)
	telemetryStorage.RecordSyncError(segmentSync, 401)
	telemetryStorage.RecordSyncError(segmentSync, 401)
	telemetryStorage.RecordSyncError(segmentSync, 401)
	telemetryStorage.RecordSyncError(segmentSync, 404)
	telemetryStorage.RecordSyncError(impressionSync, 402)
	telemetryStorage.RecordSyncError(impressionSync, 402)
	telemetryStorage.RecordSyncError(impressionSync, 402)
	telemetryStorage.RecordSyncError(impressionSync, 402)
	telemetryStorage.RecordSyncError(eventSync, 400)
	telemetryStorage.RecordSyncError(telemetrySync, 401)
	telemetryStorage.RecordSyncError(tokenSync, 400)
	telemetryStorage.RecordSyncLatency(splitSync, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))
	telemetryStorage.RecordSyncLatency(splitSync, util.Bucket((3000 * time.Nanosecond).Nanoseconds()))
	telemetryStorage.RecordSyncLatency(splitSync, util.Bucket((4000 * time.Nanosecond).Nanoseconds()))
	telemetryStorage.RecordSyncLatency(segmentSync, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))
	telemetryStorage.RecordSyncLatency(segmentSync, util.Bucket((1500 * time.Nanosecond).Nanoseconds()))

	httpErrors := telemetryStorage.PopHTTPErrors()
	if httpErrors.Splits[500] != 5 || httpErrors.Segments[401] != 3 || httpErrors.Segments[404] != 1 || httpErrors.Impressions[402] != 4 || httpErrors.Events[400] != 1 || httpErrors.Telemetry[401] != 1 || httpErrors.Token[400] != 1 {
		t.Error("Wrong result")
	}
	httpErrors = telemetryStorage.PopHTTPErrors()
	if len(httpErrors.Splits) != 0 || len(httpErrors.Segments) != 0 { // and so on
		t.Error("Wrong result")
	}

	httpLatencies := telemetryStorage.PopHTTPLatencies()
	if httpLatencies.Splits[1] != 1 { // and so on
		t.Error("Wrong result")
	}
	httpLatencies = telemetryStorage.PopHTTPLatencies()
	if httpLatencies.Splits[1] != 0 { // and so on
		t.Error("Wrong result")
	}

	telemetryStorage.RecordAuthRejections()
	telemetryStorage.RecordAuthRejections()
	telemetryStorage.RecordTokenRefreshes()
	telemetryStorage.RecordAuthRejections()

	if telemetryStorage.PopAuthRejections() != 3 {
		t.Error("Wrong result")
	}
	if telemetryStorage.PopAuthRejections() != 0 {
		t.Error("Wrong result")
	}
	if telemetryStorage.PopTokenRefreshes() != 1 {
		t.Error("Wrong result")
	}
	if telemetryStorage.PopTokenRefreshes() != 0 {
		t.Error("Wrong result")
	}

	telemetryStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})
	telemetryStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})
	telemetryStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})
	telemetryStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})
	telemetryStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})
	telemetryStorage.RecordStreamingEvent(dto.StreamingEvent{Type: 1, Data: 1, Timestamp: 123456789})

	if len(telemetryStorage.PopStreamingEvents()) != 6 {
		t.Error("Wrong result")
	}
	if len(telemetryStorage.PopStreamingEvents()) != 0 {
		t.Error("Wrong result")
	}

	telemetryStorage.RecordSessionLength(123456789)
	if telemetryStorage.GetSessionLength() != 123456789 {
		t.Error("Wrong result")
	}

	telemetryStorage.AddTag("redo")
	telemetryStorage.AddTag("yaris")
	tags := telemetryStorage.PopTags()
	if len(tags) != 2 {
		t.Error("Wrong result")
	}
	if len(telemetryStorage.PopTags()) != 0 {
		t.Error("Wrong result")
	}

	telemetryStorage.RecordBURTimeout()
	telemetryStorage.RecordBURTimeout()
	telemetryStorage.RecordBURTimeout()
	telemetryStorage.RecordFactory("123456789")
	telemetryStorage.RecordFactory("123456789")
	telemetryStorage.RecordFactory("123456789")
	telemetryStorage.RecordFactory("987654321")
	telemetryStorage.RecordNonReadyUsage()
	telemetryStorage.RecordNonReadyUsage()
	telemetryStorage.RecordNonReadyUsage()
	telemetryStorage.RecordNonReadyUsage()
	telemetryStorage.RecordNonReadyUsage()
	if telemetryStorage.GetBURTimeouts() != 3 {
		t.Error("Wrong result")
	}
	if telemetryStorage.GetNonReadyUsages() != 5 {
		t.Error("Wrong result")
	}
	if telemetryStorage.GetActiveFactories() != 4 {
		t.Error("Wrong result")
	}
	if telemetryStorage.GetRedundantActiveFactories() != 2 {
		t.Error("Wrong result")
	}
}
