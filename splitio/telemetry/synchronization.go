package telemetry

import (
	"sync/atomic"
	"time"
)

// SynchronizationTelemetryFacade keeps track of synchronization-related metrics
type SynchronizationTelemetryFacade struct {
	lastSynchronization LastSynchronization
}

// NewSynchronizationTelemetryFacade facade for SynchronizationTelemetry
func NewSynchronizationTelemetryFacade() SynchronizationTelemetry {
	return &SynchronizationTelemetryFacade{
		lastSynchronization: LastSynchronization{},
	}
}

// RecordSuccessfulSplitSync records split sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulSplitSync() {
	atomic.AddInt64(&s.lastSynchronization.Splits, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulSegmentSync records segment sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulSegmentSync() {
	atomic.AddInt64(&s.lastSynchronization.Segments, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulImpressionSync records impression sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulImpressionSync() {
	atomic.AddInt64(&s.lastSynchronization.Impressions, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulEventsSync records event sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulEventsSync() {
	atomic.AddInt64(&s.lastSynchronization.Events, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulTelemetrySync records telemetry sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulTelemetrySync() {
	atomic.AddInt64(&s.lastSynchronization.Telemetry, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulTokenGet records token sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulTokenGet() {
	atomic.AddInt64(&s.lastSynchronization.Token, time.Now().UTC().UnixNano()/1000000)
}

// GetLastSynchronization gets last sync records
func (s *SynchronizationTelemetryFacade) GetLastSynchronization() LastSynchronization {
	return s.lastSynchronization
}
