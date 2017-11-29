package storage

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/datastructures/set"
)

// SplitStorageProducer should be implemented by structs that offer writing splits in storage
type SplitStorageProducer interface {
	PutMany(splits []dtos.SplitDTO, changeNumber int64)
	Remove(splitname string)
}

// SplitStorageConsumer should be implemented by structs that offer reading splits from storage
type SplitStorageConsumer interface {
	Get(splitName string) *dtos.SplitDTO
	Till() int64
	SplitNames() []string
	SegmentNames() *set.ThreadUnsafeSet
	GetAll() []dtos.SplitDTO
}

// SegmentStorageProducer interface should be implemented by all structs that offer writing segments
type SegmentStorageProducer interface {
	Put(name string, segment *set.ThreadUnsafeSet, changeNumber int64)
	Remove(segmentName string)
}

// SegmentStorageConsumer interface should be implemented by all structs that ofer reading segments
type SegmentStorageConsumer interface {
	Get(segmentName string) *set.ThreadUnsafeSet
	Till(segmentName string) int64
}

// ImpressionStorageProducer interface should be impemented by structs that accept incoming impressions
type ImpressionStorageProducer interface {
	Put(feature string, impression *dtos.ImpressionDTO)
}

// ImpressionStorageConsumer interface should be implemented by structs that offer popping impressions
type ImpressionStorageConsumer interface {
	PopAll() []dtos.ImpressionsDTO
}

// MetricsStorageProducer interface should be impemented by structs that accept incoming metrics
type MetricsStorageProducer interface {
	PutGauge(key string, gauge float64)
	IncLatency(metricName string, index int)
	IncCounter(key string)
}

// MetricsStorageConsumer interface should be implemented by structs that offer popping metrics
type MetricsStorageConsumer interface {
	PopGauges() []dtos.GaugeDTO
	PopLatencies() []dtos.LatenciesDTO
	PopCounters() []dtos.CounterDTO
}

// --- Wide Interfaces

// SplitStorage wraps consumer & producer interfaces
type SplitStorage interface {
	SplitStorageProducer
	SplitStorageConsumer
}

// SegmentStorage wraps consumer and producer interfaces
type SegmentStorage interface {
	SegmentStorageProducer
	SegmentStorageConsumer
}

// ImpressionStorage wraps consumer & producer interfaces
type ImpressionStorage interface {
	ImpressionStorageConsumer
	ImpressionStorageProducer
}

// MetricsStorage wraps consumer and producer interfaces
type MetricsStorage interface {
	MetricsStorageConsumer
	MetricsStorageProducer
}
