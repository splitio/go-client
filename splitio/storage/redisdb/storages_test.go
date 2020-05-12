package redisdb

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
)

type LoggerInterface interface {
	Debug(msg ...interface{})
	Error(msg ...interface{})
	Info(msg ...interface{})
	Verbose(msg ...interface{})
	Warning(msg ...interface{})
	GetLog(key string) int
}

type MockedLogger struct {
	logs map[string]int
}

// NewMockedLogger creates a mocked logger to store logs
func NewMockedLogger() LoggerInterface {
	return &MockedLogger{
		logs: make(map[string]int),
	}
}

func (l *MockedLogger) Debug(msg ...interface{}) {
	messageList := make([]string, len(msg))
	for i, v := range msg {
		messageList[i] = fmt.Sprint(v)
	}
	var m string
	m = strings.Join(messageList, m)
	n, added := l.logs[m]
	if added == false {
		l.logs[m] = 1
	} else {
		l.logs[m] = n + 1
	}
}

func (l *MockedLogger) GetLog(key string) int {
	n, added := l.logs[key]
	if added == false {
		return -1
	}
	return n
}

func (l *MockedLogger) Error(msg ...interface{})   {}
func (l *MockedLogger) Info(msg ...interface{})    {}
func (l *MockedLogger) Verbose(msg ...interface{}) {}
func (l *MockedLogger) Warning(msg ...interface{}) {}

// PutMany bulk stores splits in redis
func PutMany(splitStorage *RedisSplitStorage, splits []dtos.SplitDTO, changeNumber int64) error {
	for _, split := range splits {
		keyToStore := strings.Replace(redisSplit, "{split}", split.Name, 1)
		raw, err := json.Marshal(split)
		if err != nil {
			continue
		}

		err = splitStorage.client.Set(keyToStore, raw, 0)
		if err != nil {
			return err
		}
	}
	err := splitStorage.client.Set(redisSplitTill, changeNumber, 0)
	if err != nil {
		return err
	}
	return nil
}
func TestRedisSplitStorage(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	prefixedClient, err := NewPrefixedRedisClient(&conf.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Password: "",
		Prefix:   "testPrefix",
	})
	if err != nil {
		t.Error(err.Error())
		return
	}

	splitStorage := NewRedisSplitStorage(prefixedClient, logger)

	err = PutMany(splitStorage, []dtos.SplitDTO{
		{Name: "split1", ChangeNumber: 1},
		{Name: "split2", ChangeNumber: 2},
		{Name: "split3", ChangeNumber: 3},
		{Name: "split4", ChangeNumber: 4},
	}, 123)
	if err != nil {
		t.Error(err)
	}

	s1 := splitStorage.Get("split1")
	if s1 == nil || s1.Name != "split1" || s1.ChangeNumber != 1 {
		t.Error("Incorrect split fetched/stored")
	}

	sns := splitStorage.SplitNames()
	snsSet := set.NewSet(sns[0], sns[1], sns[2], sns[3])
	if !snsSet.IsEqual(set.NewSet("split1", "split2", "split3", "split4")) {
		t.Error("Incorrect split names fetched")
		t.Error(sns)
	}

	if splitStorage.Till() != 123 {
		t.Error("Incorrect till")
		t.Error(splitStorage.Till())
	}

	err = PutMany(splitStorage, []dtos.SplitDTO{
		{
			Name: "split5",
			Conditions: []dtos.ConditionDTO{
				{
					MatcherGroup: dtos.MatcherGroupDTO{
						Matchers: []dtos.MatcherDTO{
							{
								UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
									SegmentName: "segment1",
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "split6",
			Conditions: []dtos.ConditionDTO{
				{
					MatcherGroup: dtos.MatcherGroupDTO{
						Matchers: []dtos.MatcherDTO{
							{
								UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
									SegmentName: "segment2",
								},
							},
							{
								UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
									SegmentName: "segment3",
								},
							},
						},
					},
				},
			},
		},
	}, 123)
	if err != nil {
		t.Error(err)
	}

	segmentNames := splitStorage.SegmentNames()
	hcSegments := set.NewSet("segment1", "segment2", "segment3")
	if !segmentNames.IsEqual(hcSegments) {
		t.Error("Incorrect segments retrieved")
		t.Error(segmentNames)
		t.Error(hcSegments)
	}

	allSplits := splitStorage.GetAll()
	allNames := set.NewSet()
	for _, split := range allSplits {
		allNames.Add(split.Name)
	}
	if !allNames.IsEqual(set.NewSet("split1", "split2", "split3", "split4", "split5", "split6")) {
		t.Error("GetAll returned incorrect splits")
	}

	// Test FetchMany before removing all the splits
	splitsToFetch := []string{"split1", "split2", "split3", "split4", "split5", "split6"}
	allSplitFetched := splitStorage.FetchMany(splitsToFetch)
	if len(allSplitFetched) != 6 {
		t.Error("It should return 6 splits")
	}
	for _, split := range splitsToFetch {
		if allSplitFetched[split] == nil {
			t.Error("It should not be nil")
		}
	}

	_, err = splitStorage.client.Del(
		"SPLITIO.split.split1",
		"SPLITIO.split.split2",
		"SPLITIO.split.split3",
		"SPLITIO.split.split4",
		"SPLITIO.split.split5",
		"SPLITIO.split.split6",
	)

	allSplits = splitStorage.GetAll()
	if len(allSplits) > 0 {
		t.Error("All splits should have been deleted")
		t.Error(allSplits)
	}

	allSplitFetched = splitStorage.FetchMany(splitsToFetch)
	if len(allSplitFetched) != 6 {
		t.Error("It should return 6 splits")
	}
	for _, split := range splitsToFetch {
		if allSplitFetched[split] != nil {
			t.Error("It should be nil")
		}
	}

	splitStorage.client.Del("key1", "key2")
}

func PutSegment(segmentStorage *RedisSegmentStorage, name string, segment *set.ThreadUnsafeSet, changeNumber int64) error {
	segmentKey := strings.Replace(redisSegment, "{segment}", name, 1)
	segmentTillKey := strings.Replace(redisSegmentTill, "{segment}", name, 1)
	_, err := segmentStorage.client.Del(segmentKey)
	if err != nil {
		return err
	}
	_, err = segmentStorage.client.SAdd(segmentKey, segment.List()...)
	if err != nil {
		return err
	}
	err = segmentStorage.client.Set(segmentTillKey, changeNumber, 0)
	if err != nil {
		return err
	}
	return nil
}

func TestSegmentStorage(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	prefixedClient, err := NewPrefixedRedisClient(&conf.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Password: "",
		Prefix:   "testPrefix",
	})
	if err != nil {
		t.Error(err.Error())
		return
	}

	segmentStorage := NewRedisSegmentStorage(prefixedClient, logger)

	err = PutSegment(segmentStorage, "segment1", set.NewSet("item1", "item2", "item3"), 123)
	if err != nil {
		t.Error(err)
	}
	err = PutSegment(segmentStorage, "segment2", set.NewSet("item4", "item5", "item6"), 124)
	if err != nil {
		t.Error(err)
	}

	segment1 := segmentStorage.Get("segment1")
	if segment1 == nil || !segment1.IsEqual(set.NewSet("item1", "item2", "item3")) {
		t.Error("Incorrect segment1")
		t.Error(segment1)
	}

	segment2 := segmentStorage.Get("segment2")
	if segment2 == nil || !segment2.IsEqual(set.NewSet("item4", "item5", "item6")) {
		t.Error("Incorrect segment2")
	}

	if segmentStorage.Till("segment1") != 123 || segmentStorage.Till("segment2") != 124 {
		t.Error("Incorrect till stored")
	}

	segmentStorage.client.Del("SPLITIO.segment.segment1", "SPLITIO.segment.segment1.till", "SPLITIO.segment.segment2", "SPLITIO.segment.segment2.till")
}

func TestImpressionStorage(t *testing.T) {
	logger := NewMockedLogger()
	prefixedClient, err := NewPrefixedRedisClient(&conf.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Password: "",
		Prefix:   "testPrefix",
	})
	if err != nil {
		t.Error(err.Error())
		return
	}
	metadata := &splitio.SdkMetadata{
		SDKVersion:  "go-test",
		MachineName: "instance123",
	}
	impressionStorage := NewRedisImpressionStorage(prefixedClient, metadata, logger)
	impressionStorage.client.Del(impressionStorage.redisKey)

	var ttl = impressionStorage.client.TTL(impressionStorage.redisKey)

	if ttl > impressionStorage.impressionsTTL {
		t.Error("TTL should be less than or equal to default")
	}

	var impression1 = storage.Impression{
		FeatureName:  "feature1",
		BucketingKey: "abc",
		ChangeNumber: 123,
		KeyName:      "key1",
		Label:        "label1",
		Time:         111,
		Treatment:    "on",
	}
	impressionStorage.LogImpressions([]storage.Impression{impression1})

	var impression2 = storage.Impression{
		FeatureName:  "feature2",
		BucketingKey: "abc",
		ChangeNumber: 123,
		KeyName:      "key1",
		Label:        "label1",
		Time:         111,
		Treatment:    "off",
	}
	impressionStorage.LogImpressions([]storage.Impression{impression2})

	impressions, _ := impressionStorage.PopN(2)

	if len(impressions) != 2 {
		t.Error("Incorrect number of impressions fetched")
	}

	expirationAdded := logger.GetLog("Proceeding to set expiration for: " + impressionStorage.redisKey)

	if expirationAdded != 1 {
		t.Error("It should added expiration only once")
	}

	var i1 = impressions[0]
	if i1.FeatureName != impression1.FeatureName {
		t.Error("Wrong Impression Stored, actual:", i1.FeatureName, " expected: ", impression1.FeatureName)
	}

	var i2 = impressions[1]
	if i2.FeatureName != impression2.FeatureName {
		t.Error("Wrong Impression Stored, actual:", i2.FeatureName, " expected: ", impression2.FeatureName)
	}
}

func TestMetricsStorage(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	prefixedClient, err := NewPrefixedRedisClient(&conf.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Password: "",
		Prefix:   "testPrefix",
	})
	if err != nil {
		t.Error(err.Error())
		return
	}
	metadata := &splitio.SdkMetadata{
		SDKVersion:  "go-test",
		MachineName: "instance123",
	}
	metricsStorage := NewRedisMetricsStorage(prefixedClient, metadata, logger)

	// Gauges

	metricsStorage.PutGauge("g1", 3.345)
	metricsStorage.PutGauge("g2", 4.456)
	val, _ := metricsStorage.client.Exists(
		"SPLITIO/go-test/instance123/gauge.g1",
		"SPLITIO/go-test/instance123/gauge.g2",
	)
	if val != 2 {
		t.Error("Keys or stored in an incorrect format")
	}

	metricsStorage.client.Del("SPLITIO/go-test/instance123/gauge.g1", "SPLITIO/go-test/instance123/gauge.g2")

	// Latencies
	metricsStorage.IncLatency("m1", 13)
	metricsStorage.IncLatency("m1", 13)
	metricsStorage.IncLatency("m1", 13)
	metricsStorage.IncLatency("m1", 1)
	metricsStorage.IncLatency("m1", 1)
	metricsStorage.IncLatency("m2", 1)
	metricsStorage.IncLatency("m2", 2)

	val, _ = metricsStorage.client.Exists(
		"SPLITIO/go-test/instance123/latency.m1.bucket.13",
		"SPLITIO/go-test/instance123/latency.m1.bucket.1",
		"SPLITIO/go-test/instance123/latency.m2.bucket.1",
		"SPLITIO/go-test/instance123/latency.m2.bucket.2",
	)
	if val != 4 {
		t.Error("Keys or stored in an incorrect format")
	}

	metricsStorage.client.Del(
		"SPLITIO/go-test/instance123/latency.m1.bucket.13",
		"SPLITIO/go-test/instance123/latency.m1.bucket.1",
		"SPLITIO/go-test/instance123/latency.m2.bucket.1",
		"SPLITIO/go-test/instance123/latency.m2.bucket.2",
	)

	// Counters
	metricsStorage.IncCounter("count1")
	metricsStorage.IncCounter("count1")
	metricsStorage.IncCounter("count1")
	metricsStorage.IncCounter("count2")
	metricsStorage.IncCounter("count2")
	metricsStorage.IncCounter("count2")
	metricsStorage.IncCounter("count2")
	metricsStorage.IncCounter("count2")
	metricsStorage.IncCounter("count2")

	val, _ = metricsStorage.client.Exists(
		"SPLITIO/go-test/instance123/count.count1",
		"SPLITIO/go-test/instance123/count.count2",
	)
	if val != 2 {
		t.Error("Incorrect counter keys stored in redis")
	}

	metricsStorage.client.Del(
		"SPLITIO/go-test/instance123/count.count1",
		"SPLITIO/go-test/instance123/count.count2",
	)
}

func TestTrafficTypeStorage(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	prefixedClient, err := NewPrefixedRedisClient(&conf.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 1,
		Password: "",
		Prefix:   "testPrefix",
	})
	if err != nil {
		t.Error(err.Error())
		return
	}
	ttStorage := NewRedisSplitStorage(prefixedClient, logger)

	ttStorage.client.Del("SPLITIO.trafficType.mytraffictype")
	ttStorage.client.Incr("SPLITIO.trafficType.mytraffictype")

	if !ttStorage.TrafficTypeExists("mytraffictype") {
		t.Error("Traffic type should exists")
	}

	ttStorage.client.Del("SPLITIO.trafficType.mytraffictype")
}
