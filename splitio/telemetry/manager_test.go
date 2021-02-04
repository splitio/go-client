package telemetry

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/splitio/go-client/splitio/conf"
)

func TestTelemetry(t *testing.T) {
	factoryTelemetry := NewFactoryTelemetryFacade()
	factoryTelemetry.RecordBURTimeout()
	factoryTelemetry.RecordBURTimeout()
	factoryTelemetry.RecordBURTimeout()
	factoryTelemetry.RecordFactory("123456789")
	factoryTelemetry.RecordFactory("123456789")
	factoryTelemetry.RecordFactory("123456789")
	factoryTelemetry.RecordFactory("987654321")
	factoryTelemetry.RecordNonReadyUsage()
	factoryTelemetry.RecordNonReadyUsage()
	factoryTelemetry.RecordNonReadyUsage()
	factoryTelemetry.RecordNonReadyUsage()
	factoryTelemetry.RecordNonReadyUsage()

	evaluationTelemetry := NewEvaluationTelemetryFacade()
	evaluationTelemetry.RecordException(treatment)
	evaluationTelemetry.RecordException(treatments)
	evaluationTelemetry.RecordException(treatment)
	evaluationTelemetry.RecordLatency(treatment, (1500 * time.Nanosecond).Nanoseconds())
	evaluationTelemetry.RecordLatency(treatment, (2000 * time.Nanosecond).Nanoseconds())
	evaluationTelemetry.RecordLatency(treatments, (3000 * time.Nanosecond).Nanoseconds())
	evaluationTelemetry.RecordLatency(treatments, (500 * time.Nanosecond).Nanoseconds())
	evaluationTelemetry.RecordLatency(treatmentWithConfig, (800 * time.Nanosecond).Nanoseconds())
	evaluationTelemetry.RecordLatency(treatmentsWithConfig, (1000 * time.Nanosecond).Nanoseconds())

	impressionTelemetry := NewImpressionTelemetryFacade()
	impressionTelemetry.RecordQueuedImpressions(200)
	impressionTelemetry.RecordDedupedImpressions(100)
	impressionTelemetry.RecordDroppedImpressions(50)
	impressionTelemetry.RecordQueuedImpressions(200)

	eventTelemetry := NewEventTelemetryFacade()
	eventTelemetry.RecordDroppedEvents(100)
	eventTelemetry.RecordQueuedEvents(10)

	synchronizationTelemtry := NewSynchronizationTelemetryFacade()
	synchronizationTelemtry.RecordSuccessfulSplitSync()
	time.Sleep(100 * time.Millisecond)
	synchronizationTelemtry.RecordSuccessfulSegmentSync()
	time.Sleep(100 * time.Millisecond)
	synchronizationTelemtry.RecordSuccessfulImpressionSync()
	time.Sleep(100 * time.Millisecond)
	synchronizationTelemtry.RecordSuccessfulEventsSync()
	time.Sleep(100 * time.Millisecond)
	synchronizationTelemtry.RecordSuccessfulTelemetrySync()
	time.Sleep(100 * time.Millisecond)
	synchronizationTelemtry.RecordSuccessfulTokenGet()

	httpTelemetry := NewHHTTPTelemetryFacade()
	httpTelemetry.RecordSyncError(split, 500)
	httpTelemetry.RecordSyncError(split, 500)
	httpTelemetry.RecordSyncError(split, 500)
	httpTelemetry.RecordSyncError(split, 500)
	httpTelemetry.RecordSyncError(split, 500)
	httpTelemetry.RecordSyncError(segment, 401)
	httpTelemetry.RecordSyncError(segment, 401)
	httpTelemetry.RecordSyncError(segment, 401)
	httpTelemetry.RecordSyncError(segment, 404)
	httpTelemetry.RecordSyncError(impression, 402)
	httpTelemetry.RecordSyncError(impression, 402)
	httpTelemetry.RecordSyncError(impression, 402)
	httpTelemetry.RecordSyncError(impression, 402)
	httpTelemetry.RecordSyncError(event, 400)
	httpTelemetry.RecordSyncError(telemetry, 401)
	httpTelemetry.RecordSyncError(token, 400)

	httpTelemetry.RecordSyncLatency(split, (1500 * time.Nanosecond).Nanoseconds())
	httpTelemetry.RecordSyncLatency(split, (3000 * time.Nanosecond).Nanoseconds())
	httpTelemetry.RecordSyncLatency(split, (4000 * time.Nanosecond).Nanoseconds())
	httpTelemetry.RecordSyncLatency(segment, (1500 * time.Nanosecond).Nanoseconds())
	httpTelemetry.RecordSyncLatency(segment, (1500 * time.Nanosecond).Nanoseconds())

	cacheTelemetry := NewCacheTelemetryFacade()
	cacheTelemetry.RecordSplitsCount(1000)
	cacheTelemetry.RecordSegmentsCount(50)

	pushTelemetry := NewPushTelemetryFacade()
	pushTelemetry.RecordAuthRejections()
	pushTelemetry.RecordTokenRefreshes()
	pushTelemetry.RecordAuthRejections()

	streamingTelemetry := NewStreamingTelemetryFacade()
	streamingTelemetry.RecordAblyError(40010)
	streamingTelemetry.RecordConnectionSuccess()
	streamingTelemetry.RecordPrimaryOccupancyChange(10)
	streamingTelemetry.RecordSecondaryOccupancyChange(1)
	streamingTelemetry.RecordSyncModeUpdate(0)

	sdkTelemetry := NewSDKInfoTelemetryFacade()
	sdkTelemetry.RecordSessionLength(123456789)

	miscTelemetry := NewMiscTelemetryFacade()
	miscTelemetry.AddTag("redo")
	miscTelemetry.AddTag("yaris")

	manager := NewTelemetryManager(
		factoryTelemetry,
		evaluationTelemetry,
		impressionTelemetry,
		eventTelemetry,
		synchronizationTelemtry,
		httpTelemetry,
		cacheTelemetry,
		pushTelemetry,
		streamingTelemetry,
		sdkTelemetry,
		miscTelemetry,
	)

	config := manager.BuildConfigData(conf.Default())
	data, _ := json.Marshal(config)
	if data == nil {
		t.Error("")
	}

	regular := manager.BuildUsageData()

	result, _ := json.Marshal(regular)
	if result == nil {
		t.Error("")
	}

	regular = manager.BuildUsageData()

	result, _ = json.Marshal(regular)
	if result == nil {
		t.Error("")
	}
}
