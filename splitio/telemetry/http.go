package telemetry

import (
	"sync"

	"github.com/splitio/go-split-commons/util"
)

const (
	split      = "split"
	segment    = "segment"
	impression = "impression"
	event      = "event"
	telemetry  = "telemetry"
	token      = "token"
)

type httpLatencies struct {
	splits      AtomicInt64Slice
	segments    AtomicInt64Slice
	impressions AtomicInt64Slice
	events      AtomicInt64Slice
	telemetry   AtomicInt64Slice
	token       AtomicInt64Slice
}

// HTTPTelemetryFacade struct
type HTTPTelemetryFacade struct {
	httpLatencies  httpLatencies
	httpErrors     HTTPErrors
	mutexLatencies sync.RWMutex
	mutexErrors    sync.RWMutex
}

// NewHHTTPTelemetryFacade new facade
func NewHHTTPTelemetryFacade() HTTPTelemetry {
	splits, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	segments, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	impressions, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	events, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	telemetry, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	token, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}

	return &HTTPTelemetryFacade{
		httpLatencies: httpLatencies{
			splits:      splits,
			segments:    segments,
			impressions: impressions,
			events:      events,
			token:       token,
			telemetry:   telemetry,
		},
		httpErrors: HTTPErrors{
			Splits:      make(map[int]int64),
			Segments:    make(map[int]int64),
			Impressions: make(map[int]int64),
			Events:      make(map[int]int64),
			Token:       make(map[int]int64),
			Telemetry:   make(map[int]int64),
		},
		mutexLatencies: sync.RWMutex{},
		mutexErrors:    sync.RWMutex{},
	}
}

// RecordSyncError records error
func (h *HTTPTelemetryFacade) RecordSyncError(method string, status int) {
	h.mutexErrors.Lock()
	defer h.mutexErrors.Unlock()
	switch method {
	case split:
		_, ok := h.httpErrors.Splits[status]
		if !ok {
			h.httpErrors.Splits[status] = 1
			return
		}
		h.httpErrors.Splits[status]++
	case segment:
		_, ok := h.httpErrors.Segments[status]
		if !ok {
			h.httpErrors.Segments[status] = 1
			return
		}
		h.httpErrors.Segments[status]++
	case impression:
		_, ok := h.httpErrors.Impressions[status]
		if !ok {
			h.httpErrors.Impressions[status] = 1
			return
		}
		h.httpErrors.Impressions[status]++
	case event:
		_, ok := h.httpErrors.Events[status]
		if !ok {
			h.httpErrors.Events[status] = 1
			return
		}
		h.httpErrors.Events[status]++
	case telemetry:
		_, ok := h.httpErrors.Telemetry[status]
		if !ok {
			h.httpErrors.Telemetry[status] = 1
			return
		}
		h.httpErrors.Telemetry[status]++
	case token:
		_, ok := h.httpErrors.Token[status]
		if !ok {
			h.httpErrors.Token[status] = 1
			return
		}
		h.httpErrors.Token[status]++
	}
}

// RecordSyncLatency records latencies
func (h *HTTPTelemetryFacade) RecordSyncLatency(method string, latency int64) {
	bucket := util.Bucket(latency)
	h.mutexLatencies.Lock()
	defer h.mutexLatencies.Unlock()
	switch method {
	case split:
		h.httpLatencies.splits.Incr(bucket)
	case segment:
		h.httpLatencies.segments.Incr(bucket)
	case impression:
		h.httpLatencies.impressions.Incr(bucket)
	case event:
		h.httpLatencies.events.Incr(bucket)
	case telemetry:
		h.httpLatencies.telemetry.Incr(bucket)
	case token:
		h.httpLatencies.token.Incr(bucket)
	}
}

// PopHTTPErrors returns errors stored
func (h *HTTPTelemetryFacade) PopHTTPErrors() HTTPErrors {
	h.mutexErrors.Lock()
	defer h.mutexErrors.Unlock()
	toReturn := h.httpErrors
	h.httpErrors.Splits = make(map[int]int64)
	h.httpErrors.Segments = make(map[int]int64)
	h.httpErrors.Impressions = make(map[int]int64)
	h.httpErrors.Events = make(map[int]int64)
	h.httpErrors.Telemetry = make(map[int]int64)
	h.httpErrors.Token = make(map[int]int64)
	return toReturn
}

// PopHTTPLatencies returns latencies stored
func (h *HTTPTelemetryFacade) PopHTTPLatencies() HTTPLatencies {
	h.mutexLatencies.Lock()
	defer h.mutexLatencies.Unlock()
	toReturn := HTTPLatencies{
		Splits:      h.httpLatencies.splits.FetchAndClearAll(),
		Segments:    h.httpLatencies.segments.FetchAndClearAll(),
		Impressions: h.httpLatencies.impressions.FetchAndClearAll(),
		Events:      h.httpLatencies.events.FetchAndClearAll(),
		Token:       h.httpLatencies.token.FetchAndClearAll(),
		Telemetry:   h.httpLatencies.telemetry.FetchAndClearAll(),
	}
	return toReturn
}
