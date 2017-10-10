package storage

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/datastructures/set"
)

func indexOf(array interface{}, callback func(item interface{}) bool) (int, bool) {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		castedArray := reflect.ValueOf(array)
		for i := 0; i < castedArray.Len(); i++ {
			if callback(castedArray.Index(i).Interface()) {
				return i, true
			}
		}
	}
	return 0, false
}

func TestMMSplitStorage(t *testing.T) {
	splitStorage := NewMMSplitStorage()
	splits := make([]dtos.SplitDTO, 10)
	for index := 0; index < 10; index++ {
		splits = append(splits, dtos.SplitDTO{
			Name: fmt.Sprintf("SomeSplit_%d", index),
			Algo: index,
		})
	}

	splitStorage.PutMany(&splits, 123)
	for index := 0; index < 10; index++ {
		splitName := fmt.Sprintf("SomeSplit_%d", index)
		split := splitStorage.Get(splitName)
		if split == nil || split.Name != splitName || split.Algo != index {
			t.Error("Split not returned as expected")
		}
	}

	split := splitStorage.Get("nonexistant_split")
	if split != nil {
		t.Error("Nil expected but split returned")
	}

	splitStorage.Remove("SomeSplit_7")
	for index := 0; index < 10; index++ {
		splitName := fmt.Sprintf("SomeSplit_%d", index)
		split := splitStorage.Get(splitName)
		if index == 7 {
			if split != nil {
				t.Error("Split Should have been removed and is present")
			}
		} else {
			if split == nil || split.Name != splitName || split.Algo != index {
				t.Error("Split should not have been removed or modified and it was")
			}
		}
	}
}

func TestMMSegmentStorage(t *testing.T) {
	segments := make([][]string, 3)
	segments[0] = []string{"1a", "1b", "1c"}
	segments[1] = []string{"2a", "2b", "2c"}
	segments[2] = []string{"3a", "3b", "3c"}

	segmentStorage := NewMMSegmentStorage()
	for index, segment := range segments {
		setito := set.NewSet()
		for _, item := range segment {
			setito.Add(item)
		}
		segmentStorage.Put(fmt.Sprintf("segmentito_%d", index), setito, 123)
	}

	for i := 0; i < 3; i++ {
		segmentName := fmt.Sprintf("segmentito_%d", i)
		segment := segmentStorage.Get(segmentName)
		if segment == nil {
			t.Errorf("%s should exist in storage and it doesn't.", segmentName)
		}

		for _, element := range segments[i] {
			if !segment.Has(element) {
				t.Errorf("%s should be part of set number %d and isn't.", element, i)
			}
		}
	}

	segment := segmentStorage.Get("nonexistant_segment")
	if segment != nil {
		t.Error("Nil expected but segment returned")
	}

	segmentStorage.Remove("segmentito_1")
	for index := 0; index < 3; index++ {
		segmentName := fmt.Sprintf("segmentito_%d", index)
		segment := segmentStorage.Get(segmentName)
		if index == 1 && segment != nil {
			t.Error("Segment Should have been removed and is present")
		}
		if index != 1 && segment == nil {
			t.Error("Segment should not have been removed it has")
		}
	}
}

func TestImpressionStorage(t *testing.T) {
	impressionStorage := NewMMImpressionStorage()
	impressionStorage.Put("feature_a", &dtos.ImpressionDTO{
		KeyName:   "testKey1",
		Treatment: "on",
		Time:      123,
	})
	impressionStorage.Put("feature_a", &dtos.ImpressionDTO{
		KeyName:   "testKey2",
		Treatment: "off",
		Time:      124,
	})
	impressionStorage.Put("feature_b", &dtos.ImpressionDTO{
		KeyName:   "testKey1",
		Treatment: "off",
		Time:      125,
	})
	impressionStorage.Put("feature_b", &dtos.ImpressionDTO{
		KeyName:   "testKey2",
		Treatment: "off",
		Time:      126,
	})

	if len(impressionStorage.data) != 2 {
		t.Error("Incorrect number of features in impression storage")
	}

	impressionsForFeatureA := impressionStorage.data["feature_a"]
	if len(impressionsForFeatureA) != 2 {
		t.Error("Incorrect number of impressions for feature_a")
	}

	impressionsForFeatureB := impressionStorage.data["feature_b"]
	if len(impressionsForFeatureB) != 2 {
		t.Error("Incorrect number of impressions for feature_b")
	}

	impressionsBak := impressionStorage.data // Keep a copy of impressions in storage before calling PopAll()
	impressions := impressionStorage.PopAll()
	if len(impressionStorage.data) > 0 {
		t.Error("Impressions not removed correctly from storage")
	}

	for key := range impressionsBak {
		// Find the index of the impression in the struct returned by PopAll()
		index, found := indexOf(impressions, func(i interface{}) bool {
			imps, ok := i.(dtos.ImpressionsDTO)
			if ok && imps.TestName == key {
				return true
			}
			return false
		})
		if !found {
			t.Errorf("%s not should be in storage and isn't", key)
		} else {
			if len(impressions[index].KeyImpressions) != len(impressionsBak[key]) {
				t.Errorf("Incorrect number of impressions for %s", key)
			}
		}
	}
}

func TestMetricsStorage(t *testing.T) {
	metricsStorage := NewMMMetricsStorage()

	// Gauges
	metricsStorage.PutGauge("gauge1", 123.123)
	metricsStorage.PutGauge("gauge2", 456.456)
	metricsStorage.PutGauge("gauge3", 789.789)

	if len(metricsStorage.gaugeData) != 3 {
		t.Error("Incorrect number of gauges in storage")
	}

	gaugesBak := metricsStorage.gaugeData
	gauges := metricsStorage.PopGauges()

	if len(gauges) != 3 {
		t.Error("Incorrect number of gauges popped")
	}

	for key := range gaugesBak {
		index, found := indexOf(gauges, func(i interface{}) bool {
			orig, ok := i.(dtos.GaugeDTO)
			if ok && orig.MetricName == key {
				return true
			}
			return false
		})
		if !found {
			t.Errorf("Gauge %s should be present in storage and is not.", key)
		} else {
			if gauges[index].Gauge != gaugesBak[key] {
				t.Errorf("Value for gauge %s is incorrect", key)
			}
		}
	}

	metricsStorage.IncCounter("counter1")
	metricsStorage.IncCounter("counter1")
	metricsStorage.IncCounter("counter1")
	metricsStorage.IncCounter("counter2")

	if len(metricsStorage.counterData) != 2 {
		t.Error("Incorrect number of counters in storage")
	}

	countersBak := metricsStorage.counterData
	counters := metricsStorage.PopCounters()
	if len(counters) != 2 {
		t.Error("Incorrect number of counters popped")
	}

	for key := range countersBak {
		index, found := indexOf(counters, func(i interface{}) bool {
			orig, ok := i.(dtos.CounterDTO)
			if ok && orig.MetricName == key {
				return true
			}
			return false
		})
		if !found {
			t.Errorf("Counter %s should be present in storage and is not.", key)
		} else {
			if counters[index].Count != countersBak[key] {
				t.Errorf("Value for counter %s is incorrect", key)
			}
		}
	}

	metricsStorage.IncLatency("http_io", 1)
	metricsStorage.IncLatency("http_io", 1)
	metricsStorage.IncLatency("http_io", 1)
	metricsStorage.IncLatency("http_io", 4)
	metricsStorage.IncLatency("disk_io", 7)

	if len(metricsStorage.latenciesData) != 2 {
		t.Error("Incorrect number of latencies in storage")
	}

	latenciesBak := metricsStorage.latenciesData
	latencies := metricsStorage.PopLatencies()
	if len(latencies) != 2 {
		t.Error("Incorrect number of latencies popped")
	}

	for key := range latenciesBak {
		index, found := indexOf(latencies, func(i interface{}) bool {
			orig, ok := i.(dtos.LatenciesDTO)
			if ok && orig.MetricName == key {
				return true
			}
			return false
		})
		if !found {
			t.Errorf("Counter %s should be present in storage and is not.", key)
		} else {
			eq := true
			for li := range latenciesBak[key] {
				if latencies[index].Latencies[li] != latenciesBak[key][li] {
					eq = false
				}
			}
			if !eq {
				t.Errorf("Value for counter %s is incorrect", key)
			}
		}
	}

}
