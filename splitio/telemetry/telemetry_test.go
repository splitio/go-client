package telemetry

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTelemetry(t *testing.T) {
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

	httpTelemetry := NewHTTPErrorTelemetryFacade()
	httpTelemetry.RecordSplitSyncErr(500)
	httpTelemetry.RecordSplitSyncErr(500)
	httpTelemetry.RecordSplitSyncErr(500)
	httpTelemetry.RecordSplitSyncErr(500)
	httpTelemetry.RecordSplitSyncErr(500)
	httpTelemetry.RecordSegmentSyncErr(401)
	httpTelemetry.RecordSegmentSyncErr(401)
	httpTelemetry.RecordSegmentSyncErr(401)
	httpTelemetry.RecordSegmentSyncErr(404)
	httpTelemetry.RecordImpressionSyncErr(402)
	httpTelemetry.RecordImpressionSyncErr(402)
	httpTelemetry.RecordImpressionSyncErr(402)
	httpTelemetry.RecordEventSyncErr(400)
	httpTelemetry.RecordEventSyncErr(401)
	httpTelemetry.RecordEventSyncErr(400)

	regular := RegularMetrics{
		MethodLatencies:      evaluationTelemetry.GetLatencies(),
		MethodExceptions:     evaluationTelemetry.GetExceptions(),
		ImpressionsDropped:   impressionTelemetry.GetDroppedImpressions(),
		ImpressionsDeduped:   impressionTelemetry.GetDedupedImpressions(),
		ImpressionsQueued:    impressionTelemetry.GetQueuedmpressions(),
		LastSynchronizations: synchronizationTelemtry.GetLastSynchronization(),
		HTTPErrors:           httpTelemetry.GetHTTPErrors(),
	}

	result, _ := json.Marshal(regular)
	if result == nil {
		t.Error("")
	}
}
