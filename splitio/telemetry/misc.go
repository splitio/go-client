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
	defer m.mutex.Unlock()
	m.mutex.Lock()
	if len(m.tags) < 10 {
		m.tags = append(m.tags, tag)
	}
}

// GetTags returns all the tags stored
func (m *MiscTelemetryFacade) GetTags() []string {
	defer m.mutex.Unlock()
	m.mutex.Lock()
	toReturn := m.tags
	m.tags = make([]string, 0, 10)
	return toReturn
}
