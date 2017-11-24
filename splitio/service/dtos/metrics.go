package dtos

// LatenciesDTO struct mapping latencies post
type LatenciesDTO struct {
	MetricName string  `json:"name"`
	Latencies  []int64 `json:"latencies"`
}

// CounterDTO struct mapping counts post
type CounterDTO struct {
	MetricName string `json:"name"`
	Count      int64  `json:"delta"`
}

// GaugeDTO struct mapping gauges post
type GaugeDTO struct {
	MetricName string  `json:"name"`
	Gauge      float64 `json:"value"`
}
