package telemetry

import "sync"

// MiscTelemetryFacade keeps track of misc-related metrics
type MiscTelemetryFacade struct {
	mutex sync.RWMutex
	tags  []string
}

// NewMiscTelemetryFacade create
func NewMiscTelemetryFacade() MiscTelemetry {
	return &MiscTelemetryFacade{
		mutex: sync.RWMutex{},
		tags:  make([]string, 0, 10),
	}
}

// AddTag add tags to misc data
func (m *MiscTelemetryFacade) AddTag(tag string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if len(m.tags) < 10 {
		m.tags = append(m.tags, tag)
	}
}

// PopTags returns all the tags stored
func (m *MiscTelemetryFacade) PopTags() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	toReturn := m.tags
	m.tags = make([]string, 0, 10)
	return toReturn
}
