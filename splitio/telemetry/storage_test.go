package telemetry

import (
	"fmt"
	"testing"
	"time"

	"github.com/splitio/go-split-commons/dtos"
	"github.com/splitio/go-split-commons/storage/mutexmap"
	"github.com/splitio/go-toolkit/datastructures/set"
)

func TestTelemetryStorage(t *testing.T) {
	splitStorage := mutexmap.NewMMSplitStorage()
	splits := make([]dtos.SplitDTO, 0, 10)
	for index := 0; index < 10; index++ {
		splits = append(splits, dtos.SplitDTO{
			Name: fmt.Sprintf("SomeSplit_%d", index),
			Algo: index,
		})
	}
	splitStorage.PutMany(splits, 123)
	segmentStorage := mutexmap.NewMMSegmentStorage()
	segmentStorage.Update("some", set.NewSet("yaris", "redo"), set.NewSet(), 123456789)

	telemetryService := NewTelemetry(NewIMTelemetryStorage(), splitStorage, segmentStorage)

	telemetryService.RecordException(treatment)
	telemetryService.RecordException(treatments)
	telemetryService.RecordException(treatment)
	telemetryService.RecordLatency(treatment, (1500 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordLatency(treatment, (2000 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordLatency(treatments, (3000 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordLatency(treatments, (500 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordLatency(treatmentWithConfig, (800 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordLatency(treatmentsWithConfig, (1000 * time.Nanosecond).Nanoseconds())

	exceptions := telemetryService.PopExceptions()
	if exceptions.Treatment != 2 || exceptions.Treatments != 1 || exceptions.TreatmentWithConfig != 0 || exceptions.TreatmentWithConfigs != 0 || exceptions.Track != 0 {
		t.Error("Wrong result")
	}
	exceptions = telemetryService.PopExceptions()
	if exceptions.Treatment != 0 || exceptions.Treatments != 0 || exceptions.TreatmentWithConfig != 0 || exceptions.TreatmentWithConfigs != 0 || exceptions.Track != 0 {
		t.Error("Wrong result")
	}
	latencies := telemetryService.PopLatencies()
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
	latencies = telemetryService.PopLatencies()
	if latencies.Treatment[1] != 0 {
		t.Error("Wrong result")
	}

	telemetryService.RecordQueuedImpressions(200)
	telemetryService.RecordDedupedImpressions(100)
	telemetryService.RecordDroppedImpressions(50)
	telemetryService.RecordQueuedImpressions(200)
	if telemetryService.GetDedupedImpressions() != 100 {
		t.Error("Wrong result")
	}
	if telemetryService.GetQueuedmpressions() != 400 {
		t.Error("Wrong result")
	}
	if telemetryService.GetDroppedImpressions() != 50 {
		t.Error("Wrong result")
	}

	telemetryService.RecordDroppedEvents(100)
	telemetryService.RecordQueuedEvents(10)
	telemetryService.RecordDroppedEvents(100)
	telemetryService.RecordQueuedEvents(10)
	if telemetryService.GetDroppedEvents() != 200 {
		t.Error("Wrong result")
	}
	if telemetryService.GetQueuedEvents() != 20 {
		t.Error("Wrong result")
	}

	telemetryService.RecordSuccessfulSplitSync()
	time.Sleep(100 * time.Millisecond)
	telemetryService.RecordSuccessfulSegmentSync()
	time.Sleep(100 * time.Millisecond)
	telemetryService.RecordSuccessfulImpressionSync()
	time.Sleep(100 * time.Millisecond)
	telemetryService.RecordSuccessfulEventsSync()
	time.Sleep(100 * time.Millisecond)
	telemetryService.RecordSuccessfulTelemetrySync()
	time.Sleep(100 * time.Millisecond)
	telemetryService.RecordSuccessfulTokenGet()

	lastSynchronization := telemetryService.GetLastSynchronization()
	if lastSynchronization.Splits == 0 || lastSynchronization.Segments == 0 || lastSynchronization.Impressions == 0 || lastSynchronization.Events == 0 || lastSynchronization.Telemetry == 0 {
		t.Error("Wrong result")
	}

	telemetryService.RecordSyncError(split, 500)
	telemetryService.RecordSyncError(split, 500)
	telemetryService.RecordSyncError(split, 500)
	telemetryService.RecordSyncError(split, 500)
	telemetryService.RecordSyncError(split, 500)
	telemetryService.RecordSyncError(segment, 401)
	telemetryService.RecordSyncError(segment, 401)
	telemetryService.RecordSyncError(segment, 401)
	telemetryService.RecordSyncError(segment, 404)
	telemetryService.RecordSyncError(impression, 402)
	telemetryService.RecordSyncError(impression, 402)
	telemetryService.RecordSyncError(impression, 402)
	telemetryService.RecordSyncError(impression, 402)
	telemetryService.RecordSyncError(event, 400)
	telemetryService.RecordSyncError(telemetry, 401)
	telemetryService.RecordSyncError(token, 400)
	telemetryService.RecordSyncLatency(split, (1500 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(split, (3000 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(split, (4000 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(segment, (1500 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(segment, (1500 * time.Nanosecond).Nanoseconds())

	httpErrors := telemetryService.PopHTTPErrors()
	if httpErrors.Splits[500] != 5 || httpErrors.Segments[401] != 3 || httpErrors.Segments[404] != 1 || httpErrors.Impressions[402] != 4 || httpErrors.Events[400] != 1 || httpErrors.Telemetry[401] != 1 || httpErrors.Token[400] != 1 {
		t.Error("Wrong result")
	}
	httpErrors = telemetryService.PopHTTPErrors()
	if len(httpErrors.Splits) != 0 || len(httpErrors.Segments) != 0 { // and so on
		t.Error("Wrong result")
	}

	httpLatencies := telemetryService.PopHTTPLatencies()
	if httpLatencies.Splits[1] != 1 { // and so on
		t.Error("Wrong result")
	}
	httpLatencies = telemetryService.PopHTTPLatencies()
	if httpLatencies.Splits[1] != 0 { // and so on
		t.Error("Wrong result")
	}

	if telemetryService.GetSplitsCount() != 10 {
		t.Error("Wrong result")
	}

	telemetryService.RecordAuthRejections()
	telemetryService.RecordAuthRejections()
	telemetryService.RecordTokenRefreshes()
	telemetryService.RecordAuthRejections()

	if telemetryService.PopAuthRejections() != 3 {
		t.Error("Wrong result")
	}
	if telemetryService.PopAuthRejections() != 0 {
		t.Error("Wrong result")
	}
	if telemetryService.PopTokenRefreshes() != 1 {
		t.Error("Wrong result")
	}
	if telemetryService.PopTokenRefreshes() != 0 {
		t.Error("Wrong result")
	}

	telemetryService.RecordAblyError(40010)
	telemetryService.RecordConnectionSuccess()
	telemetryService.RecordPrimaryOccupancyChange(10)
	telemetryService.RecordSecondaryOccupancyChange(1)
	telemetryService.RecordSyncModeUpdate(0)

	if len(telemetryService.PopStreamingEvents()) != 5 {
		t.Error("Wrong result")
	}
	if len(telemetryService.PopStreamingEvents()) != 0 {
		t.Error("Wrong result")
	}

	telemetryService.RecordSessionLength(123456789)
	if telemetryService.GetSessionLength() != 123456789 {
		t.Error("Wrong result")
	}

	telemetryService.AddTag("redo")
	telemetryService.AddTag("yaris")
	tags := telemetryService.PopTags()
	if len(tags) != 2 {
		t.Error("Wrong result")
	}
	if len(telemetryService.PopTags()) != 0 {
		t.Error("Wrong result")
	}

	telemetryService.RecordBURTimeout()
	telemetryService.RecordBURTimeout()
	telemetryService.RecordBURTimeout()
	telemetryService.RecordFactory("123456789")
	telemetryService.RecordFactory("123456789")
	telemetryService.RecordFactory("123456789")
	telemetryService.RecordFactory("987654321")
	telemetryService.RecordNonReadyUsage()
	telemetryService.RecordNonReadyUsage()
	telemetryService.RecordNonReadyUsage()
	telemetryService.RecordNonReadyUsage()
	telemetryService.RecordNonReadyUsage()
	if telemetryService.GetBURTimeouts() != 3 {
		t.Error("Wrong result")
	}
	if telemetryService.GetNonReadyUsages() != 5 {
		t.Error("Wrong result")
	}
	if telemetryService.GetActiveFactories() != 4 {
		t.Error("Wrong result")
	}
	if telemetryService.GetRedundantActiveFactories() != 2 {
		t.Error("Wrong result")
	}
}
