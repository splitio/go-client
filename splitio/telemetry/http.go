package telemetry

import "sync"

// HTTPErrorTelemetryFacade struct
type HTTPErrorTelemetryFacade struct {
	splits      map[int]int64
	segments    map[int]int64
	impressions map[int]int64
	events      map[int]int64
	telemetry   map[int]int64
	token       map[int]int64
	mutex       sync.RWMutex
}

// NewHTTPErrorTelemetryFacade new facade
func NewHTTPErrorTelemetryFacade() HTTPErrorTelemetry {
	return &HTTPErrorTelemetryFacade{
		splits:      make(map[int]int64, 0),
		segments:    make(map[int]int64, 0),
		impressions: make(map[int]int64, 0),
		events:      make(map[int]int64, 0),
		telemetry:   make(map[int]int64, 0),
		token:       make(map[int]int64, 0),
		mutex:       sync.RWMutex{},
	}
}

// RecordSplitSyncErr records error in split
func (h *HTTPErrorTelemetryFacade) RecordSplitSyncErr(status int) {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	_, ok := h.splits[status]
	if !ok {
		h.splits[status] = 1
		return
	}
	h.splits[status]++
}

// RecordSegmentSyncErr records error in segment
func (h *HTTPErrorTelemetryFacade) RecordSegmentSyncErr(status int) {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	_, ok := h.segments[status]
	if !ok {
		h.segments[status] = 1
		return
	}
	h.segments[status]++
}

// RecordImpressionSyncErr records error in impression
func (h *HTTPErrorTelemetryFacade) RecordImpressionSyncErr(status int) {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	_, ok := h.impressions[status]
	if !ok {
		h.impressions[status] = 1
		return
	}
	h.impressions[status]++
}

// RecordEventSyncErr records error in event
func (h *HTTPErrorTelemetryFacade) RecordEventSyncErr(status int) {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	_, ok := h.events[status]
	if !ok {
		h.events[status] = 1
		return
	}
	h.events[status]++
}

// RecordTelemetrySyncErr records error in telemetry
func (h *HTTPErrorTelemetryFacade) RecordTelemetrySyncErr(status int) {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	_, ok := h.telemetry[status]
	if !ok {
		h.telemetry[status] = 1
		return
	}
	h.telemetry[status]++
}

// RecordTokenGetErr records error in auth
func (h *HTTPErrorTelemetryFacade) RecordTokenGetErr(status int) {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	_, ok := h.token[status]
	if !ok {
		h.token[status] = 1
		return
	}
	h.token[status]++
}

// PopHTTPErrors returns errors stored
func (h *HTTPErrorTelemetryFacade) PopHTTPErrors() HTTPErrors {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	toReturn := HTTPErrors{
		Splits:      h.splits,
		Segments:    h.segments,
		Impressions: h.impressions,
		Events:      h.events,
		Telemetry:   h.telemetry,
		Token:       h.token,
	}
	h.splits = make(map[int]int64, 0)
	h.segments = make(map[int]int64, 0)
	h.impressions = make(map[int]int64, 0)
	h.events = make(map[int]int64, 0)
	h.telemetry = make(map[int]int64, 0)
	h.token = make(map[int]int64, 0)
	return toReturn
}
