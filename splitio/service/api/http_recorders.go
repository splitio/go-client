package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-client/splitio/util/logging"
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

	h.client.ResetHeaders()
	h.client.AddHeader("SplitSDKVersion", sdkVersion)
	h.client.AddHeader("SplitSDKMachineIP", machineIP)
	if machineName == "" && machineIP != "" {
		h.client.AddHeader("SplitSDKMachineName", fmt.Sprintf("ip-%s", strings.Replace(machineIP, ".", "-", -1)))
	} else {
		h.client.AddHeader("SplitSDKMachineName", machineName)
	}

	return h.client.Post(url, data)
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
func NewHTTPImpressionRecorder(cfg *configuration.SplitSdkConfig, logger logging.LoggerInterface) *HTTPImpressionRecorder {
	_, eventsURL := getUrls()
	client := NewHTTPClient(cfg, eventsURL, logger)
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

	err = m.recordRaw("/metrics/latencies", data, sdkVersion, machineIP, machineName)
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
func NewHTTPMetricsRecorder(cfg *configuration.SplitSdkConfig, logger logging.LoggerInterface) *HTTPMetricsRecorder {
	_, eventsURL := getUrls()
	client := NewHTTPClient(cfg, eventsURL, logger)
	return &HTTPMetricsRecorder{
		httpRecorderBase: httpRecorderBase{
			client: client,
			logger: logger,
		},
	}
}