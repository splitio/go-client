package telemetry

import (
	"github.com/splitio/go-split-commons/util"
)

// Method constants
const (
	MethodTreatment = iota
	MethodTreatments
	MethodTreatmentWithConfig
	MethodTreatmentsWithConfig
	MethodTrack

	TotalMethodCount   = 5
	LatencyBucketCount = 20

	treatment            = "getTreatment"
	treatments           = "getTreatments"
	treatmentWithConfig  = "getTreatmentWithConfig"
	treatmentsWithConfig = "getTreatmentsWithConfig"
	track                = "track"
)

// MethodLatencies struct
type MethodLatencies struct {
	treatmentLatencies            AtomicInt64Slice
	treatmentsLatencies           AtomicInt64Slice
	treatmentWithConfigLatencies  AtomicInt64Slice
	treatmentsWithConfigLatencies AtomicInt64Slice
}

// EvaluationTelemetryFacade keeps track of evaluation-related metrics
type EvaluationTelemetryFacade struct {
	methodLatencies  MethodLatencies
	methodExceptions AtomicInt64Slice
}

// NewEvaluationTelemetryFacade facade for Evaluation telemetry
func NewEvaluationTelemetryFacade() EvaluationTelemetry {
	treatmentLatencies, err := NewAtomicInt64Slice(LatencyBucketCount)
	if err != nil {
		return nil
	}
	treatmentWithConfigLatencies, err := NewAtomicInt64Slice(LatencyBucketCount)
	if err != nil {
		return nil
	}
	treatmentsLatencies, err := NewAtomicInt64Slice(LatencyBucketCount)
	if err != nil {
		return nil
	}
	treatmentsWithConfigLatencies, err := NewAtomicInt64Slice(LatencyBucketCount)
	if err != nil {
		return nil
	}
	exceptions, err := NewAtomicInt64Slice(TotalMethodCount)
	if err != nil {
		return nil
	}
	return &EvaluationTelemetryFacade{
		methodLatencies: MethodLatencies{
			treatmentLatencies:            treatmentLatencies,
			treatmentWithConfigLatencies:  treatmentWithConfigLatencies,
			treatmentsLatencies:           treatmentsLatencies,
			treatmentsWithConfigLatencies: treatmentsWithConfigLatencies,
		},
		methodExceptions: exceptions,
	}
}

func getMethod(method string) int {
	switch method {
	case treatment:
		return MethodTreatment
	case treatments:
		return MethodTreatments
	case treatmentWithConfig:
		return MethodTreatmentWithConfig
	case treatmentsWithConfig:
		return MethodTreatmentsWithConfig
	}
	return -1
}

// RecordException records exceptions
func (e *EvaluationTelemetryFacade) RecordException(method string) {
	e.methodExceptions.Incr(getMethod(method))
}

// RecordLatency records latencies for method
func (e *EvaluationTelemetryFacade) RecordLatency(method string, latency int64) {
	bucket := util.Bucket(latency)
	switch method {
	case treatment:
		e.methodLatencies.treatmentLatencies.Incr(bucket)
	case treatments:
		e.methodLatencies.treatmentsLatencies.Incr(bucket)
	case treatmentWithConfig:
		e.methodLatencies.treatmentWithConfigLatencies.Incr(bucket)
	case treatmentsWithConfig:
		e.methodLatencies.treatmentsWithConfigLatencies.Incr(bucket)
	}
}

// GetLatencies returns latencies
func (e *EvaluationTelemetryFacade) GetLatencies() map[string][]int64 {
	toReturn := make(map[string][]int64, 0)
	toReturn[treatment] = e.methodLatencies.treatmentLatencies.FetchAndClearAll()
	toReturn[treatments] = e.methodLatencies.treatmentsLatencies.FetchAndClearAll()
	toReturn[treatmentWithConfig] = e.methodLatencies.treatmentWithConfigLatencies.FetchAndClearAll()
	toReturn[treatmentsWithConfig] = e.methodLatencies.treatmentsWithConfigLatencies.FetchAndClearAll()
	return toReturn
}

// GetExceptions returns exceptions
func (e *EvaluationTelemetryFacade) GetExceptions() map[string]int64 {
	methodExceptions := e.methodExceptions.FetchAndClearAll()
	toReturn := make(map[string]int64, 0)
	toReturn[treatment] = methodExceptions[MethodTreatment]
	toReturn[treatments] = methodExceptions[MethodTreatments]
	toReturn[treatmentWithConfig] = methodExceptions[MethodTreatmentWithConfig]
	toReturn[treatmentsWithConfig] = methodExceptions[MethodTreatmentsWithConfig]
	toReturn[track] = methodExceptions[MethodTrack]
	return toReturn
}
