package api

import (
	"encoding/json"

	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-split-commons/dtos"
	"github.com/splitio/go-toolkit/logging"
)

type httpRecorderBase struct {
	client   *HTTPClient
	logger   logging.LoggerInterface
	metadata dtos.Metadata
}

func (h *httpRecorderBase) recordRaw(url string, data []byte) error {
	headers := make(map[string]string)
	headers["SplitSDKVersion"] = h.metadata.SDKVersion
	if h.metadata.MachineName != "NA" && h.metadata.MachineName != "unknown" {
		headers["SplitSDKMachineName"] = h.metadata.MachineName
	}
	if h.metadata.MachineIP != "NA" && h.metadata.MachineIP != "unknown" {
		headers["SplitSDKMachineIP"] = h.metadata.MachineIP
	}
	return h.client.Post(url, data, headers)
}

// HTTPImpressionRecorder is a struct responsible for submitting impression bulks to the backend
type HTTPImpressionRecorder struct {
	httpRecorderBase
}

type impressionRecord struct {
	KeyName      string `json:"keyName"`
	Treatment    string `json:"treatment"`
	Time         int64  `json:"time"`
	ChangeNumber int64  `json:"changeNumber"`
	Label        string `json:"label"`
	BucketingKey string `json:"bucketingKey,omitempty"`
}

type impressionsRecord struct {
	TestName       string             `json:"testName"`
	KeyImpressions []impressionRecord `json:"keyImpressions"`
}

// Record sends an array (or slice) of impressionsRecord to the backend
func (i *HTTPImpressionRecorder) Record(impressions []dtos.Impression) error {
	impressionsToPost := make(map[string][]impressionRecord)
	for _, impression := range impressions {
		keyImpression := impressionRecord{
			KeyName:      impression.KeyName,
			Treatment:    impression.Treatment,
			Time:         impression.Time,
			ChangeNumber: impression.ChangeNumber,
			Label:        impression.Label,
			BucketingKey: impression.BucketingKey,
		}
		v, ok := impressionsToPost[impression.FeatureName]
		if ok {
			v = append(v, keyImpression)
		} else {
			v = []impressionRecord{keyImpression}
		}
		impressionsToPost[impression.FeatureName] = v
	}

	bulkImpressions := make([]impressionsRecord, 0)
	for testName, testImpressions := range impressionsToPost {
		bulkImpressions = append(bulkImpressions, impressionsRecord{
			TestName:       testName,
			KeyImpressions: testImpressions,
		})
	}

	data, err := json.Marshal(bulkImpressions)
	if err != nil {
		i.logger.Error("Error marshaling JSON", err.Error())
		return err
	}

	err = i.recordRaw("/testImpressions/bulk", data)
	if err != nil {
		i.logger.Error("Error posting impressions", err.Error())
		return err
	}

	return nil
}

// NewHTTPImpressionRecorder instantiates an HTTPImpressionRecorder
func NewHTTPImpressionRecorder(
	apikey string,
	cfg *conf.SplitSdkConfig,
	metadata dtos.Metadata,
	logger logging.LoggerInterface,
) *HTTPImpressionRecorder {
	_, eventsURL := getUrls(&cfg.Advanced)
	client := NewHTTPClient(apikey, cfg, eventsURL, splitio.Version, logger)
	return &HTTPImpressionRecorder{
		httpRecorderBase: httpRecorderBase{
			client:   client,
			logger:   logger,
			metadata: metadata,
		},
	}
}

// HTTPMetricsRecorder is a struct responsible for submitting metrics (latency, gauge, counters) to the backend
type HTTPMetricsRecorder struct {
	httpRecorderBase
}

// RecordCounters method submits counter metrics to the backend
func (m *HTTPMetricsRecorder) RecordCounters(counters []dtos.CounterDTO) error {
	data, err := json.Marshal(counters)
	if err != nil {
		m.logger.Error("Error marshaling JSON", err.Error())
		return err
	}

	err = m.recordRaw("/metrics/counters", data)
	if err != nil {
		m.logger.Error("Error posting counters", err.Error())
		return err
	}

	return nil
}

// RecordLatencies method submits latency metrics to the backend
func (m *HTTPMetricsRecorder) RecordLatencies(latencies []dtos.LatenciesDTO) error {
	data, err := json.Marshal(latencies)
	if err != nil {
		m.logger.Error("Error marshaling JSON", err.Error())
		return err
	}

	err = m.recordRaw("/metrics/times", data)
	if err != nil {
		m.logger.Error("Error posting latencies", err.Error())
		return err
	}

	return nil
}

// RecordGauge method submits gauge metrics to the backend
func (m *HTTPMetricsRecorder) RecordGauge(gauge dtos.GaugeDTO) error {
	data, err := json.Marshal(gauge)
	if err != nil {
		m.logger.Error("Error marshaling JSON", err.Error())
		return err
	}

	err = m.recordRaw("/metrics/gauge", data)
	if err != nil {
		m.logger.Error("Error posting gauges", err.Error())
		return err
	}

	return nil
}

// NewHTTPMetricsRecorder instantiates an HTTPMetricsRecorder
func NewHTTPMetricsRecorder(
	apikey string,
	cfg *conf.SplitSdkConfig,
	metadata dtos.Metadata,
	logger logging.LoggerInterface,
) *HTTPMetricsRecorder {
	_, eventsURL := getUrls(&cfg.Advanced)
	client := NewHTTPClient(apikey, cfg, eventsURL, splitio.Version, logger)
	return &HTTPMetricsRecorder{
		httpRecorderBase: httpRecorderBase{
			client:   client,
			metadata: metadata,
			logger:   logger,
		},
	}
}

// HTTPEventsRecorder is a struct responsible for submitting events bulks to the backend
type HTTPEventsRecorder struct {
	httpRecorderBase
}

// Record sends an array (or slice) of dtos.EventDTO to the backend
func (i *HTTPEventsRecorder) Record(events []dtos.EventDTO) error {
	data, err := json.Marshal(events)
	if err != nil {
		i.logger.Error("Error marshaling JSON", err.Error())
		return err
	}

	err = i.recordRaw("/events/bulk", data)
	if err != nil {
		i.logger.Error("Error posting events", err.Error())
		return err
	}

	return nil
}

// NewHTTPEventsRecorder instantiates an HTTPEventsRecorder
func NewHTTPEventsRecorder(
	apikey string,
	cfg *conf.SplitSdkConfig,
	metadata dtos.Metadata,
	logger logging.LoggerInterface,
) *HTTPEventsRecorder {
	_, eventsURL := getUrls(&cfg.Advanced)
	client := NewHTTPClient(apikey, cfg, eventsURL, splitio.Version, logger)
	return &HTTPEventsRecorder{
		httpRecorderBase: httpRecorderBase{
			client:   client,
			logger:   logger,
			metadata: metadata,
		},
	}
}
