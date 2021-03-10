package telemetry

import (
	"fmt"
	"testing"
	"time"

	"github.com/splitio/go-client/splitio/constants"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-split-commons/dtos"
	"github.com/splitio/go-split-commons/storage/mutexmap"
	"github.com/splitio/go-toolkit/datastructures/set"
)

func TestTelemetryService(t *testing.T) {
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
	segmentStorage.Update("some", set.NewSet("doc", "redo"), set.NewSet(), 123456789)

	telemetryService := NewTelemetry(storage.NewIMTelemetryStorage(), splitStorage, segmentStorage)

	telemetryService.RecordException(constants.Treatment)
	telemetryService.RecordException(constants.Treatments)
	telemetryService.RecordException(constants.Treatment)
	telemetryService.RecordLatency(constants.Treatment, (1500 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordLatency(constants.Treatment, (2000 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordLatency(constants.Treatments, (3000 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordLatency(constants.Treatments, (500 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordLatency(constants.TreatmentWithConfig, (800 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordLatency(constants.TreatmentsWithConfig, (1000 * time.Nanosecond).Nanoseconds())

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

	telemetryService.RecordImpressionsStats(constants.ImpressionsQueued, 200)
	telemetryService.RecordImpressionsStats(constants.ImpressionsDeduped, 100)
	telemetryService.RecordImpressionsStats(constants.ImpressionsDropped, 50)
	telemetryService.RecordImpressionsStats(constants.ImpressionsQueued, 200)
	if telemetryService.GetImpressionsStats(constants.ImpressionsDeduped) != 100 {
		t.Error("Wrong result")
	}
	if telemetryService.GetImpressionsStats(constants.ImpressionsQueued) != 400 {
		t.Error("Wrong result")
	}
	if telemetryService.GetImpressionsStats(constants.ImpressionsDropped) != 50 {
		t.Error("Wrong result")
	}

	telemetryService.RecordEventsStats(constants.EventsDropped, 100)
	telemetryService.RecordEventsStats(constants.EventsQueued, 10)
	telemetryService.RecordEventsStats(constants.EventsDropped, 100)
	telemetryService.RecordEventsStats(constants.EventsQueued, 10)
	if telemetryService.GetEventsStats(constants.EventsDropped) != 200 {
		t.Error("Wrong result")
	}
	if telemetryService.GetEventsStats(constants.EventsQueued) != 20 {
		t.Error("Wrong result")
	}

	telemetryService.RecordSuccessfulSync(constants.SplitSync)
	time.Sleep(100 * time.Millisecond)
	telemetryService.RecordSuccessfulSync(constants.SegmentSync)
	time.Sleep(100 * time.Millisecond)
	telemetryService.RecordSuccessfulSync(constants.ImpressionSync)
	time.Sleep(100 * time.Millisecond)
	telemetryService.RecordSuccessfulSync(constants.EventSync)
	time.Sleep(100 * time.Millisecond)
	telemetryService.RecordSuccessfulSync(constants.TelemetrySync)
	time.Sleep(100 * time.Millisecond)
	telemetryService.RecordSuccessfulSync(constants.TokenSync)

	lastSynchronization := telemetryService.GetLastSynchronization()
	if lastSynchronization.Splits == 0 || lastSynchronization.Segments == 0 || lastSynchronization.Impressions == 0 || lastSynchronization.Events == 0 || lastSynchronization.Telemetry == 0 {
		t.Error("Wrong result")
	}

	telemetryService.RecordSyncError(constants.SplitSync, 500)
	telemetryService.RecordSyncError(constants.SplitSync, 500)
	telemetryService.RecordSyncError(constants.SplitSync, 500)
	telemetryService.RecordSyncError(constants.SplitSync, 500)
	telemetryService.RecordSyncError(constants.SplitSync, 500)
	telemetryService.RecordSyncError(constants.SegmentSync, 401)
	telemetryService.RecordSyncError(constants.SegmentSync, 401)
	telemetryService.RecordSyncError(constants.SegmentSync, 401)
	telemetryService.RecordSyncError(constants.SegmentSync, 404)
	telemetryService.RecordSyncError(constants.ImpressionSync, 402)
	telemetryService.RecordSyncError(constants.ImpressionSync, 402)
	telemetryService.RecordSyncError(constants.ImpressionSync, 402)
	telemetryService.RecordSyncError(constants.ImpressionSync, 402)
	telemetryService.RecordSyncError(constants.EventSync, 400)
	telemetryService.RecordSyncError(constants.TelemetrySync, 401)
	telemetryService.RecordSyncError(constants.TokenSync, 400)
	telemetryService.RecordSyncLatency(constants.SplitSync, (1500 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(constants.SplitSync, (3000 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(constants.SplitSync, (4000 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(constants.SegmentSync, (1500 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(constants.SegmentSync, (1500 * time.Nanosecond).Nanoseconds())

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

	telemetryService.RecordStreamingEvent(constants.EventTypeAblyError, 40010)
	telemetryService.RecordStreamingEvent(constants.EventTypeSSEConnectionEstablished, 0)
	telemetryService.RecordStreamingEvent(constants.EventTypeOccupancyPri, 10)
	telemetryService.RecordStreamingEvent(constants.EventTypeOccupancySec, 1)
	telemetryService.RecordStreamingEvent(constants.EventTypeSyncMode, 1)

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
	telemetryService.AddTag("doc")
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
}
