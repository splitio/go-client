package storage

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/splitio/go-client/splitio/constants"
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
	nonReadyUsagesKey     = "nonreadyusage"
	burTimeoutsKey        = "burtimeouts"
)

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
func (u *UserCustomTelemetryAdapter) RecordLatency(resource int, bucket int) {
	u.wrapper.Increment(fmt.Sprintf("%s::latency::%d::%d", methodKey, resource, bucket), 1)
}

// RecordException stores exceptions for method
func (u *UserCustomTelemetryAdapter) RecordException(method int) {
	u.wrapper.Increment(fmt.Sprintf("%s::exception::%d", methodKey, method), 1)
}

// RecordImpressionsStats records impressions by type
func (u *UserCustomTelemetryAdapter) RecordImpressionsStats(dataType int, count int64) {
	switch dataType {
	case constants.ImpressionsDropped:
		u.wrapper.Increment(fmt.Sprintf("%s::%s", dataKey, impressionsDroppedKey), count)
	case constants.ImpressionsDeduped:
		u.wrapper.Increment(fmt.Sprintf("%s::%s", dataKey, impressionsDedupedKey), count)
	case constants.ImpressionsQueued:
		u.wrapper.Increment(fmt.Sprintf("%s::%s", dataKey, impressionsQueuedKey), count)
	}
}

// RecordEventsStats recirds events by type
func (u *UserCustomTelemetryAdapter) RecordEventsStats(dataType int, count int64) {
	switch dataType {
	case constants.EventsDropped:
		u.wrapper.Increment(fmt.Sprintf("%s::%s", dataKey, eventsDroppedKey), count)
	case constants.EventsQueued:
		u.wrapper.Increment(fmt.Sprintf("%s::%s", dataKey, eventsQueuedKey), count)
	}
}

// RecordSuccessfulSync records sync for resource
func (u *UserCustomTelemetryAdapter) RecordSuccessfulSync(resource int, timestamp int64) {
	switch resource {
	case constants.SplitSync:
		u.wrapper.Set(fmt.Sprintf("%s::%d", synchronizationKey, constants.SplitSync), timestamp)
	case constants.SegmentSync:
		u.wrapper.Set(fmt.Sprintf("%s::%d", synchronizationKey, constants.SegmentSync), timestamp)
	case constants.ImpressionSync:
		u.wrapper.Set(fmt.Sprintf("%s::%d", synchronizationKey, constants.ImpressionSync), timestamp)
	case constants.EventSync:
		u.wrapper.Set(fmt.Sprintf("%s::%d", synchronizationKey, constants.EventSync), timestamp)
	case constants.TelemetrySync:
		u.wrapper.Set(fmt.Sprintf("%s::%d", synchronizationKey, constants.TelemetrySync), timestamp)
	case constants.TokenSync:
		u.wrapper.Set(fmt.Sprintf("%s::%d", synchronizationKey, constants.TokenSync), timestamp)
	}
}

// RecordSyncError records http error
func (u *UserCustomTelemetryAdapter) RecordSyncError(resource int, status int) {
	u.wrapper.Increment(fmt.Sprintf("%s::%d::%d", httpKey, resource, status), 1)
}

// RecordSyncLatency records http error
func (u *UserCustomTelemetryAdapter) RecordSyncLatency(resource int, bucket int) {
	u.wrapper.Increment(fmt.Sprintf("%s::latency::%d::%d", httpKey, resource, bucket), 1)
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

// RecordNonReadyUsage records non ready usage
func (u *UserCustomTelemetryAdapter) RecordNonReadyUsage() {
	u.wrapper.Increment(fmt.Sprintf("%s::%s", factoryKey, nonReadyUsagesKey), 1)
}

// RecordBURTimeout records bur timeodout
func (u *UserCustomTelemetryAdapter) RecordBURTimeout() {
	u.wrapper.Increment(fmt.Sprintf("%s::%s", factoryKey, burTimeoutsKey), 1)
}

// TELEMETRY STORAGE CONSUMER

func (u *UserCustomTelemetryAdapter) parseLatency(key string, method int) []int64 {
	latencies := make([]int64, constants.LatencyBucketCount)
	for i := 0; i < constants.LatencyBucketCount; i++ {
		key := fmt.Sprintf("%s::latency::%d::%d", key, method, i)
		asInterface := u.wrapper.PopItem(key)
		latencies[i] = convertToInt64(asInterface)
	}
	return latencies
}

// PopLatencies gets and clears method latencies
func (u *UserCustomTelemetryAdapter) PopLatencies() dto.MethodLatencies {
	return dto.MethodLatencies{
		Treatment:            u.parseLatency(methodKey, constants.Treatment),
		Treatments:           u.parseLatency(methodKey, constants.Treatments),
		TreatmentWithConfig:  u.parseLatency(methodKey, constants.TreatmentWithConfig),
		TreatmentWithConfigs: u.parseLatency(methodKey, constants.TreatmentsWithConfig),
		Track:                u.parseLatency(methodKey, constants.Track),
	}
}

// PopExceptions gets and clears method exceptions
func (u *UserCustomTelemetryAdapter) PopExceptions() dto.MethodExceptions {
	return dto.MethodExceptions{
		Treatment:            convertToInt64(u.wrapper.PopItem(fmt.Sprintf("%s::exception::%d", methodKey, constants.Treatment))),
		Treatments:           convertToInt64(u.wrapper.PopItem(fmt.Sprintf("%s::exception::%d", methodKey, constants.Treatments))),
		TreatmentWithConfig:  convertToInt64(u.wrapper.PopItem(fmt.Sprintf("%s::exception::%d", methodKey, constants.TreatmentWithConfig))),
		TreatmentWithConfigs: convertToInt64(u.wrapper.PopItem(fmt.Sprintf("%s::exception::%d", methodKey, constants.TreatmentsWithConfig))),
		Track:                convertToInt64(u.wrapper.PopItem(fmt.Sprintf("%s::exception::%d", methodKey, constants.Track))),
	}
}

// GetImpressionsStats gets impressions by type
func (u *UserCustomTelemetryAdapter) GetImpressionsStats(dataType int) int64 {
	switch dataType {
	case constants.ImpressionsDropped:
		return convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%s", dataKey, impressionsDroppedKey)))
	case constants.ImpressionsDeduped:
		return convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%s", dataKey, impressionsDedupedKey)))
	case constants.ImpressionsQueued:
		return convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%s", dataKey, impressionsQueuedKey)))
	}
	return 0
}

// GetEventsStats gets events by type
func (u *UserCustomTelemetryAdapter) GetEventsStats(dataType int) int64 {
	switch dataType {
	case constants.EventsDropped:
		return convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%s", dataKey, eventsDroppedKey)))
	case constants.EventsQueued:
		return convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%s", dataKey, eventsQueuedKey)))
	}
	return 0
}

// GetLastSynchronization gets last synchronization stats for fetchers and recorders
func (u *UserCustomTelemetryAdapter) GetLastSynchronization() dto.LastSynchronization {
	return dto.LastSynchronization{
		Splits:      convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%d", synchronizationKey, constants.SplitSync))),
		Segments:    convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%d", synchronizationKey, constants.SegmentSync))),
		Impressions: convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%d", synchronizationKey, constants.ImpressionSync))),
		Events:      convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%d", synchronizationKey, constants.EventSync))),
		Telemetry:   convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%d", synchronizationKey, constants.TelemetrySync))),
		Token:       convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%d", synchronizationKey, constants.TokenSync))),
	}
}

func (u *UserCustomTelemetryAdapter) parseHTTPError(resource int) map[int]int64 {
	toReturn := make(map[int]int64, 0)
	items := u.wrapper.GetByPrefix(fmt.Sprintf("%s::%d::", httpKey, resource))
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
		toReturn[statusCode] = convertToInt64(u.wrapper.PopItem(fmt.Sprintf("%s::%d::%d", httpKey, resource, statusCode)))
	}
	return toReturn
}

// PopHTTPErrors gets http errors
func (u *UserCustomTelemetryAdapter) PopHTTPErrors() dto.HTTPErrors {
	return dto.HTTPErrors{
		Splits:      u.parseHTTPError(constants.SplitSync),
		Segments:    u.parseHTTPError(constants.SegmentSync),
		Impressions: u.parseHTTPError(constants.ImpressionSync),
		Events:      u.parseHTTPError(constants.EventSync),
		Telemetry:   u.parseHTTPError(constants.TelemetrySync),
		Token:       u.parseHTTPError(constants.TokenSync),
	}
}

// PopHTTPLatencies gets http latencies
func (u *UserCustomTelemetryAdapter) PopHTTPLatencies() dto.HTTPLatencies {
	return dto.HTTPLatencies{
		Splits:      u.parseLatency(httpKey, constants.SplitSync),
		Segments:    u.parseLatency(httpKey, constants.SegmentSync),
		Impressions: u.parseLatency(httpKey, constants.ImpressionSync),
		Events:      u.parseLatency(httpKey, constants.EventSync),
		Telemetry:   u.parseLatency(httpKey, constants.TelemetrySync),
		Token:       u.parseLatency(httpKey, constants.TokenSync),
	}
}

// PopAuthRejections gets total amount of auth rejections
func (u *UserCustomTelemetryAdapter) PopAuthRejections() int64 {
	return convertToInt64(u.wrapper.PopItem(fmt.Sprintf("%s::%s", pushKey, authRejectionsKey)))
}

// PopTokenRefreshes gets total amount of token refreshes
func (u *UserCustomTelemetryAdapter) PopTokenRefreshes() int64 {
	return convertToInt64(u.wrapper.PopItem(fmt.Sprintf("%s::%s", pushKey, tokenRefreshesKey)))
}

// PopStreamingEvents gets streamingEvents data
func (u *UserCustomTelemetryAdapter) PopStreamingEvents() []dto.StreamingEvent {
	toReturn := make([]dto.StreamingEvent, 0, constants.MaxStreamingEvents)
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
	toReturn := make([]string, 0, constants.MaxTags)
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
	return convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%s", sdkKey, sessionKey)))
}

// GetNonReadyUsages gets non usages on ready
func (u *UserCustomTelemetryAdapter) GetNonReadyUsages() int64 {
	return convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%s", factoryKey, nonReadyUsagesKey)))
}

// GetBURTimeouts gets timedouts data
func (u *UserCustomTelemetryAdapter) GetBURTimeouts() int64 {
	return convertToInt64(u.wrapper.GetItem(fmt.Sprintf("%s::%s", factoryKey, burTimeoutsKey)))
}
