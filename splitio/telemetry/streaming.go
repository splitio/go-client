package telemetry

import (
	"sync"
	"time"
)

const (
	requested = iota
	nonRequested
)

const (
	disabled = iota
	enabled
	paused
)

const (
	eventTypeSSEConnectionEstablished = iota * 10
	eventTypeOccupancyPri
	eventTypeOccupancySec
	eventTypeStreamingStatus
	eventTypeConnectionError
	eventTypeTokenRefresh
	eventTypeAblyError
	eventTypeSyncMode

	maxLength = 20
)

// StreamingTelemetryFacade keeps track of streaming-related metrics
type StreamingTelemetryFacade struct {
	streamingEvents []StreamingEvent // Max Length 20
	mutex           sync.RWMutex
}

// NewStreamingTelemetryFacade create
func NewStreamingTelemetryFacade() StreamingTelemetry {
	return &StreamingTelemetryFacade{
		streamingEvents: make([]StreamingEvent, 0, maxLength),
		mutex:           sync.RWMutex{},
	}
}

func (s *StreamingTelemetryFacade) canAdd() bool {
	defer s.mutex.RUnlock()
	s.mutex.RLock()
	if len(s.streamingEvents) < maxLength {
		return true
	}
	return false
}

// RecordPrimaryOccupancyChange records occupancy on primary
func (s *StreamingTelemetryFacade) RecordPrimaryOccupancyChange(newPublisherCount int64) {
	if !s.canAdd() {
		return
	}
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      eventTypeOccupancyPri,
		Data:      newPublisherCount,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordSecondaryOccupancyChange records occupancy on secondary
func (s *StreamingTelemetryFacade) RecordSecondaryOccupancyChange(newPublisherCount int64) {
	if !s.canAdd() {
		return
	}
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      eventTypeOccupancySec,
		Data:      newPublisherCount,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordConnectionSuccess records success on SSE
func (s *StreamingTelemetryFacade) RecordConnectionSuccess() {
	if !s.canAdd() {
		return
	}
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      eventTypeSSEConnectionEstablished,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordStreamingServiceStatus records new status on streaming
func (s *StreamingTelemetryFacade) RecordStreamingServiceStatus(newStatus int) {
	if !s.canAdd() {
		return
	}
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      eventTypeStreamingStatus,
		Data:      int64(newStatus),
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordTokenRefresh records next expiration event
func (s *StreamingTelemetryFacade) RecordTokenRefresh(tokenExpirationUtcTs int64) {
	if !s.canAdd() {
		return
	}
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      eventTypeTokenRefresh,
		Data:      tokenExpirationUtcTs,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordAblyError records erros in ably SSE
func (s *StreamingTelemetryFacade) RecordAblyError(statusCode int64) {
	if !s.canAdd() {
		return
	}
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      eventTypeAblyError,
		Data:      statusCode,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordNonRequestedConnectionClose records connections close
func (s *StreamingTelemetryFacade) RecordNonRequestedConnectionClose() {
	if !s.canAdd() {
		return
	}
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      eventTypeConnectionError,
		Data:      nonRequested,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordSyncModeUpdate records updates in streaming
func (s *StreamingTelemetryFacade) RecordSyncModeUpdate(newSyncMode int64) {
	if !s.canAdd() {
		return
	}
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      eventTypeSyncMode,
		Data:      newSyncMode,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// PopStreamingEvents returns all the stored StreamingEvents
func (s *StreamingTelemetryFacade) PopStreamingEvents() []StreamingEvent {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	toReturn := s.streamingEvents
	s.streamingEvents = make([]StreamingEvent, 0, 20)
	return toReturn
}
