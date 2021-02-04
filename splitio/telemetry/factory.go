package telemetry

import (
	"sync"
	"sync/atomic"
)

// FactoryTelemetryFacade keeps track of cache-related metrics
type FactoryTelemetryFacade struct {
	factories      map[string]int64
	burTimeouts    int64
	nonReadyUsages int64
	mutex          sync.RWMutex
}

// NewFactoryTelemetryFacade builds new facade
func NewFactoryTelemetryFacade() FactoryTelemetry {
	return &FactoryTelemetryFacade{
		factories:      make(map[string]int64),
		burTimeouts:    0,
		nonReadyUsages: 0,
		mutex:          sync.RWMutex{},
	}
}

// RecordFactory stores factory
func (f *FactoryTelemetryFacade) RecordFactory(apikey string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	_, ok := f.factories[apikey]
	if !ok {
		f.factories[apikey] = 1
		return
	}
	f.factories[apikey]++
}

// RecordNonReadyUsage stores non ready usage
func (f *FactoryTelemetryFacade) RecordNonReadyUsage() {
	atomic.AddInt64(&f.nonReadyUsages, 1)
}

// RecordBURTimeout stores timeouts
func (f *FactoryTelemetryFacade) RecordBURTimeout() {
	atomic.AddInt64(&f.burTimeouts, 1)
}

// GetActiveFactories gets active factories
func (f *FactoryTelemetryFacade) GetActiveFactories() int64 {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	var toReturn int64 = 0
	for _, activeFactories := range f.factories {
		toReturn += activeFactories
	}
	return toReturn
}

// GetRedundantActiveFactories gets redundant factories
func (f *FactoryTelemetryFacade) GetRedundantActiveFactories() int64 {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	var toReturn int64 = 0
	for _, activeFactory := range f.factories {
		if activeFactory > 1 {
			toReturn += activeFactory - 1
		}
	}
	return toReturn
}

// GetNonReadyUsages gets non ready usages
func (f *FactoryTelemetryFacade) GetNonReadyUsages() int64 { return atomic.LoadInt64(&f.nonReadyUsages) }

// GetBURTimeouts gets bur timeots
func (f *FactoryTelemetryFacade) GetBURTimeouts() int64 { return atomic.LoadInt64(&f.burTimeouts) }
