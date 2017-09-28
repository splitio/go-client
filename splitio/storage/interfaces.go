package storage

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/splitio/set"
)

// SplitStorage Interface should be implemented by all split storage storage forms
type SplitStorage interface {
	Get(splitName string) (dtos.SplitDTO, bool)
	PutMany(splits *[]dtos.SplitDTO)
	Remove(splitname string)
	Till() int64
}

// SegmentStorage Interface should be implemented by all segments storage storage forms
type SegmentStorage interface {
	Get(segmentName string) (set.ThreadUnsafeSet, bool)
	Put(name string, segment set.ThreadUnsafeSet)
	Remove(segmentName string)
	Till(segmentName string) int64
}

// ImpressionStorage Interface should be implemented by all impressions storage storage forms
type ImpressionStorage interface {
	Put(impression *dtos.ImpressionDTO)
	PopAll() *dtos.ImpressionDTO
}

// MetricsStorage Interface should be implemented by all metrics storage storage forms
type MetricsStorage interface {
	PutGauge(gauge *dtos.GaugeDTO)
	PutLatencies(latencies *[]dtos.LatenciesDTO)
	PutCounters(counters *[]dtos.CounterDTO)
	GetGauge() dtos.GaugeDTO
	GetLatencies() *[]dtos.LatenciesDTO
	GetCounters() *[]dtos.CounterDTO
}
