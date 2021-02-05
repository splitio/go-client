package telemetry

import (
	"sync/atomic"

	"github.com/splitio/go-split-commons/util"
)

// Method constants
const (
	treatment            = "getTreatment"
	treatments           = "getTreatments"
	treatmentWithConfig  = "getTreatmentWithConfig"
	treatmentsWithConfig = "getTreatmentsWithConfig"
	track                = "track"
)

type methodLatencies struct {
	treatmentLatencies            AtomicInt64Slice
	treatmentsLatencies           AtomicInt64Slice
	treatmentWithConfigLatencies  AtomicInt64Slice
	treatmentsWithConfigLatencies AtomicInt64Slice
	track                         AtomicInt64Slice
}

// EvaluationTelemetryFacade keeps track of evaluation-related metrics
type EvaluationTelemetryFacade struct {
	methodLatencies  methodLatencies
	methodExceptions MethodExceptions
}

// NewEvaluationTelemetryFacade facade for Evaluation telemetry
func NewEvaluationTelemetryFacade() EvaluationTelemetry {
	treatmentLatencies, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	treatmentWithConfigLatencies, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	treatmentsLatencies, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	treatmentsWithConfigLatencies, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	track, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	return &EvaluationTelemetryFacade{
		methodLatencies: methodLatencies{
			treatmentLatencies:            treatmentLatencies,
			treatmentWithConfigLatencies:  treatmentWithConfigLatencies,
			treatmentsLatencies:           treatmentsLatencies,
			treatmentsWithConfigLatencies: treatmentsWithConfigLatencies,
			track:                         track,
		},
		methodExceptions: MethodExceptions{},
	}
}

// RecordException records exceptions
func (e *EvaluationTelemetryFacade) RecordException(method string) {
	switch method {
	case treatment:
		atomic.AddInt64(&e.methodExceptions.Treatment, 1)
	case treatments:
		atomic.AddInt64(&e.methodExceptions.Treatments, 1)
	case treatmentWithConfig:
		atomic.AddInt64(&e.methodExceptions.TreatmentWithConfig, 1)
	case treatmentsWithConfig:
		atomic.AddInt64(&e.methodExceptions.TreatmentWithConfigs, 1)
	case track:
		atomic.AddInt64(&e.methodExceptions.Track, 1)
	}
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
	case track:
		e.methodLatencies.track.Incr(bucket)
	}
}

// PopLatencies returns latencies
func (e *EvaluationTelemetryFacade) PopLatencies() MethodLatencies {
	return MethodLatencies{
		Treatment:            e.methodLatencies.treatmentLatencies.FetchAndClearAll(),
		Treatments:           e.methodLatencies.treatmentsLatencies.FetchAndClearAll(),
		TreatmentWithConfig:  e.methodLatencies.treatmentWithConfigLatencies.FetchAndClearAll(),
		TreatmentWithConfigs: e.methodLatencies.treatmentsWithConfigLatencies.FetchAndClearAll(),
		Track:                e.methodLatencies.track.FetchAndClearAll(),
	}
}

// PopExceptions returns exceptions
func (e *EvaluationTelemetryFacade) PopExceptions() MethodExceptions {
	return MethodExceptions{
		Treatment:            atomic.SwapInt64(&e.methodExceptions.Treatment, 0),
		Treatments:           atomic.SwapInt64(&e.methodExceptions.Treatments, 0),
		TreatmentWithConfig:  atomic.SwapInt64(&e.methodExceptions.TreatmentWithConfig, 0),
		TreatmentWithConfigs: atomic.SwapInt64(&e.methodExceptions.TreatmentWithConfigs, 0),
		Track:                atomic.SwapInt64(&e.methodExceptions.Track, 0),
	}
}
