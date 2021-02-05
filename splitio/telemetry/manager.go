package telemetry

import (
	"os"
	"strings"

	"github.com/splitio/go-client/splitio/conf"
	defaultConfig "github.com/splitio/go-split-commons/conf"
)

const (
	operationModeStandalone = iota
	operationModeConsumer
)

const (
	impressionsModeOptimized = iota
	impressionsModeDebug

	redis  = "redis"
	memory = "memory"
)

// ManagerImpl struct for building metrics
type ManagerImpl struct {
	factory         FactoryTelemetryConsumer
	evaluation      EvaluationTelemetryConsumer
	impression      ImpressionTelemetryConsumer
	event           EventTelemetryConsumer
	synchronization SynchronizationTelemetryConsumer
	http            HTTPTelemetryConsumer
	cache           CacheTelemetryConsumer
	push            PushTelemetryConsumer
	streaming       StreamingTelemetryConsumer
	sdk             SDKInfoTelemetryConsumer
	misc            MiscTelemetryConsumer
}

// NewTelemetryManager creates new manager
func NewTelemetryManager(
	factory FactoryTelemetryConsumer,
	evaluation EvaluationTelemetryConsumer,
	impression ImpressionTelemetryConsumer,
	event EventTelemetryConsumer,
	synchronization SynchronizationTelemetryConsumer,
	http HTTPTelemetryConsumer,
	cache CacheTelemetryConsumer,
	push PushTelemetryConsumer,
	streaming StreamingTelemetryConsumer,
	sdk SDKInfoTelemetryConsumer,
	misc MiscTelemetryConsumer,
) TelemetryManager {
	return &ManagerImpl{
		factory:         factory,
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

func getURLOverrides(cfg conf.AdvancedConfig) URLOverrides {
	sdk := false
	events := false
	auth := false
	streaming := false
	defaults := defaultConfig.GetDefaultAdvancedConfig()
	if cfg.SdkURL != defaults.SdkURL {
		sdk = true
	}
	if cfg.EventsURL != defaults.EventsURL {
		events = true
	}
	if cfg.AuthServiceURL != defaults.AuthServiceURL {
		auth = true
	}
	if cfg.StreamingServiceURL != defaults.StreamingServiceURL {
		streaming = true
	}
	return URLOverrides{
		Sdk:    sdk,
		Events: events,
		Auth:   auth,
		Stream: streaming,
	}
}

// BuildConfigData returns config data
func (m *ManagerImpl) BuildConfigData(cfg *conf.SplitSdkConfig) ConfigMetrics {
	operationMode := operationModeStandalone
	storage := memory
	if cfg.OperationMode == conf.RedisConsumer {
		operationMode = operationModeConsumer
		storage = redis
	}
	impressionsMode := impressionsModeOptimized
	if cfg.ImpressionsMode == defaultConfig.ImpressionsModeDebug {
		impressionsMode = impressionsModeDebug
	}
	proxyEnabled := false
	if len(strings.TrimSpace(os.Getenv("HTTP_PROXY"))) > 0 {
		proxyEnabled = true
	}
	return ConfigMetrics{
		OperationMode:    operationMode,
		Storage:          storage,
		StreamingEnabled: cfg.Advanced.StreamingEnabled,
		Rates: Rates{
			Splits:      int64(cfg.TaskPeriods.SplitSync),
			Segments:    int64(cfg.TaskPeriods.SegmentSync),
			Impressions: int64(cfg.TaskPeriods.ImpressionSync),
			Events:      int64(cfg.TaskPeriods.EventsSync),
			Telemetry:   int64(cfg.TaskPeriods.CounterSync), // It should be TelemetrySync after refactor in go
		},
		URLOverrides:               getURLOverrides(cfg.Advanced),
		ImpressionsQueueSize:       int64(cfg.Advanced.ImpressionsQueueSize),
		EventsQueueSize:            int64(cfg.Advanced.EventsQueueSize),
		ImpressionsMode:            impressionsMode,
		ImpressionsListenerEnabled: cfg.Advanced.ImpressionListener != nil,
		HTTPProxyDetected:          proxyEnabled,
		ActiveFactories:            m.factory.GetActiveFactories(),
		RedundantFactories:         m.factory.GetRedundantActiveFactories(),
		TimeUntilReady:             m.factory.GetTimeUntilReady(),
		BurTimeouts:                m.factory.GetBURTimeouts(),
		NonReadyUsages:             m.factory.GetNonReadyUsages(),
		Tags:                       m.misc.PopTags(),
	}
}

// BuildUsageData returns usage data
func (m *ManagerImpl) BuildUsageData() RegularMetrics {
	return RegularMetrics{
		MethodLatencies:      m.evaluation.PopLatencies(),
		MethodExceptions:     m.evaluation.PopExceptions(),
		ImpressionsDropped:   m.impression.GetDroppedImpressions(),
		ImpressionsDeduped:   m.impression.GetDedupedImpressions(),
		ImpressionsQueued:    m.impression.GetQueuedmpressions(),
		EventsQueued:         m.event.GetQueuedEvents(),
		EventsDropped:        m.event.GetDroppedEvents(),
		LastSynchronizations: m.synchronization.GetLastSynchronization(),
		HTTPErrors:           m.http.PopHTTPErrors(),
		HTTPLatencies:        m.http.PopHTTPLatencies(),
		SplitCount:           m.cache.GetSplitsCount(),
		SegmentCount:         m.cache.GetSegmentsCount(),
		SegmentKeyCount:      m.cache.GetSegmentKeysCount(),
		TokenRefreshes:       m.push.PopTokenRefreshes(),
		AuthRejections:       m.push.PopAuthRejections(),
		StreamingEvents:      m.streaming.PopStreamingEvents(),
		SessionLengthMs:      m.sdk.PopSessionLength(),
		Tags:                 m.misc.PopTags(),
	}
}
