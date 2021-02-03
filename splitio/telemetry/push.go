package telemetry

import "sync/atomic"

// PushTelemetryFacade keeps track of push-related metrics
type PushTelemetryFacade struct {
	authRejections int64
	tokenRefreshes int64
}

// NewPushTelemetryFacade builds new facade
func NewPushTelemetryFacade() PushTelemetry {
	return &PushTelemetryFacade{
		authRejections: 0,
		tokenRefreshes: 0,
	}
}

// RecordAuthRejections records auth rejection
func (p *PushTelemetryFacade) RecordAuthRejections() {
	atomic.AddInt64(&p.authRejections, 1)
}

// RecordTokenRefreshes records token refresh
func (p *PushTelemetryFacade) RecordTokenRefreshes() {
	atomic.AddInt64(&p.tokenRefreshes, 1)
}

// GetAuthRejections returns all the rejections
func (p *PushTelemetryFacade) GetAuthRejections() int64 { return atomic.SwapInt64(&p.authRejections, 0) }

// GetTokenRefreshes returns all the refreshes made
func (p *PushTelemetryFacade) GetTokenRefreshes() int64 { return atomic.SwapInt64(&p.tokenRefreshes, 0) }
