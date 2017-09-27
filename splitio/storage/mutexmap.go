package storage

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"sync"
)

// ** SPLIT STORAGE **

// MMSplitStorage struct contains is an in-memory implementation of split storage
type MMSplitStorage struct {
	data         map[string]dtos.SplitDTO
	mutex        *sync.RWMutex
	ChangeNumber int64
}

// NewMMSplitStorage instantiates a new MMSplitStorage
func NewMMSplitStorage() *MMSplitStorage {
	return &MMSplitStorage{
		data:  make(map[string]dtos.SplitDTO),
		mutex: &sync.RWMutex{},
	}

}

// Get retrieves a split from the MMSplitStorage
func (m *MMSplitStorage) Get(splitName string) (*dtos.SplitDTO, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	item, exists := m.data[splitName]
	return &item, exists
}

// PutMany bulk inserts splits into the in-memory storage
func (m *MMSplitStorage) PutMany(splits *[]dtos.SplitDTO) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, split := range *splits {
		m.data[split.Name] = split
	}
}

// Remove deletes a split from the in-memory storage
func (m *MMSplitStorage) Remove(splitName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.data, splitName)
}

// ** SEGMENT STORAGE **

// MMSegmentStorage contains is an in-memory implementation of segment storage
type MMSegmentStorage struct {
	data         map[string][]string
	mutex        *sync.RWMutex
	ChangeNumber int64
}

// NewMMSegmentStorage instantiates a new MMSegmentStorage
func NewMMSegmentStorage() *MMSegmentStorage {
	return &MMSegmentStorage{
		data:  make(map[string][]string),
		mutex: &sync.RWMutex{},
	}
}

// Get retrieves a segment from the in-memory storage
func (m *MMSegmentStorage) Get(segmentName string) (*[]string, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	item, exists := m.data[segmentName]
	return &item, exists
}

// Put adds a new segment to the in-memory storage
func (m *MMSegmentStorage) Put(name string, segment []string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[name] = segment
}

// Remove deletes a segment from the in-memmory storage
func (m *MMSegmentStorage) Remove(splitName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.data, splitName)
}

// ** IMPRESSIONS STORAGE **

//MMImpressionStorage contains an in-memory implementation of Impressions storage
type MMImpressionStorage struct {
	data  map[string][]dtos.ImpressionDTO
	mutex *sync.Mutex
}

// NewMMImpressionStorage instantiates an MMImpressionStorage
func NewMMImpressionStorage() *MMImpressionStorage {
	return &MMImpressionStorage{
		data:  make(map[string][]dtos.ImpressionDTO),
		mutex: &sync.Mutex{},
	}
}

// Put stores an impression for a feature in the in-memory storage
func (m *MMImpressionStorage) Put(feature string, impression *dtos.ImpressionDTO) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[feature] = append(m.data[feature], *impression)
}

// PopAll Returns and removes all the impressions currently stored
func (m *MMImpressionStorage) PopAll() []dtos.ImpressionsDTO {
	m.mutex.Lock()

	// After the function finishes, first replace the map with a fresh empty new one,
	// and then fully unlock the mutex
	defer func() {
		m.mutex.Unlock()
		m.data = make(map[string][]dtos.ImpressionDTO) // THIS deffered statement is executed first!
	}()

	impressions := make([]dtos.ImpressionsDTO, len(m.data))
	for testName, testImpressions := range m.data {
		impressions = append(impressions, dtos.ImpressionsDTO{
			TestName:       testName,
			KeyImpressions: testImpressions,
		})
	}

	return impressions
}

// ** Metrics Storage

// MMMetricsStorage contains an in-memory implementation of Metrics storage
type MMMetricsStorage struct {
	gaugeData      map[string]float64
	gaugeMutex     *sync.Mutex
	counterData    map[string]int64
	countersMutex  *sync.Mutex
	latenciesData  map[string][]int64
	latenciesMutex *sync.Mutex
}

// NewMMMetricsStorage instantiates a new MMMetricsStorage
func NewMMMetricsStorage() *MMMetricsStorage {
	return &MMMetricsStorage{
		counterData:    make(map[string]int64),
		countersMutex:  &sync.Mutex{},
		gaugeData:      make(map[string]float64),
		gaugeMutex:     &sync.Mutex{},
		latenciesData:  make(map[string][]int64),
		latenciesMutex: &sync.Mutex{},
	}
}

// PutGauge stores a new gauge value for a specific key
func (m *MMMetricsStorage) PutGauge(key string, gauge float64) {
	m.gaugeMutex.Lock()
	defer m.gaugeMutex.Unlock()
	m.gaugeData[key] = gauge
}

// PopGauges returns and deletes all gauges currently stored
func (m *MMMetricsStorage) PopGauges() []dtos.GaugeDTO {
	m.gaugeMutex.Lock()
	defer func() {
		m.gaugeData = make(map[string]float64)
		m.gaugeMutex.Unlock()
	}()

	gauges := make([]dtos.GaugeDTO, len(m.gaugeData))
	for key, gauge := range m.gaugeData {
		gauges = append(gauges, dtos.GaugeDTO{
			MetricName: key,
			Gauge:      gauge,
		})
	}
	return gauges
}

// IncCounter increments the counter for a specific key. It initializes it in 1 if it doesn't exist when this function
// is called.
func (m *MMMetricsStorage) IncCounter(key string) {
	m.countersMutex.Lock()
	defer m.countersMutex.Unlock()
	_, exists := m.counterData[key]
	if !exists {
		m.counterData[key] = 1
	} else {
		m.counterData[key]++
	}
}

// PopCounters returns and deletes all the counters stored
func (m *MMMetricsStorage) PopCounters() []dtos.CounterDTO {
	m.countersMutex.Lock()
	defer func() {
		m.counterData = make(map[string]int64)
		m.countersMutex.Unlock()
	}()

	counters := make([]dtos.CounterDTO, len(m.counterData))
	for key, counter := range m.counterData {
		counters = append(counters, dtos.CounterDTO{
			MetricName: key,
			Count:      counter,
		})
	}
	return counters
}

// IncLatency increments the latency for a specific key and bucket. If the key doesn't exist it's initialized to
// an empty array of 23 items.
func (m *MMMetricsStorage) IncLatency(metricName string, index int) {
	if index < 0 || index > 22 {
		return
	}
	m.latenciesMutex.Lock()
	defer m.latenciesMutex.Unlock()
	_, exists := m.latenciesData[metricName]
	if !exists {
		m.latenciesData[metricName] = make([]int64, 23)
		m.latenciesData[metricName][index] = 1
	} else {
		m.latenciesData[metricName][index]++
	}
}

// PopLatencies Returns and delete all the latencies currently stored
func (m *MMMetricsStorage) PopLatencies() []dtos.LatenciesDTO {
	m.latenciesMutex.Lock()
	defer func() {
		m.latenciesData = make(map[string][]int64)
		m.latenciesMutex.Unlock()
	}()

	latencies := make([]dtos.LatenciesDTO, len(m.latenciesData))
	for key, latency := range m.latenciesData {
		latencies = append(latencies, dtos.LatenciesDTO{
			Latencies:  latency,
			MetricName: key,
		})
	}
	return latencies
}
