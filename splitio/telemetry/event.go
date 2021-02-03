package telemetry

// EventTelemetryFacade keeps track of event-related metrics
type EventTelemetryFacade struct {
	eventsQueued  int64
	eventsDropped int64
}

// NewEventTelemetryFacade facade for EventTelemetry
func NewEventTelemetryFacade() EventTelemetry {
	return &EventTelemetryFacade{
		eventsQueued:  0,
		eventsDropped: 0,
	}
}

// RecordDroppedEvents increments dropped events
func (e *EventTelemetryFacade) RecordDroppedEvents(count int64) {
	e.eventsDropped += count
}

// RecordQueuedEvents increments queued events
func (e *EventTelemetryFacade) RecordQueuedEvents(count int64) {
	e.eventsQueued += count
}

// GetDroppedEvents returns dropped events
func (e *EventTelemetryFacade) GetDroppedEvents() int64 {
	return e.eventsDropped
}

// GetQueuedEvents returns queued events
func (e *EventTelemetryFacade) GetQueuedEvents() int64 {
	return e.eventsQueued
}
