package telemetry

import (
	"os"
	"strings"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/constants"
	"github.com/splitio/go-client/splitio/dto"
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
	telemetryConsumer FacadeConsumer
}

// NewTelemetryManager creates new manager
func NewTelemetryManager(telemetryConsumer FacadeConsumer) Manager {
	return &ManagerImpl{
		telemetryConsumer: telemetryConsumer,
	}
}

func getURLOverrides(cfg conf.AdvancedConfig) dto.URLOverrides {
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
	return dto.URLOverrides{
		Sdk:    sdk,
		Events: events,
		Auth:   auth,
		Stream: streaming,
	}
}

// BuildInitData returns config data
func (m *ManagerImpl) BuildInitData(cfg *conf.SplitSdkConfig, timedUntilReady int64, factoryInstances map[string]int64) dto.InitData {
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
	return dto.InitData{
		OperationMode:    operationMode,
		Storage:          storage,
		StreamingEnabled: cfg.Advanced.StreamingEnabled,
		Rates: dto.Rates{
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
		// ActiveFactories:            m.telemetryConsumer.GetActiveFactories(),
		// RedundantFactories:         m.telemetryConsumer.GetRedundantActiveFactories(),
		TimeUntilReady: timedUntilReady,
		BurTimeouts:    m.telemetryConsumer.GetBURTimeouts(),
		NonReadyUsages: m.telemetryConsumer.GetNonReadyUsages(),
		Tags:           m.telemetryConsumer.PopTags(),
	}
}

// BuildStatsData returns usage data
func (m *ManagerImpl) BuildStatsData() dto.StatsData {
	return dto.StatsData{
		MethodLatencies:      m.telemetryConsumer.PopLatencies(),
		MethodExceptions:     m.telemetryConsumer.PopExceptions(),
		ImpressionsDropped:   m.telemetryConsumer.GetImpressionsStats(constants.ImpressionsDropped),
		ImpressionsDeduped:   m.telemetryConsumer.GetImpressionsStats(constants.ImpressionsDeduped),
		ImpressionsQueued:    m.telemetryConsumer.GetImpressionsStats(constants.ImpressionsQueued),
		EventsQueued:         m.telemetryConsumer.GetEventsStats(constants.EventsQueued),
		EventsDropped:        m.telemetryConsumer.GetEventsStats(constants.EventsDropped),
		LastSynchronizations: m.telemetryConsumer.GetLastSynchronization(),
		HTTPErrors:           m.telemetryConsumer.PopHTTPErrors(),
		HTTPLatencies:        m.telemetryConsumer.PopHTTPLatencies(),
		SplitCount:           m.telemetryConsumer.GetSplitsCount(),
		SegmentCount:         m.telemetryConsumer.GetSegmentsCount(),
		SegmentKeyCount:      m.telemetryConsumer.GetSegmentKeysCount(),
		TokenRefreshes:       m.telemetryConsumer.PopTokenRefreshes(),
		AuthRejections:       m.telemetryConsumer.PopAuthRejections(),
		StreamingEvents:      m.telemetryConsumer.PopStreamingEvents(),
		SessionLengthMs:      m.telemetryConsumer.GetSessionLength(),
		Tags:                 m.telemetryConsumer.PopTags(),
	}
}
