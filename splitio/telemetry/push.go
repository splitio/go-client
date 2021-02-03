package telemetry

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
	p.authRejections++
}

// RecordTokenRefreshes records token refresh
func (p *PushTelemetryFacade) RecordTokenRefreshes() {
	p.tokenRefreshes++
}

// GetAuthRejections returns all the rejections
func (p *PushTelemetryFacade) GetAuthRejections() int64 { return p.authRejections }

// GetTokenRefreshes returns all the refreshes made
func (p *PushTelemetryFacade) GetTokenRefreshes() int64 { return p.tokenRefreshes }
