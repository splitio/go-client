package conf

import (
	"strings"

	"github.com/splitio/go-split-commons/v6/conf"
	"github.com/splitio/go-split-commons/v6/flagsets"
)

// NormalizeSDKConf compares against SDK Config to set defaults
func NormalizeSDKConf(sdkConfig AdvancedConfig) (conf.AdvancedConfig, []error) {
	config := conf.GetDefaultAdvancedConfig()
	if sdkConfig.HTTPTimeout > 0 {
		config.HTTPTimeout = sdkConfig.HTTPTimeout
	}
	if sdkConfig.EventsBulkSize > 0 {
		config.EventsBulkSize = sdkConfig.EventsBulkSize
	}
	if sdkConfig.EventsQueueSize > 0 {
		config.EventsQueueSize = sdkConfig.EventsQueueSize
	}
	if sdkConfig.ImpressionsBulkSize > 0 {
		config.ImpressionsBulkSize = sdkConfig.ImpressionsBulkSize
	}
	if sdkConfig.ImpressionsQueueSize > 0 {
		config.ImpressionsQueueSize = sdkConfig.ImpressionsQueueSize
	}
	if sdkConfig.SegmentQueueSize > 0 {
		config.SegmentQueueSize = sdkConfig.SegmentQueueSize
	}
	if sdkConfig.SegmentWorkers > 0 {
		config.SegmentWorkers = sdkConfig.SegmentWorkers
	}
	if strings.TrimSpace(sdkConfig.EventsURL) != "" {
		config.EventsURL = sdkConfig.EventsURL
	}
	if strings.TrimSpace(sdkConfig.SdkURL) != "" {
		config.SdkURL = sdkConfig.SdkURL
	}
	if strings.TrimSpace(sdkConfig.AuthServiceURL) != "" {
		config.AuthServiceURL = sdkConfig.AuthServiceURL
	}
	if strings.TrimSpace(sdkConfig.StreamingServiceURL) != "" {
		config.StreamingServiceURL = sdkConfig.StreamingServiceURL
	}
	if strings.TrimSpace(sdkConfig.TelemetryServiceURL) != "" {
		config.TelemetryServiceURL = sdkConfig.TelemetryServiceURL
	}
	config.StreamingEnabled = sdkConfig.StreamingEnabled

	flagSets, errs := flagsets.SanitizeMany(sdkConfig.FlagSetsFilter)
	config.FlagSetsFilter = flagSets
	return config, errs
}
