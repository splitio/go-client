package telemetry

import (
	"sync/atomic"
	"time"
)

// SynchronizationTelemetryFacade keeps track of synchronization-related metrics
type SynchronizationTelemetryFacade struct {
	lastSplitSync      int64
	lastSegmentSync    int64
	lastImpressionSync int64
	lastEventSync      int64
	lastTelemetrySync  int64
	lastTokenGet       int64
}

// NewSynchronizationTelemetryFacade facade for SynchronizationTelemetry
func NewSynchronizationTelemetryFacade() SynchronizationTelemetry {
	return &SynchronizationTelemetryFacade{
		lastSplitSync:      0,
		lastSegmentSync:    0,
		lastImpressionSync: 0,
		lastEventSync:      0,
		lastTelemetrySync:  0,
		lastTokenGet:       0,
	}
}

// RecordSuccessfulSplitSync records split sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulSplitSync() {
	atomic.AddInt64(&s.lastSplitSync, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulSegmentSync records segment sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulSegmentSync() {
	atomic.AddInt64(&s.lastSegmentSync, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulImpressionSync records impression sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulImpressionSync() {
	atomic.AddInt64(&s.lastImpressionSync, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulEventsSync records event sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulEventsSync() {
	atomic.AddInt64(&s.lastEventSync, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulTelemetrySync records telemetry sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulTelemetrySync() {
	atomic.AddInt64(&s.lastTelemetrySync, time.Now().UTC().UnixNano()/1000000)
}

// RecordSuccessfulTokenGet records token sync
func (s *SynchronizationTelemetryFacade) RecordSuccessfulTokenGet() {
	atomic.AddInt64(&s.lastTokenGet, time.Now().UTC().UnixNano()/1000000)
}

// GetLastSynchronization gets last sync records
func (s *SynchronizationTelemetryFacade) GetLastSynchronization() LastSynchronization {
	return LastSynchronization{
		Splits:      atomic.LoadInt64(&s.lastSplitSync),
		Segments:    atomic.LoadInt64(&s.lastSegmentSync),
		Impressions: atomic.LoadInt64(&s.lastImpressionSync),
		Events:      atomic.LoadInt64(&s.lastEventSync),
		Telemetry:   atomic.LoadInt64(&s.lastTelemetrySync),
		Token:       atomic.LoadInt64(&s.lastTokenGet),
	}
}
