package telemetry

import (
	"sync"
	"time"
)

const (
	disabled = iota
	enabled
	paused

	requested = iota
	nonRequested

	streaming = iota
	polling

	occupancyPri       = "OCCUPANCY_PRI"
	occupancySec       = "OCCUPANCY_SEC"
	streamingStatus    = "STREAMING_STATUS"
	tokenRefresh       = "TOKEN_REFRESH"
	ablyError          = "ABLY_ERROR"
	sseConnectionError = "SSE_CONNECTION_ERROR"
	syncModeUpdate     = "SYNC_MODE_UPDATE"
)

// StreamingTelemetryFacade keeps track of streaming-related metrics
type StreamingTelemetryFacade struct {
	streamingEvents []StreamingEvent
	mutex           sync.RWMutex
}

// NewStreamingTelemetryFacade create
func NewStreamingTelemetryFacade() StreamingTelemetry {
	return &StreamingTelemetryFacade{
		streamingEvents: make([]StreamingEvent, 0),
		mutex:           sync.RWMutex{},
	}
}

// RecordPrimaryOccupancyChange records occupancy on primary
func (s *StreamingTelemetryFacade) RecordPrimaryOccupancyChange(newPublisherCount int64) {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      occupancyPri,
		Data:      newPublisherCount,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordSecondaryOccupancyChange records occupancy on secondary
func (s *StreamingTelemetryFacade) RecordSecondaryOccupancyChange(newPublisherCount int64) {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      occupancySec,
		Data:      newPublisherCount,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordConnectionSuccess records success on SSE
func (s *StreamingTelemetryFacade) RecordConnectionSuccess() {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      streamingStatus,
		Data:      enabled,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordStreamingServiceStatus records new status on streaming
func (s *StreamingTelemetryFacade) RecordStreamingServiceStatus(newStatus int) {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      streamingStatus,
		Data:      int64(newStatus),
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordTokenRefresh records next expiration event
func (s *StreamingTelemetryFacade) RecordTokenRefresh(tokenExpirationUtcTs int64) {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      tokenRefresh,
		Data:      tokenExpirationUtcTs,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordAblyError records erros in ably SSE
func (s *StreamingTelemetryFacade) RecordAblyError(statusCode int64) {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      ablyError,
		Data:      statusCode,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordNonRequestedConnectionClose records connections close
func (s *StreamingTelemetryFacade) RecordNonRequestedConnectionClose() {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      sseConnectionError,
		Data:      nonRequested,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// RecordSyncModeUpdate records updates in streaming
func (s *StreamingTelemetryFacade) RecordSyncModeUpdate(newSyncMode int64) {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.streamingEvents = append(s.streamingEvents, StreamingEvent{
		Type:      syncModeUpdate,
		Data:      newSyncMode,
		Timestamp: time.Now().UTC().Unix(),
	})
}

// GetStreamingEvents returns all the stored StreamingEvents
func (s *StreamingTelemetryFacade) GetStreamingEvents() []StreamingEvent {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	toReturn := s.streamingEvents
	s.streamingEvents = make([]StreamingEvent, 0)
	return toReturn
}
