package telemetry

import "sync/atomic"

// SDKInfoTelemetryFacade keeps track of sdk-related metrics
type SDKInfoTelemetryFacade struct {
	session int64
}

// NewSDKInfoTelemetryFacade create
func NewSDKInfoTelemetryFacade() SDKInfoTelemetry {
	return &SDKInfoTelemetryFacade{
		session: 0,
	}
}

// RecordSessionLength stores session duration
func (s *SDKInfoTelemetryFacade) RecordSessionLength(session int64) {
	atomic.AddInt64(&s.session, session)
}

// PopSessionLength returns stored session
func (s *SDKInfoTelemetryFacade) PopSessionLength() int64 {
	return atomic.SwapInt64(&s.session, 0)
}
