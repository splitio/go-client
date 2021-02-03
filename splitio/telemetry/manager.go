package telemetry

// ManagerImpl struct for building metrics
type ManagerImpl struct {
	evaluation      EvaluationTelemetryConsumer
	impression      ImpressionTelemetryConsumer
	event           EventTelemetryConsumer
	synchronization SynchronizationTelemetryConsumer
	http            HTTPErrorTelemetryConsumer
	cache           CacheTelemetryConsumer
	push            PushTelemetryConsumer
	streaming       StreamingTelemetryConsumer
	sdk             SDKInfoTelemetryConsumer
	misc            MiscTelemetryConsumer
}

// NewTelemetryManager creates new manager
func NewTelemetryManager(
	evaluation EvaluationTelemetryConsumer,
	impression ImpressionTelemetryConsumer,
	event EventTelemetryConsumer,
	synchronization SynchronizationTelemetryConsumer,
	http HTTPErrorTelemetryConsumer,
	cache CacheTelemetryConsumer,
	push PushTelemetryConsumer,
	streaming StreamingTelemetryConsumer,
	sdk SDKInfoTelemetryConsumer,
	misc MiscTelemetryConsumer,
) TelemetryManager {
	return &ManagerImpl{
		evaluation:      evaluation,
		impression:      impression,
		event:           event,
		synchronization: synchronization,
		http:            http,
		cache:           cache,
		push:            push,
		streaming:       streaming,
		sdk:             sdk,
		misc:            misc,
	}
}

// BuildRegularData implements regular data
func (m *ManagerImpl) BuildRegularData() RegularMetrics {
	return RegularMetrics{
		MethodLatencies:      m.evaluation.GetLatencies(),
		MethodExceptions:     m.evaluation.GetExceptions(),
		ImpressionsDropped:   m.impression.GetDroppedImpressions(),
		ImpressionsDeduped:   m.impression.GetDedupedImpressions(),
		ImpressionsQueued:    m.impression.GetQueuedmpressions(),
		EventsQueued:         m.event.GetQueuedEvents(),
		EventsDropped:        m.event.GetDroppedEvents(),
		LastSynchronizations: m.synchronization.GetLastSynchronization(),
		HTTPErrors:           m.http.GetHTTPErrors(),
		SplitCount:           m.cache.GetSplitsCount(),
		SegmentCount:         m.cache.GetSegmentCount(),
		SegmentKeyCount:      m.cache.GetSegmentKeyCount(),
		TokenRefreshes:       m.push.GetTokenRefreshes(),
		AuthRejections:       m.push.GetAuthRejections(),
		StreamingEvents:      m.streaming.GetStreamingEvents(),
		SessionLengthMs:      m.sdk.GetSessionLength(),
		Tags:                 m.misc.GetTags(),
	}
}
