package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-toolkit/logging"
)

type httpRecorderBase struct {
	client *HTTPClient
	logger logging.LoggerInterface
}

func (h *httpRecorderBase) recordRaw(
	url string,
	data []byte,
	sdkVersion string,
	machineIP string,
	machineName string,
) error {
	headers := make(map[string]string)
	headers["SplitSDKVersion"] = sdkVersion
	headers["SplitSDKMachineIP"] = machineIP
	if machineName == "" && machineIP != "" {
		headers["SplitSDKMachineName"] = fmt.Sprintf("ip-%s", strings.Replace(machineIP, ".", "-", -1))
	} else {
		headers["SplitSDKMachineName"] = machineName
	}
	return h.client.Post(url, data, headers)
}

// HTTPImpressionRecorder is a struct responsible for submitting impression bulks to the backend
type HTTPImpressionRecorder struct {
	httpRecorderBase
}

// Record sends an array (or slice) of dtos.ImpressionsDTO to the backend
func (i *HTTPImpressionRecorder) Record(
	impressions []dtos.ImpressionsDTO,
	sdkVersion string,
	machineIP string,
	machineName string,
) error {
	data, err := json.Marshal(impressions)
	if err != nil {
		i.logger.Error("Error marshaling JSON", err.Error())
		return err
	}

	err = i.recordRaw("/testImpressions/bulk", data, sdkVersion, machineIP, machineName)
	if err != nil {
		i.logger.Error("Error posting impressions", err.Error())
		return err
	}

	return nil
}

// NewHTTPImpressionRecorder instantiates an HTTPImpressionRecorder
func NewHTTPImpressionRecorder(
	apikey string,
	cfg *configuration.SplitSdkConfig,
	logger logging.LoggerInterface,
) *HTTPImpressionRecorder {
	_, eventsURL := getUrls(cfg.Advanced)
	client := NewHTTPClient(apikey, cfg, eventsURL, splitio.Version, logger)
	return &HTTPImpressionRecorder{
		httpRecorderBase: httpRecorderBase{
			client: client,
			logger: logger,
		},
	}
}

// HTTPMetricsRecorder is a struct responsible for submitting metrics (latency, gauge, counters) to the backend
type HTTPMetricsRecorder struct {
	httpRecorderBase
}

// RecordCounters method submits counter metrics to the backend
func (m *HTTPMetricsRecorder) RecordCounters(
	counters []dtos.CounterDTO,
	sdkVersion string,
	machineIP string,
	machineName string,
) error {
	data, err := json.Marshal(counters)
	if err != nil {
		m.logger.Error("Error marshaling JSON", err.Error())
		return err
	}

	err = m.recordRaw("/metrics/counters", data, sdkVersion, machineIP, machineName)
	if err != nil {
		m.logger.Error("Error posting impressions", err.Error())
		return err
	}

	return nil
}

// RecordLatencies method submits latency metrics to the backend
func (m *HTTPMetricsRecorder) RecordLatencies(
	latencies []dtos.LatenciesDTO,
	sdkVersion string,
	machineIP string,
	machineName string,
) error {
	data, err := json.Marshal(latencies)
	if err != nil {
		m.logger.Error("Error marshaling JSON", err.Error())
		return err
	}

	err = m.recordRaw("/metrics/times", data, sdkVersion, machineIP, machineName)
	if err != nil {
		m.logger.Error("Error posting impressions", err.Error())
		return err
	}

	return nil
}

// RecordGauge method submits gauge metrics to the backend
func (m *HTTPMetricsRecorder) RecordGauge(
	gauge dtos.GaugeDTO,
	sdkVersion string,
	machineIP string,
	machineName string,
) error {
	data, err := json.Marshal(gauge)
	if err != nil {
		m.logger.Error("Error marshaling JSON", err.Error())
		return err
	}

	err = m.recordRaw("/metrics/gauge", data, sdkVersion, machineIP, machineName)
	if err != nil {
		m.logger.Error("Error posting impressions", err.Error())
		return err
	}

	return nil
}

// NewHTTPMetricsRecorder instantiates an HTTPMetricsRecorder
func NewHTTPMetricsRecorder(
	apikey string,
	cfg *configuration.SplitSdkConfig,
	logger logging.LoggerInterface,
) *HTTPMetricsRecorder {
	_, eventsURL := getUrls(cfg.Advanced)
	client := NewHTTPClient(apikey, cfg, eventsURL, splitio.Version, logger)
	return &HTTPMetricsRecorder{
		httpRecorderBase: httpRecorderBase{
			client: client,
			logger: logger,
		},
	}
}
