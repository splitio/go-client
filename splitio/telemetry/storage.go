package telemetry

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	requested = iota
	nonRequested
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
)

const (
	treatment            = "getTreatment"
	treatments           = "getTreatments"
	treatmentWithConfig  = "getTreatmentWithConfig"
	treatmentsWithConfig = "getTreatmentsWithConfig"
	track                = "track"

	factoryType   = "factories"
	streamingType = "streaming"
	tagsType      = "tags"
	sessionType   = "session"

	split      = "split"
	segment    = "segment"
	impression = "impression"
	event      = "event"
	telemetry  = "telemetry"
	token      = "token"

	impressionsQueued  = "impressionsQueued"
	impressionsDeduped = "impressionsDeduped"
	impressionsDropped = "impressionsDropped"
	eventsQueued       = "eventsQueued"
	eventsDropped      = "eventsDropped"

	authRejections = "authRejections"
	tokenRefreshes = "tokenRefreshes"

	burTimeouts    = "burTimeouts"
	nonReadyUsages = "nonReadyUsages"
	timeUntilReady = "timeUntilReady"

	maxStreamingEvents = 20
	maxTags            = 10
)

type latencies struct {
	// MethodLatencies
	treatment            AtomicInt64Slice
	treatments           AtomicInt64Slice
	treatmentWithConfig  AtomicInt64Slice
	treatmentsWithConfig AtomicInt64Slice
	track                AtomicInt64Slice
	// HTTPLatencies
	splits      AtomicInt64Slice
	segments    AtomicInt64Slice
	impressions AtomicInt64Slice
	events      AtomicInt64Slice
	telemetry   AtomicInt64Slice
	token       AtomicInt64Slice
}

type counters struct {
	treatment            int64
	treatments           int64
	treatmentWithConfig  int64
	treatmentsWithConfig int64
	track                int64

	authRejections int64
	tokenRefreshes int64

	burTimeouts    int64
	nonReadyUsages int64
}

type records struct {
	// Impressions Data
	impressionsQueued  int64
	impressionsDropped int64
	impressionsDeduped int64

	// Events Data
	eventsQueued  int64
	eventsDropped int64

	// LastSynchronization
	splits      int64
	segments    int64
	impressions int64
	events      int64
	token       int64
	telemetry   int64

	// SDK
	session int64

	// Factory
	timeUntilReady int64
}

// IMTelemetryStorage In Memoty Telemetry Storage struct
type IMTelemetryStorage struct {
	counters             counters
	httpErrors           HTTPErrors
	latencies            latencies
	records              records
	streamingEvents      []StreamingEvent // Max Length 20
	mutexStreamingEvents sync.RWMutex
	tags                 []string
	mutexTags            sync.RWMutex
	factories            map[string]int64
	mutexFactories       sync.RWMutex
}

// NewIMTelemetryStorage builds in memory telemetry storage
func NewIMTelemetryStorage() TelemetryStorage {
	treatmentLatencies, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	treatmentWithConfigLatencies, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	treatmentsLatencies, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	treatmentsWithConfigLatencies, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}
	track, err := NewAtomicInt64Slice(latencyBucketCount)
	if err != nil {
		return nil
	}

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

	return &IMTelemetryStorage{
		counters: counters{},
		httpErrors: HTTPErrors{
			Splits:      make(map[int]int64),
			Segments:    make(map[int]int64),
			Impressions: make(map[int]int64),
			Events:      make(map[int]int64),
			Token:       make(map[int]int64),
			Telemetry:   make(map[int]int64),
		},
		latencies: latencies{
			treatment:            treatmentLatencies,
			treatmentWithConfig:  treatmentWithConfigLatencies,
			treatments:           treatmentsLatencies,
			treatmentsWithConfig: treatmentsWithConfigLatencies,
			track:                track,

			splits:      splits,
			segments:    segments,
			impressions: impressions,
			events:      events,
			token:       token,
			telemetry:   telemetry,
		},
		records:              records{},
		streamingEvents:      make([]StreamingEvent, 0, maxStreamingEvents),
		mutexStreamingEvents: sync.RWMutex{},
		tags:                 make([]string, 0, maxTags),
		mutexTags:            sync.RWMutex{},
		factories:            make(map[string]int64),
		mutexFactories:       sync.RWMutex{},
	}
}

////////////////////////////// PRODUCER //////////////////////////////

// AddLatency adds latencies for method
func (i *IMTelemetryStorage) AddLatency(name string, bucket int) {
	switch name {
	case treatment:
		i.latencies.treatment.Incr(bucket)
	case treatments:
		i.latencies.treatments.Incr(bucket)
	case treatmentWithConfig:
		i.latencies.treatmentWithConfig.Incr(bucket)
	case treatmentsWithConfig:
		i.latencies.treatmentsWithConfig.Incr(bucket)
	case track:
		i.latencies.track.Incr(bucket)
	case split:
		i.latencies.splits.Incr(bucket)
	case segment:
		i.latencies.segments.Incr(bucket)
	case impression:
		i.latencies.impressions.Incr(bucket)
	case event:
		i.latencies.events.Incr(bucket)
	case telemetry:
		i.latencies.telemetry.Incr(bucket)
	case token:
		i.latencies.token.Incr(bucket)
	}
}

func (i *IMTelemetryStorage) incrementHTTPCounters(path string, statusCode int, value int64) {
	switch path {
	case split:
		_, ok := i.httpErrors.Splits[statusCode]
		if !ok {
			i.httpErrors.Splits[statusCode] = value
			return
		}
		i.httpErrors.Splits[statusCode] += value
	case segment:
		_, ok := i.httpErrors.Segments[statusCode]
		if !ok {
			i.httpErrors.Segments[statusCode] = value
			return
		}
		i.httpErrors.Segments[statusCode] += value
	case impression:
		_, ok := i.httpErrors.Impressions[statusCode]
		if !ok {
			i.httpErrors.Impressions[statusCode] = value
			return
		}
		i.httpErrors.Impressions[statusCode] += value
	case event:
		_, ok := i.httpErrors.Events[statusCode]
		if !ok {
			i.httpErrors.Events[statusCode] = value
			return
		}
		i.httpErrors.Events[statusCode] += value
	case telemetry:
		_, ok := i.httpErrors.Telemetry[statusCode]
		if !ok {
			i.httpErrors.Telemetry[statusCode] = value
			return
		}
		i.httpErrors.Telemetry[statusCode] += value
	case token:
		_, ok := i.httpErrors.Token[statusCode]
		if !ok {
			i.httpErrors.Token[statusCode] = value
			return
		}
		i.httpErrors.Token[statusCode] += value
	}
}

// Increment increments counter for method
func (i *IMTelemetryStorage) Increment(name string, value int64) {
	splitted := strings.Split(name, "::")
	path := splitted[0]
	if len(splitted) == 1 {
		switch path {
		case treatment:
			atomic.AddInt64(&i.counters.treatment, value)
		case treatments:
			atomic.AddInt64(&i.counters.treatments, value)
		case treatmentWithConfig:
			atomic.AddInt64(&i.counters.treatmentWithConfig, value)
		case treatmentsWithConfig:
			atomic.AddInt64(&i.counters.treatmentsWithConfig, value)
		case track:
			atomic.AddInt64(&i.counters.track, value)
		case impressionsQueued:
			atomic.AddInt64(&i.records.impressionsQueued, value)
		case impressionsDeduped:
			atomic.AddInt64(&i.records.impressionsDeduped, value)
		case impressionsDropped:
			atomic.AddInt64(&i.records.impressionsDropped, value)
		case eventsQueued:
			atomic.AddInt64(&i.records.eventsQueued, value)
		case eventsDropped:
			atomic.AddInt64(&i.records.eventsDropped, value)
		case authRejections:
			atomic.AddInt64(&i.counters.authRejections, 1)
		case tokenRefreshes:
			atomic.AddInt64(&i.counters.tokenRefreshes, 1)
		case burTimeouts:
			atomic.AddInt64(&i.counters.burTimeouts, 1)
		case nonReadyUsages:
			atomic.AddInt64(&i.counters.nonReadyUsages, 1)
		}
		return
	}
	if len(splitted) == 2 {
		switch path {
		case factoryType:
			apikey := splitted[1]
			i.mutexFactories.Lock()
			defer i.mutexFactories.Unlock()
			_, ok := i.factories[apikey]
			if !ok {
				i.factories[apikey] = 1
				return
			}
			i.factories[apikey]++
		default:
			statusCode, err := strconv.Atoi(splitted[1])
			if err != nil {
				return
			}
			i.incrementHTTPCounters(path, statusCode, value)
			return
		}
	}
}

// PushItem adds item into list
func (i *IMTelemetryStorage) PushItem(name string, item interface{}) {
	switch name {
	case streamingType:
		i.mutexStreamingEvents.Lock()
		defer i.mutexStreamingEvents.Unlock()
		if len(i.streamingEvents) >= maxStreamingEvents {
			return
		}
		streamingEvent, ok := item.(StreamingEvent)
		if !ok {
			return
		}
		i.streamingEvents = append(i.streamingEvents, streamingEvent)
	case tagsType:
		i.mutexTags.Lock()
		defer i.mutexTags.Unlock()
		if len(i.tags) >= maxTags {
			return
		}
		tag, ok := item.(string)
		if !ok {
			return
		}
		i.tags = append(i.tags, tag)
	}
}

// Set implementation
func (i *IMTelemetryStorage) Set(name string, value int64) {
	switch name {
	case split:
		atomic.StoreInt64(&i.records.splits, value)
	case segment:
		atomic.StoreInt64(&i.records.segments, value)
	case impression:
		atomic.StoreInt64(&i.records.impressions, value)
	case event:
		atomic.StoreInt64(&i.records.events, value)
	case telemetry:
		atomic.StoreInt64(&i.records.telemetry, value)
	case token:
		atomic.StoreInt64(&i.records.token, value)
	case sessionType:
		atomic.StoreInt64(&i.records.session, value)
	case timeUntilReady:
		atomic.AddInt64(&i.records.timeUntilReady, value)
	}
}

////////////////////////////// CONSUMER //////////////////////////////

// GetFactories returns factories
func (i *IMTelemetryStorage) GetFactories() map[string]int64 {
	i.mutexFactories.RLock()
	defer i.mutexFactories.RUnlock()
	return i.factories
}

// GetRecord gets record
func (i *IMTelemetryStorage) GetRecord(name string) int64 {
	switch name {
	case impressionsQueued:
		return atomic.LoadInt64(&i.records.impressionsQueued)
	case impressionsDeduped:
		return atomic.LoadInt64(&i.records.impressionsDeduped)
	case impressionsDropped:
		return atomic.LoadInt64(&i.records.impressionsDropped)
	case eventsQueued:
		return atomic.LoadInt64(&i.records.eventsQueued)
	case eventsDropped:
		return atomic.LoadInt64(&i.records.eventsDropped)
	case sessionType:
		return atomic.LoadInt64(&i.records.session)
	case timeUntilReady:
		return atomic.LoadInt64(&i.records.timeUntilReady)
	case split:
		return atomic.LoadInt64(&i.records.splits)
	case segment:
		return atomic.LoadInt64(&i.records.segments)
	case impression:
		return atomic.LoadInt64(&i.records.impressions)
	case event:
		return atomic.LoadInt64(&i.records.events)
	case telemetry:
		return atomic.LoadInt64(&i.records.telemetry)
	case token:
		return atomic.LoadInt64(&i.records.token)
	}
	return 0
}

// PopCounter returns exceptions
func (i *IMTelemetryStorage) PopCounter(name string) int64 {
	switch name {
	case treatment:
		return atomic.SwapInt64(&i.counters.treatment, 0)
	case treatments:
		return atomic.SwapInt64(&i.counters.treatments, 0)
	case treatmentWithConfig:
		return atomic.SwapInt64(&i.counters.treatmentWithConfig, 0)
	case treatmentsWithConfig:
		return atomic.SwapInt64(&i.counters.treatmentsWithConfig, 0)
	case track:
		return atomic.SwapInt64(&i.counters.track, 0)
	case authRejections:
		return atomic.SwapInt64(&i.counters.authRejections, 0)
	case tokenRefreshes:
		return atomic.SwapInt64(&i.counters.tokenRefreshes, 0)
	case burTimeouts:
		return atomic.SwapInt64(&i.counters.burTimeouts, 0)
	case nonReadyUsages:
		return atomic.SwapInt64(&i.counters.nonReadyUsages, 0)
	}
	return 0
}

// PopHTTPErrors returns errors stored
func (i *IMTelemetryStorage) PopHTTPErrors() HTTPErrors {
	toReturn := i.httpErrors
	i.httpErrors.Splits = make(map[int]int64)
	i.httpErrors.Segments = make(map[int]int64)
	i.httpErrors.Impressions = make(map[int]int64)
	i.httpErrors.Events = make(map[int]int64)
	i.httpErrors.Telemetry = make(map[int]int64)
	i.httpErrors.Token = make(map[int]int64)
	return toReturn
}

// PopItems returns items
func (i *IMTelemetryStorage) PopItems(name string) interface{} {
	switch name {
	case streamingType:
		toReturn := i.streamingEvents
		i.streamingEvents = make([]StreamingEvent, 0, maxStreamingEvents)
		return toReturn
	case tagsType:
		toReturn := i.tags
		i.tags = make([]string, 0, maxTags)
		return toReturn
	}
	return nil
}

// PopLatency returns latency
func (i *IMTelemetryStorage) PopLatency(name string) []int64 {
	switch name {
	case treatment:
		return i.latencies.treatment.FetchAndClearAll()
	case treatments:
		return i.latencies.treatments.FetchAndClearAll()
	case treatmentWithConfig:
		return i.latencies.treatmentWithConfig.FetchAndClearAll()
	case treatmentsWithConfig:
		return i.latencies.treatmentsWithConfig.FetchAndClearAll()
	case track:
		return i.latencies.track.FetchAndClearAll()
	case split:
		return i.latencies.splits.FetchAndClearAll()
	case segment:
		return i.latencies.segments.FetchAndClearAll()
	case impression:
		return i.latencies.impressions.FetchAndClearAll()
	case event:
		return i.latencies.events.FetchAndClearAll()
	case telemetry:
		return i.latencies.telemetry.FetchAndClearAll()
	case token:
		return i.latencies.token.FetchAndClearAll()
	}
	return []int64{}
}
