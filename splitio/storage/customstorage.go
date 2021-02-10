package storage

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/splitio/go-client/splitio/dto"
)

const (
	methodKey             = "method"
	dataKey               = "data"
	impressionsDroppedKey = "impressionsdropped"
	impressionsDedupedKey = "impressionsdeduped"
	impressionsQueuedKey  = "impressionsqueued"
	eventsDroppedKey      = "eventsdropped"
	eventsQueuedKey       = "eventsqueued"
	synchronizationKey    = "synchronization"
	httpKey               = "http"
	pushKey               = "push"
	authRejectionsKey     = "authrejections"
	tokenRefreshesKey     = "tokenrefreshes"
	streamingKey          = "streaming"
	tagKey                = "tags"
	sdkKey                = "sdk"
	sessionKey            = "session"
	factoryKey            = "factory"
	factoriesKey          = "factory"
	nonReadyUsagesKey     = "nonreadyusage"
	burTimeoutsKey        = "burtimeouts"
	timeUntilReadyKey     = "timeuntilready"
	integrationKey        = "integrations"
)

// CustomTelemetryWrapper interface for pluggable storages
type CustomTelemetryWrapper interface {
	Increment(key string, value int64)
	Set(key string, value int64)
	PushItem(key string, item interface{})

	GetItem(key string) int64
	GetByPrefix(prefix string) []interface{}
	PopItem(key string) int64
	PopItems(key string) interface{}
}

// UserCustomTelemetryAdapter struct
type UserCustomTelemetryAdapter struct {
	wrapper CustomTelemetryWrapper
}

// NewUserCustomTelemetryAdapter builds new user adapter
func NewUserCustomTelemetryAdapter(wrapper CustomTelemetryWrapper) TelemetryStorage {
	return &UserCustomTelemetryAdapter{
		wrapper: wrapper,
	}
}

// RecordLatency stores latency for method
func (u *UserCustomTelemetryAdapter) RecordLatency(method string, bucket int) {
	u.wrapper.Increment(fmt.Sprintf("%s::latency::%s::%d", methodKey, method, bucket), 1)
}

// RecordException stores exceptions for method
func (u *UserCustomTelemetryAdapter) RecordException(method string) {
	u.wrapper.Increment(fmt.Sprintf("%s::exception::%s", methodKey, method), 1)
}

// RecordDroppedImpressions increments dropped impressions
func (u *UserCustomTelemetryAdapter) RecordDroppedImpressions(count int64) {
	u.wrapper.Increment(fmt.Sprintf("%s::%s", dataKey, impressionsDroppedKey), count)
}

// RecordDedupedImpressions increments deduped impressions
func (u *UserCustomTelemetryAdapter) RecordDedupedImpressions(count int64) {
	u.wrapper.Increment(fmt.Sprintf("%s::%s", dataKey, impressionsDedupedKey), count)
}

// RecordQueuedImpressions increments queued impressions
func (u *UserCustomTelemetryAdapter) RecordQueuedImpressions(count int64) {
	u.wrapper.Increment(fmt.Sprintf("%s::%s", dataKey, impressionsQueuedKey), count)
}

// RecordDroppedEvents increments dropped events
func (u *UserCustomTelemetryAdapter) RecordDroppedEvents(count int64) {
	u.wrapper.Increment(fmt.Sprintf("%s::%s", dataKey, eventsDroppedKey), count)
}

// RecordQueuedEvents increments queued events
func (u *UserCustomTelemetryAdapter) RecordQueuedEvents(count int64) {
	u.wrapper.Increment(fmt.Sprintf("%s::%s", dataKey, eventsQueuedKey), count)
}

// RecordSuccessfulSplitSync records succesful sync
func (u *UserCustomTelemetryAdapter) RecordSuccessfulSplitSync(timestamp int64) {
	u.wrapper.Set(fmt.Sprintf("%s::%s", synchronizationKey, splitSync), timestamp)
}

// RecordSuccessfulSegmentSync records succesful sync
func (u *UserCustomTelemetryAdapter) RecordSuccessfulSegmentSync(timestamp int64) {
	u.wrapper.Set(fmt.Sprintf("%s::%s", synchronizationKey, segmentSync), timestamp)
}

// RecordSuccessfulImpressionSync records succesful sync
func (u *UserCustomTelemetryAdapter) RecordSuccessfulImpressionSync(timestamp int64) {
	u.wrapper.Set(fmt.Sprintf("%s::%s", synchronizationKey, impressionSync), timestamp)
}

// RecordSuccessfulEventsSync records succesful sync
func (u *UserCustomTelemetryAdapter) RecordSuccessfulEventsSync(timestamp int64) {
	u.wrapper.Set(fmt.Sprintf("%s::%s", synchronizationKey, eventSync), timestamp)
}

// RecordSuccessfulTelemetrySync records succesful sync
func (u *UserCustomTelemetryAdapter) RecordSuccessfulTelemetrySync(timestamp int64) {
	u.wrapper.Set(fmt.Sprintf("%s::%s", synchronizationKey, telemetrySync), timestamp)
}

// RecordSuccessfulTokenGet records succesful sync
func (u *UserCustomTelemetryAdapter) RecordSuccessfulTokenGet(timestamp int64) {
	u.wrapper.Set(fmt.Sprintf("%s::%s", synchronizationKey, tokenSync), timestamp)
}

// RecordSyncError records http error
func (u *UserCustomTelemetryAdapter) RecordSyncError(path string, status int) {
	u.wrapper.Increment(fmt.Sprintf("%s::%s::%d", httpKey, path, status), 1)
}

// RecordSyncLatency records http error
func (u *UserCustomTelemetryAdapter) RecordSyncLatency(path string, bucket int) {
	u.wrapper.Increment(fmt.Sprintf("%s::latency::%s::%d", httpKey, path, bucket), 1)
}

// RecordAuthRejections records auth rejections
func (u *UserCustomTelemetryAdapter) RecordAuthRejections() {
	u.wrapper.Increment(fmt.Sprintf("%s::%s", pushKey, authRejectionsKey), 1)
}

// RecordTokenRefreshes records token
func (u *UserCustomTelemetryAdapter) RecordTokenRefreshes() {
	u.wrapper.Increment(fmt.Sprintf("%s::%s", pushKey, tokenRefreshesKey), 1)
}

// RecordStreamingEvent appends new streaming event
func (u *UserCustomTelemetryAdapter) RecordStreamingEvent(event dto.StreamingEvent) {
	u.wrapper.PushItem(fmt.Sprintf("%s", streamingKey), event)
}

// AddTag adds particular tag
func (u *UserCustomTelemetryAdapter) AddTag(tag string) {
	u.wrapper.PushItem(fmt.Sprintf("%s", tagKey), tag)
}

// RecordSessionLength records session length
func (u *UserCustomTelemetryAdapter) RecordSessionLength(session int64) {
	u.wrapper.Set(fmt.Sprintf("%s::%s", sdkKey, sessionKey), session)
}

// RecordFactory tracks factory
func (u *UserCustomTelemetryAdapter) RecordFactory(apikey string) {
	u.wrapper.Increment(fmt.Sprintf("%s::%s::%s", factoryKey, factoriesKey, apikey), 1)
}

// RecordNonReadyUsage records non ready usage
func (u *UserCustomTelemetryAdapter) RecordNonReadyUsage() {
	u.wrapper.Increment(fmt.Sprintf("%s::%s", factoryKey, nonReadyUsagesKey), 1)
}

// RecordBURTimeout records bur timeodout
func (u *UserCustomTelemetryAdapter) RecordBURTimeout() {
	u.wrapper.Increment(fmt.Sprintf("%s::%s", factoryKey, burTimeoutsKey), 1)
}

// RecordTimeUntilReady records time took until ready
func (u *UserCustomTelemetryAdapter) RecordTimeUntilReady(time int64) {
	u.wrapper.Set(fmt.Sprintf("%s::%s", factoryKey, timeUntilReadyKey), 1)
}

// AddIntegration adds particular integration
func (u *UserCustomTelemetryAdapter) AddIntegration(integration string) {
	u.wrapper.PushItem(fmt.Sprintf("%s", integrationKey), integration)
}

// TELEMETRY STORAGE CONSUMER

func (u *UserCustomTelemetryAdapter) parseLatency(key string, method string) []int64 {
	latencies := make([]int64, latencyBucketCount)
	for i := 0; i < latencyBucketCount; i++ {
		key := fmt.Sprintf("%s::latency::%s::%d", key, method, i)
		latencies[i] = u.wrapper.PopItem(key)
	}
	return latencies
}

// PopLatencies gets and clears method latencies
func (u *UserCustomTelemetryAdapter) PopLatencies() dto.MethodLatencies {
	return dto.MethodLatencies{
		Treatment:            u.parseLatency(methodKey, treatment),
		Treatments:           u.parseLatency(methodKey, treatments),
		TreatmentWithConfig:  u.parseLatency(methodKey, treatmentWithConfig),
		TreatmentWithConfigs: u.parseLatency(methodKey, treatmentsWithConfig),
		Track:                u.parseLatency(methodKey, track),
	}
}

// PopExceptions gets and clears method exceptions
func (u *UserCustomTelemetryAdapter) PopExceptions() dto.MethodExceptions {
	return dto.MethodExceptions{
		Treatment:            u.wrapper.PopItem(fmt.Sprintf("%s::exception::%s", methodKey, treatment)),
		Treatments:           u.wrapper.PopItem(fmt.Sprintf("%s::exception::%s", methodKey, treatments)),
		TreatmentWithConfig:  u.wrapper.PopItem(fmt.Sprintf("%s::exception::%s", methodKey, treatmentWithConfig)),
		TreatmentWithConfigs: u.wrapper.PopItem(fmt.Sprintf("%s::exception::%s", methodKey, treatmentsWithConfig)),
		Track:                u.wrapper.PopItem(fmt.Sprintf("%s::exception::%s", methodKey, track)),
	}
}

// GetDroppedImpressions gets total amount of impressions dropped
func (u *UserCustomTelemetryAdapter) GetDroppedImpressions() int64 {
	return u.wrapper.GetItem(fmt.Sprintf("%s::%s", dataKey, impressionsDroppedKey))
}

// GetDedupedImpressions gets total amount of impressions deduped
func (u *UserCustomTelemetryAdapter) GetDedupedImpressions() int64 {
	return u.wrapper.GetItem(fmt.Sprintf("%s::%s", dataKey, impressionsDedupedKey))
}

// GetQueuedmpressions gets total amount of impressions queued
func (u *UserCustomTelemetryAdapter) GetQueuedmpressions() int64 {
	return u.wrapper.GetItem(fmt.Sprintf("%s::%s", dataKey, impressionsQueuedKey))
}

// GetDroppedEvents gets total amount of events dropped
func (u *UserCustomTelemetryAdapter) GetDroppedEvents() int64 {
	return u.wrapper.GetItem(fmt.Sprintf("%s::%s", dataKey, eventsDroppedKey))
}

// GetQueuedEvents gets total amount of events queued
func (u *UserCustomTelemetryAdapter) GetQueuedEvents() int64 {
	return u.wrapper.GetItem(fmt.Sprintf("%s::%s", dataKey, eventsQueuedKey))
}

// GetLastSynchronization gets last synchronization stats for fetchers and recorders
func (u *UserCustomTelemetryAdapter) GetLastSynchronization() dto.LastSynchronization {
	return dto.LastSynchronization{
		Splits:      u.wrapper.GetItem(fmt.Sprintf("%s::%s", synchronizationKey, splitSync)),
		Segments:    u.wrapper.GetItem(fmt.Sprintf("%s::%s", synchronizationKey, segmentSync)),
		Impressions: u.wrapper.GetItem(fmt.Sprintf("%s::%s", synchronizationKey, impressionSync)),
		Events:      u.wrapper.GetItem(fmt.Sprintf("%s::%s", synchronizationKey, eventSync)),
		Telemetry:   u.wrapper.GetItem(fmt.Sprintf("%s::%s", synchronizationKey, telemetrySync)),
		Token:       u.wrapper.GetItem(fmt.Sprintf("%s::%s", synchronizationKey, tokenSync)),
	}
}

func (u *UserCustomTelemetryAdapter) parseHTTPError(path string) map[int]int64 {
	toReturn := make(map[int]int64, 0)
	items := u.wrapper.GetByPrefix(fmt.Sprintf("%s::%s::", httpKey, path))
	for _, item := range items {
		asString, ok := item.(string)
		if !ok {
			continue
		}
		splitted := strings.Split(asString, "::")
		if len(splitted) != 3 {
			continue
		}
		statusCodeString := splitted[2]
		statusCode, err := strconv.Atoi(statusCodeString)
		if err != nil {
			continue
		}
		toReturn[statusCode] = u.wrapper.PopItem(fmt.Sprintf("%s::%s::%d", httpKey, path, statusCode))
	}
	return toReturn
}

// PopHTTPErrors gets http errors
func (u *UserCustomTelemetryAdapter) PopHTTPErrors() dto.HTTPErrors {
	return dto.HTTPErrors{
		Splits:      u.parseHTTPError(splitSync),
		Segments:    u.parseHTTPError(segmentSync),
		Impressions: u.parseHTTPError(impressionSync),
		Events:      u.parseHTTPError(eventSync),
		Telemetry:   u.parseHTTPError(telemetrySync),
		Token:       u.parseHTTPError(tokenSync),
	}
}

// PopHTTPLatencies gets http latencies
func (u *UserCustomTelemetryAdapter) PopHTTPLatencies() dto.HTTPLatencies {
	return dto.HTTPLatencies{
		Splits:      u.parseLatency(httpKey, splitSync),
		Segments:    u.parseLatency(httpKey, segmentSync),
		Impressions: u.parseLatency(httpKey, impressionSync),
		Events:      u.parseLatency(httpKey, eventSync),
		Telemetry:   u.parseLatency(httpKey, telemetrySync),
		Token:       u.parseLatency(httpKey, tokenSync),
	}
}

// PopAuthRejections gets total amount of auth rejections
func (u *UserCustomTelemetryAdapter) PopAuthRejections() int64 {
	return u.wrapper.PopItem(fmt.Sprintf("%s::%s", pushKey, authRejectionsKey))
}

// PopTokenRefreshes gets total amount of token refreshes
func (u *UserCustomTelemetryAdapter) PopTokenRefreshes() int64 {
	return u.wrapper.PopItem(fmt.Sprintf("%s::%s", pushKey, tokenRefreshesKey))
}

// PopStreamingEvents gets streamingEvents data
func (u *UserCustomTelemetryAdapter) PopStreamingEvents() []dto.StreamingEvent {
	toReturn := make([]dto.StreamingEvent, 0, maxStreamingEvents)
	items := u.wrapper.PopItems(fmt.Sprintf("%s", streamingKey))
	streamingEventsAsInterface, ok := items.([]interface{})
	if !ok {
		return toReturn
	}
	for _, asInterface := range streamingEventsAsInterface {
		streamingEvent, ok := asInterface.(dto.StreamingEvent)
		if !ok {
			continue
		}
		toReturn = append(toReturn, streamingEvent)
	}
	return toReturn
}

// PopTags gets total amount of tags
func (u *UserCustomTelemetryAdapter) PopTags() []string {
	toReturn := make([]string, 0, maxTags)
	items := u.wrapper.PopItems(fmt.Sprintf("%s", tagKey))
	tagsAsInterface, ok := items.([]interface{})
	if !ok {
		return toReturn
	}
	for _, asInterface := range tagsAsInterface {
		tag, ok := asInterface.(string)
		if !ok {
			continue
		}
		toReturn = append(toReturn, tag)
	}
	return toReturn
}

// GetSessionLength gets session duration
func (u *UserCustomTelemetryAdapter) GetSessionLength() int64 {
	return u.wrapper.GetItem(fmt.Sprintf("%s::%s", sdkKey, sessionKey))
}

// GetActiveFactories gets active factories instantiated
func (u *UserCustomTelemetryAdapter) GetActiveFactories() int64 {
	var toReturn int64 = 0
	items := u.wrapper.GetByPrefix(fmt.Sprintf("%s::%s::", factoryKey, factoriesKey))
	for _, item := range items {
		asString, ok := item.(string)
		if !ok {
			continue
		}
		splitted := strings.Split(asString, "::")
		if len(splitted) != 3 {
			continue
		}
		apikey := splitted[2]
		toReturn += u.wrapper.GetItem(fmt.Sprintf("%s::%s::%s", factoryKey, factoriesKey, apikey))
	}
	return toReturn
}

// GetRedundantActiveFactories gets redundant factories
func (u *UserCustomTelemetryAdapter) GetRedundantActiveFactories() int64 {
	var toReturn int64 = 0
	items := u.wrapper.GetByPrefix(fmt.Sprintf("%s::%s::", factoryKey, factoriesKey))
	for _, item := range items {
		asString, ok := item.(string)
		if !ok {
			continue
		}
		splitted := strings.Split(asString, "::")
		if len(splitted) != 3 {
			continue
		}
		apikey := splitted[2]
		factories := u.wrapper.GetItem(fmt.Sprintf("%s::%s::%s", factoryKey, factoriesKey, apikey))
		if factories > 1 {
			toReturn += factories - 1
		}
	}
	return toReturn
}

// GetNonReadyUsages gets non usages on ready
func (u *UserCustomTelemetryAdapter) GetNonReadyUsages() int64 {
	return u.wrapper.GetItem(fmt.Sprintf("%s::%s", factoryKey, nonReadyUsagesKey))
}

// GetBURTimeouts gets timedouts data
func (u *UserCustomTelemetryAdapter) GetBURTimeouts() int64 {
	return u.wrapper.GetItem(fmt.Sprintf("%s::%s", factoryKey, burTimeoutsKey))
}

// GetTimeUntilReady gets duration until ready
func (u *UserCustomTelemetryAdapter) GetTimeUntilReady() int64 {
	return u.wrapper.GetItem(fmt.Sprintf("%s::%s", factoryKey, timeUntilReadyKey))
}

// GetIntegrations gets total amount of integrations
func (u *UserCustomTelemetryAdapter) GetIntegrations() []string {
	toReturn := make([]string, 0, maxTags)
	items := u.wrapper.PopItems(fmt.Sprintf("%s", integrationKey))
	integrationAsInterface, ok := items.([]interface{})
	if !ok {
		return toReturn
	}
	for _, asInterface := range integrationAsInterface {
		integration, ok := asInterface.(string)
		if !ok {
			continue
		}
		toReturn = append(toReturn, integration)
	}
	return toReturn
}
