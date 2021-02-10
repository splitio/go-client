package storage

import (
	"sync"
	"sync/atomic"

	"github.com/splitio/go-client/splitio/dto"
)

type latencies struct {
	// MethodLatencies
	treatment            AtomicInt64Slice
	treatments           AtomicInt64Slice
	treatmentWithConfig  AtomicInt64Slice
	treatmentWithConfigs AtomicInt64Slice
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
	// Evaluation Counters
	treatment            int64
	treatments           int64
	treatmentWithConfig  int64
	treatmentWithConfigs int64
	track                int64

	// Push Counters
	authRejections int64
	tokenRefreshes int64

	// Factory Counters
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
	httpErrors           dto.HTTPErrors
	latencies            latencies
	records              records
	streamingEvents      []dto.StreamingEvent // Max Length 20
	mutexStreamingEvents sync.RWMutex
	tags                 []string
	mutexTags            sync.RWMutex
	factories            map[string]int64
	mutexFactories       sync.RWMutex
	integrations         []string
	mutexIntegrations    sync.RWMutex
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
	treatmentWithConfigsLatencies, err := NewAtomicInt64Slice(latencyBucketCount)
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
		httpErrors: dto.HTTPErrors{
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
			treatmentWithConfigs: treatmentWithConfigsLatencies,
			track:                track,

			splits:      splits,
			segments:    segments,
			impressions: impressions,
			events:      events,
			token:       token,
			telemetry:   telemetry,
		},
		records:              records{},
		streamingEvents:      make([]dto.StreamingEvent, 0, maxStreamingEvents),
		mutexStreamingEvents: sync.RWMutex{},
		tags:                 make([]string, 0, maxTags),
		mutexTags:            sync.RWMutex{},
		factories:            make(map[string]int64),
		mutexFactories:       sync.RWMutex{},
		integrations:         make([]string, 0, maxIntegrations),
		mutexIntegrations:    sync.RWMutex{},
	}
}

// TELEMETRY STORAGE PRODUCER

// RecordLatency stores latency for method
func (i *IMTelemetryStorage) RecordLatency(method string, bucket int) {
	switch method {
	case treatment:
		i.latencies.treatment.Incr(bucket)
	case treatments:
		i.latencies.treatments.Incr(bucket)
	case treatmentWithConfig:
		i.latencies.treatmentWithConfig.Incr(bucket)
	case treatmentsWithConfig:
		i.latencies.treatmentWithConfigs.Incr(bucket)
	case track:
		i.latencies.track.Incr(bucket)
	}
}

// RecordException stores exceptions for method
func (i *IMTelemetryStorage) RecordException(method string) {
	switch method {
	case treatment:
		atomic.AddInt64(&i.counters.treatment, 1)
	case treatments:
		atomic.AddInt64(&i.counters.treatments, 1)
	case treatmentWithConfig:
		atomic.AddInt64(&i.counters.treatmentWithConfig, 1)
	case treatmentsWithConfig:
		atomic.AddInt64(&i.counters.treatmentWithConfigs, 1)
	case track:
		atomic.AddInt64(&i.counters.track, 1)
	}
}

// RecordDroppedImpressions increments dropped impressions
func (i *IMTelemetryStorage) RecordDroppedImpressions(count int64) {
	atomic.AddInt64(&i.records.impressionsDropped, count)
}

// RecordDedupedImpressions increments deduped impressions
func (i *IMTelemetryStorage) RecordDedupedImpressions(count int64) {
	atomic.AddInt64(&i.records.impressionsDeduped, count)
}

// RecordQueuedImpressions increments queued impressions
func (i *IMTelemetryStorage) RecordQueuedImpressions(count int64) {
	atomic.AddInt64(&i.records.impressionsQueued, count)
}

// RecordDroppedEvents increments dropped events
func (i *IMTelemetryStorage) RecordDroppedEvents(count int64) {
	atomic.AddInt64(&i.records.eventsDropped, count)
}

// RecordQueuedEvents increments queued events
func (i *IMTelemetryStorage) RecordQueuedEvents(count int64) {
	atomic.AddInt64(&i.records.eventsQueued, count)
}

// RecordSuccessfulSplitSync records succesful sync
func (i *IMTelemetryStorage) RecordSuccessfulSplitSync(timestamp int64) {
	atomic.StoreInt64(&i.records.splits, timestamp)
}

// RecordSuccessfulSegmentSync records succesful sync
func (i *IMTelemetryStorage) RecordSuccessfulSegmentSync(timestamp int64) {
	atomic.StoreInt64(&i.records.segments, timestamp)
}

// RecordSuccessfulImpressionSync records succesful sync
func (i *IMTelemetryStorage) RecordSuccessfulImpressionSync(timestamp int64) {
	atomic.StoreInt64(&i.records.impressions, timestamp)
}

// RecordSuccessfulEventsSync records succesful sync
func (i *IMTelemetryStorage) RecordSuccessfulEventsSync(timestamp int64) {
	atomic.StoreInt64(&i.records.events, timestamp)
}

// RecordSuccessfulTelemetrySync records succesful sync
func (i *IMTelemetryStorage) RecordSuccessfulTelemetrySync(timestamp int64) {
	atomic.StoreInt64(&i.records.telemetry, timestamp)
}

// RecordSuccessfulTokenGet records succesful sync
func (i *IMTelemetryStorage) RecordSuccessfulTokenGet(timestamp int64) {
	atomic.StoreInt64(&i.records.token, timestamp)
}

func (i *IMTelemetryStorage) createOrUpdate(status int, item map[int]int64) {
	if item == nil {
		item[status] = 1
		return
	}
	item[status]++
}

// RecordSyncError records http error
func (i *IMTelemetryStorage) RecordSyncError(path string, status int) {
	switch path {
	case splitSync:
		i.createOrUpdate(status, i.httpErrors.Splits)
	case segmentSync:
		i.createOrUpdate(status, i.httpErrors.Segments)
	case impressionSync:
		i.createOrUpdate(status, i.httpErrors.Impressions)
	case eventSync:
		i.createOrUpdate(status, i.httpErrors.Events)
	case telemetrySync:
		i.createOrUpdate(status, i.httpErrors.Telemetry)
	case tokenSync:
		i.createOrUpdate(status, i.httpErrors.Token)
	}
}

// RecordSyncLatency records http error
func (i *IMTelemetryStorage) RecordSyncLatency(path string, bucket int) {
	switch path {
	case splitSync:
		i.latencies.splits.Incr(bucket)
	case segmentSync:
		i.latencies.segments.Incr(bucket)
	case impressionSync:
		i.latencies.impressions.Incr(bucket)
	case eventSync:
		i.latencies.events.Incr(bucket)
	case telemetrySync:
		i.latencies.telemetry.Incr(bucket)
	case tokenSync:
		i.latencies.token.Incr(bucket)
	}
}

// RecordAuthRejections records auth rejections
func (i *IMTelemetryStorage) RecordAuthRejections() {
	atomic.AddInt64(&i.counters.authRejections, 1)
}

// RecordTokenRefreshes records token
func (i *IMTelemetryStorage) RecordTokenRefreshes() {
	atomic.AddInt64(&i.counters.tokenRefreshes, 1)
}

// RecordStreamingEvent appends new streaming event
func (i *IMTelemetryStorage) RecordStreamingEvent(event dto.StreamingEvent) {
	i.mutexStreamingEvents.Lock()
	defer i.mutexStreamingEvents.Unlock()
	if len(i.streamingEvents) < maxStreamingEvents {
		i.streamingEvents = append(i.streamingEvents, event)
	}
}

// AddTag adds particular tag
func (i *IMTelemetryStorage) AddTag(tag string) {
	i.mutexTags.Lock()
	defer i.mutexTags.Unlock()
	if len(i.tags) < maxTags {
		i.tags = append(i.tags, tag)
	}
}

// RecordSessionLength records session length
func (i *IMTelemetryStorage) RecordSessionLength(session int64) {
	atomic.StoreInt64(&i.records.session, session)
}

// RecordFactory tracks factory
func (i *IMTelemetryStorage) RecordFactory(apikey string) {
	i.mutexFactories.Lock()
	defer i.mutexFactories.Unlock()
	_, ok := i.factories[apikey]
	if !ok {
		i.factories[apikey] = 1
		return
	}
	i.factories[apikey]++
}

// RecordNonReadyUsage records non ready usage
func (i *IMTelemetryStorage) RecordNonReadyUsage() {
	atomic.AddInt64(&i.counters.nonReadyUsages, 1)
}

// RecordBURTimeout records bur timeodout
func (i *IMTelemetryStorage) RecordBURTimeout() {
	atomic.AddInt64(&i.counters.burTimeouts, 1)
}

// RecordTimeUntilReady records time took until ready
func (i *IMTelemetryStorage) RecordTimeUntilReady(time int64) {
	atomic.StoreInt64(&i.records.timeUntilReady, 1)
}

// AddIntegration adds particular integration
func (i *IMTelemetryStorage) AddIntegration(integration string) {
	i.mutexIntegrations.Lock()
	defer i.mutexIntegrations.Unlock()
	if len(i.integrations) < maxIntegrations {
		i.integrations = append(i.integrations, integration)
	}
}

// TELEMETRY STORAGE CONSUMER

// PopLatencies gets and clears method latencies
func (i *IMTelemetryStorage) PopLatencies() dto.MethodLatencies {
	return dto.MethodLatencies{
		Treatment:            i.latencies.treatment.FetchAndClearAll(),
		Treatments:           i.latencies.treatments.FetchAndClearAll(),
		TreatmentWithConfig:  i.latencies.treatmentWithConfig.FetchAndClearAll(),
		TreatmentWithConfigs: i.latencies.treatmentWithConfigs.FetchAndClearAll(),
		Track:                i.latencies.track.FetchAndClearAll(),
	}
}

// PopExceptions gets and clears method exceptions
func (i *IMTelemetryStorage) PopExceptions() dto.MethodExceptions {
	return dto.MethodExceptions{
		Treatment:            atomic.SwapInt64(&i.counters.treatment, 0),
		Treatments:           atomic.SwapInt64(&i.counters.treatments, 0),
		TreatmentWithConfig:  atomic.SwapInt64(&i.counters.treatmentWithConfig, 0),
		TreatmentWithConfigs: atomic.SwapInt64(&i.counters.treatmentWithConfigs, 0),
		Track:                atomic.SwapInt64(&i.counters.track, 0),
	}
}

// GetDroppedImpressions gets total amount of impressions dropped
func (i *IMTelemetryStorage) GetDroppedImpressions() int64 {
	return atomic.LoadInt64(&i.records.impressionsDropped)
}

// GetDedupedImpressions gets total amount of impressions deduped
func (i *IMTelemetryStorage) GetDedupedImpressions() int64 {
	return atomic.LoadInt64(&i.records.impressionsDeduped)
}

// GetQueuedmpressions gets total amount of impressions queued
func (i *IMTelemetryStorage) GetQueuedmpressions() int64 {
	return atomic.LoadInt64(&i.records.impressionsQueued)
}

// GetDroppedEvents gets total amount of events dropped
func (i *IMTelemetryStorage) GetDroppedEvents() int64 {
	return atomic.LoadInt64(&i.records.eventsDropped)
}

// GetQueuedEvents gets total amount of events queued
func (i *IMTelemetryStorage) GetQueuedEvents() int64 {
	return atomic.LoadInt64(&i.records.eventsQueued)
}

// GetLastSynchronization gets last synchronization stats for fetchers and recorders
func (i *IMTelemetryStorage) GetLastSynchronization() dto.LastSynchronization {
	return dto.LastSynchronization{
		Splits:      atomic.LoadInt64(&i.records.splits),
		Segments:    atomic.LoadInt64(&i.records.segments),
		Impressions: atomic.LoadInt64(&i.records.impressions),
		Events:      atomic.LoadInt64(&i.records.events),
		Telemetry:   atomic.LoadInt64(&i.records.telemetry),
		Token:       atomic.LoadInt64(&i.records.token),
	}
}

// PopHTTPErrors gets http errors
func (i *IMTelemetryStorage) PopHTTPErrors() dto.HTTPErrors {
	toReturn := i.httpErrors
	i.httpErrors.Splits = make(map[int]int64)
	i.httpErrors.Segments = make(map[int]int64)
	i.httpErrors.Impressions = make(map[int]int64)
	i.httpErrors.Events = make(map[int]int64)
	i.httpErrors.Telemetry = make(map[int]int64)
	i.httpErrors.Token = make(map[int]int64)
	return toReturn
}

// PopHTTPLatencies gets http latencies
func (i *IMTelemetryStorage) PopHTTPLatencies() dto.HTTPLatencies {
	return dto.HTTPLatencies{
		Splits:      i.latencies.splits.FetchAndClearAll(),
		Segments:    i.latencies.segments.FetchAndClearAll(),
		Impressions: i.latencies.impressions.FetchAndClearAll(),
		Events:      i.latencies.events.FetchAndClearAll(),
		Telemetry:   i.latencies.telemetry.FetchAndClearAll(),
		Token:       i.latencies.token.FetchAndClearAll(),
	}
}

// PopAuthRejections gets total amount of auth rejections
func (i *IMTelemetryStorage) PopAuthRejections() int64 {
	return atomic.SwapInt64(&i.counters.authRejections, 0)
}

// PopTokenRefreshes gets total amount of token refreshes
func (i *IMTelemetryStorage) PopTokenRefreshes() int64 {
	return atomic.SwapInt64(&i.counters.tokenRefreshes, 0)
}

// PopStreamingEvents gets streamingEvents data
func (i *IMTelemetryStorage) PopStreamingEvents() []dto.StreamingEvent {
	i.mutexStreamingEvents.Lock()
	defer i.mutexStreamingEvents.Unlock()
	toReturn := i.streamingEvents
	i.streamingEvents = make([]dto.StreamingEvent, 0, maxStreamingEvents)
	return toReturn
}

// PopTags gets total amount of tags
func (i *IMTelemetryStorage) PopTags() []string {
	i.mutexTags.Lock()
	defer i.mutexTags.Unlock()
	toReturn := i.tags
	i.tags = make([]string, 0, maxTags)
	return toReturn
}

// GetSessionLength gets session duration
func (i *IMTelemetryStorage) GetSessionLength() int64 {
	return atomic.LoadInt64(&i.records.session)
}

// GetActiveFactories gets active factories instantiated
func (i *IMTelemetryStorage) GetActiveFactories() int64 {
	i.mutexFactories.RLock()
	defer i.mutexFactories.RUnlock()
	var toReturn int64 = 0
	for _, value := range i.factories {
		toReturn += value
	}
	return toReturn
}

// GetRedundantActiveFactories gets redundant factories
func (i *IMTelemetryStorage) GetRedundantActiveFactories() int64 {
	i.mutexFactories.RLock()
	defer i.mutexFactories.RUnlock()
	var toReturn int64 = 0
	for _, activeFactory := range i.factories {
		if activeFactory > 1 {
			toReturn += activeFactory - 1
		}
	}
	return toReturn
}

// GetNonReadyUsages gets non usages on ready
func (i *IMTelemetryStorage) GetNonReadyUsages() int64 {
	return atomic.LoadInt64(&i.counters.nonReadyUsages)
}

// GetBURTimeouts gets timedouts data
func (i *IMTelemetryStorage) GetBURTimeouts() int64 {
	return atomic.LoadInt64(&i.counters.burTimeouts)
}

// GetTimeUntilReady gets duration until ready
func (i *IMTelemetryStorage) GetTimeUntilReady() int64 {
	return atomic.LoadInt64(&i.records.timeUntilReady)
}

// GetIntegrations returns all the integrations stored
func (i *IMTelemetryStorage) GetIntegrations() []string {
	i.mutexIntegrations.Lock()
	defer i.mutexIntegrations.Unlock()
	toReturn := i.integrations
	i.integrations = make([]string, 0, maxIntegrations)
	return toReturn
}
