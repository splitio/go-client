package telemetry

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/constants"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-split-commons/dtos"
	"github.com/splitio/go-split-commons/storage/mutexmap"
	"github.com/splitio/go-toolkit/datastructures/set"
)

func TestManager(t *testing.T) {
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
	telemetryService.RecordImpressionsStats(constants.ImpressionsQueued, 200)
	telemetryService.RecordImpressionsStats(constants.ImpressionsDeduped, 100)
	telemetryService.RecordImpressionsStats(constants.ImpressionsDropped, 50)
	telemetryService.RecordEventsStats(constants.EventsDropped, 100)
	telemetryService.RecordEventsStats(constants.EventsQueued, 10)
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
	telemetryService.RecordAuthRejections()
	telemetryService.RecordAuthRejections()
	telemetryService.RecordTokenRefreshes()
	telemetryService.RecordAuthRejections()
	telemetryService.RecordStreamingEvent(constants.EventTypeAblyError, 40010)
	telemetryService.RecordStreamingEvent(constants.EventTypeSSEConnectionEstablished, 0)
	telemetryService.RecordStreamingEvent(constants.EventTypeOccupancyPri, 10)
	telemetryService.RecordStreamingEvent(constants.EventTypeOccupancySec, 1)
	telemetryService.RecordStreamingEvent(constants.EventTypeSyncMode, 1)
	telemetryService.RecordSessionLength(123456789)
	telemetryService.RecordBURTimeout()
	telemetryService.RecordBURTimeout()
	telemetryService.RecordBURTimeout()
	telemetryService.RecordNonReadyUsage()
	telemetryService.RecordNonReadyUsage()
	telemetryService.RecordNonReadyUsage()
	telemetryService.RecordNonReadyUsage()
	telemetryService.RecordNonReadyUsage()

	manager := NewTelemetryManager(telemetryService)

	initData := manager.BuildInitData(conf.Default(), 123456789, make(map[string]int64))
	data, _ := json.Marshal(initData)
	if data == nil {
		t.Error("")
	}
	if len(string(data)) == 0 {
		t.Error("It should generated json")
	}

	statsData := manager.BuildStatsData()

	result, _ := json.Marshal(statsData)
	if result == nil {
		t.Error("")
	}
	if len(string(result)) == 0 {
		t.Error("It should generated json")
	}

	statsData = manager.BuildStatsData()
	result, _ = json.Marshal(statsData)
	if result == nil {
		t.Error("")
	}
	if len(string(result)) == 0 {
		t.Error("It should generated json")
	}
}
