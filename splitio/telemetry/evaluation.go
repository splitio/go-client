package telemetry

// EvaluationTelemetryFacade keeps track of evaluation-related metrics
type EvaluationTelemetryFacade struct {
	treatmentLatencies            AtomicInt64Slice
	treatmentsLatencies           AtomicInt64Slice
	treatmentWithConfigLatencies  AtomicInt64Slice
	treatmentsWithConfigLatencies AtomicInt64Slice
	exceptions                    AtomicInt64Slice
}

func NewEvaluationTelemetryFacade() *EvaluationTelemetryFacade {
	return &EvaluationTelemetryFacade{
		treatmentLatencies:            NewAtomicInt64Slice(LatencyBucketCount),
		treatmentWithConfigLatencies:  NewAtomicInt64Slice(LatencyBucketCount),
		treatmentsLatencies:           NewAtomicInt64Slice(LatencyBucketCount),
		treatmentsWithConfigLatencies: NewAtomicInt64Slice(LatencyBucketCount),
		exceptions:                    NewAtomicInt64Slice(TotalMethodCount),
	}
}
