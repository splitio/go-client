package storage

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/splitio/set"
)

func indexOf(caca interface{}, callback func(item interface{}) bool) (int, bool) {
	switch reflect.TypeOf(caca).Kind() {
	case reflect.Slice:
		castedCaca := reflect.ValueOf(caca)
		for i := 0; i < castedCaca.Len(); i++ {
			if callback(castedCaca.Index(i).Interface()) {
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
		split, found := splitStorage.Get(splitName)
		if !found || split.Name != splitName || split.Algo != index {
			t.Error("Split not returned as expected")
		}
	}

	_, found := splitStorage.Get("nonexistant_split")
	if found {
		t.Error("Nil expected but split returned")
	}

	splitStorage.Remove("SomeSplit_7")
	for index := 0; index < 10; index++ {
		splitName := fmt.Sprintf("SomeSplit_%d", index)
		split, found := splitStorage.Get(splitName)
		if index == 7 {
			if found {
				t.Error("Split Should have been removed and is present")
			}
		} else {
			if !found || split.Name != splitName || split.Algo != index {
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
		segment, exists := segmentStorage.Get(segmentName)
		if !exists {
			t.Errorf("%s should exist in storage and it doesn't.", segmentName)
		}

		for _, element := range segments[i] {
			if !segment.Has(element) {
				t.Errorf("%s should be part of set number %d and isn't.", element, i)
			}
		}
	}

	_, found := segmentStorage.Get("nonexistant_segment")
	if found {
		t.Error("Nil expected but segment returned")
	}

	segmentStorage.Remove("segmentito_1")
	for index := 0; index < 3; index++ {
		segmentName := fmt.Sprintf("segmentito_%d", index)
		_, found := segmentStorage.Get(segmentName)
		if index == 1 && found {
			t.Error("Segment Should have been removed and is present")
		}
		if index != 1 && !found {
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

	impressionsBak := impressionStorage.data
	impressions := impressionStorage.PopAll()
	if len(impressionStorage.data) > 0 {
		t.Error("Impressions not removed correctly from storage")
	}

	for key := range impressionsBak {
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

	gauges := metricsStorage.PopGauges()

	if len(gauges) != 3 {
		t.Error("Incorrect number of gauges popped")
	}

	metricsStorage.IncCounter("counter1")
	metricsStorage.IncCounter("counter1")
	metricsStorage.IncCounter("counter1")
	metricsStorage.IncCounter("counter2")

	if len(metricsStorage.counterData) != 2 {
		t.Error("Incorrect number of counters in storage")
	}

	counters := metricsStorage.PopCounters()
	if len(counters) != 2 {
		t.Error("Incorrect number of counters popped")
	}

	metricsStorage.IncLatency("http_io", 1)
	metricsStorage.IncLatency("http_io", 1)
	metricsStorage.IncLatency("http_io", 1)
	metricsStorage.IncLatency("http_io", 4)
	metricsStorage.IncLatency("disk_io", 7)

	if len(metricsStorage.latenciesData) != 2 {
		t.Error("Incorrect number of latencies in storage")
	}

	latencies := metricsStorage.PopLatencies()
	if len(latencies) != 2 {
		t.Error("Incorrect number of latencies popped")
	}
}
