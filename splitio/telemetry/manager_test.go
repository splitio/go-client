package telemetry

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/splitio/go-client/splitio/conf"
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
	telemetryService.RecordQueuedImpressions(200)
	telemetryService.RecordDedupedImpressions(100)
	telemetryService.RecordDroppedImpressions(50)
	telemetryService.RecordQueuedImpressions(200)
	telemetryService.RecordDroppedEvents(100)
	telemetryService.RecordQueuedEvents(10)
	telemetryService.RecordDroppedEvents(100)
	telemetryService.RecordQueuedEvents(10)
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
	telemetryService.RecordSyncError(splitSync, 500)
	telemetryService.RecordSyncError(splitSync, 500)
	telemetryService.RecordSyncError(splitSync, 500)
	telemetryService.RecordSyncError(splitSync, 500)
	telemetryService.RecordSyncError(splitSync, 500)
	telemetryService.RecordSyncError(segmentSync, 401)
	telemetryService.RecordSyncError(segmentSync, 401)
	telemetryService.RecordSyncError(segmentSync, 401)
	telemetryService.RecordSyncError(segmentSync, 404)
	telemetryService.RecordSyncError(impressionSync, 402)
	telemetryService.RecordSyncError(impressionSync, 402)
	telemetryService.RecordSyncError(impressionSync, 402)
	telemetryService.RecordSyncError(impressionSync, 402)
	telemetryService.RecordSyncError(eventSync, 400)
	telemetryService.RecordSyncError(telemetrySync, 401)
	telemetryService.RecordSyncError(tokenSync, 400)
	telemetryService.RecordSyncLatency(splitSync, (1500 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(splitSync, (3000 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(splitSync, (4000 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(segmentSync, (1500 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordSyncLatency(segmentSync, (1500 * time.Nanosecond).Nanoseconds())
	telemetryService.RecordAuthRejections()
	telemetryService.RecordAuthRejections()
	telemetryService.RecordTokenRefreshes()
	telemetryService.RecordAuthRejections()
	telemetryService.RecordAblyError(40010)
	telemetryService.RecordConnectionSuccess()
	telemetryService.RecordPrimaryOccupancyChange(10)
	telemetryService.RecordSecondaryOccupancyChange(1)
	telemetryService.RecordSyncModeUpdate(0)
	telemetryService.RecordSessionLength(123456789)
	if telemetryService.GetSessionLength() != 123456789 {
		t.Error("Wrong result")
	}
	telemetryService.AddTag("redo")
	telemetryService.AddTag("yaris")
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

	manager := NewTelemetryManager(
		telemetryService,
		telemetryService,
		telemetryService,
		telemetryService,
		telemetryService,
		telemetryService,
		telemetryService,
		telemetryService,
		telemetryService,
		telemetryService,
		telemetryService,
	)

	config := manager.BuildInitData(conf.Default())
	data, _ := json.Marshal(config)
	if data == nil {
		t.Error("")
	}

	regular := manager.BuildStatsData()

	result, _ := json.Marshal(regular)
	if result == nil {
		t.Error("")
	}

	regular = manager.BuildStatsData()

	result, _ = json.Marshal(regular)
	if result == nil {
		t.Error("")
	}
}
