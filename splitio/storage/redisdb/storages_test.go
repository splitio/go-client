package redisdb

import (
	"testing"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
)

func TestRedisSplitStorage(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitStorage := NewRedisSplitStorage("localhost", 6379, 1, "", "testPrefix", logger)

	splitStorage.PutMany([]dtos.SplitDTO{
		dtos.SplitDTO{Name: "split1", ChangeNumber: 1},
		dtos.SplitDTO{Name: "split2", ChangeNumber: 2},
		dtos.SplitDTO{Name: "split3", ChangeNumber: 3},
		dtos.SplitDTO{Name: "split4", ChangeNumber: 4},
	}, 123)

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

	splitStorage.PutMany([]dtos.SplitDTO{
		dtos.SplitDTO{
			Name: "split5",
			Conditions: []dtos.ConditionDTO{
				dtos.ConditionDTO{
					MatcherGroup: dtos.MatcherGroupDTO{
						Matchers: []dtos.MatcherDTO{
							dtos.MatcherDTO{
								UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
									SegmentName: "segment1",
								},
							},
						},
					},
				},
			},
		},
		dtos.SplitDTO{
			Name: "split6",
			Conditions: []dtos.ConditionDTO{
				dtos.ConditionDTO{
					MatcherGroup: dtos.MatcherGroupDTO{
						Matchers: []dtos.MatcherDTO{
							dtos.MatcherDTO{
								UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
									SegmentName: "segment2",
								},
							},
							dtos.MatcherDTO{
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

	segmentNames := splitStorage.SegmentNames()
	segmentSet := set.NewSet(segmentNames[0], segmentNames[1], segmentNames[2])
	hcSegments := set.NewSet("segment1", "segment2", "segment3")
	if !segmentSet.IsEqual(hcSegments) {
		t.Error("Incorrect segments retrieved")
		t.Error(segmentSet)
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

	for _, name := range []string{"split1", "split2", "split3", "split4", "split5", "split6"} {
		splitStorage.Remove(name)
	}

	allSplits = splitStorage.GetAll()
	if len(allSplits) > 0 {
		t.Error("All splits should have been deleted")
		t.Error(allSplits)
	}
}

func TestSegmentStorage(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	segmentStorage := NewRedisSegmentStorage("localhost", 6379, 1, "", "testPrefix", logger)

	segmentStorage.Put("segment1", set.NewSet("item1", "item2", "item3"), 123)
	segmentStorage.Put("segment2", set.NewSet("item4", "item5", "item6"), 124)

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

	segmentStorage.Put("segment1", set.NewSet("item7"), 222)
	segment1 = segmentStorage.Get("segment1")
	if !segment1.IsEqual(set.NewSet("item7")) {
		t.Error("Segment 1 not overwritten correctly")
	}

	if segmentStorage.Till("segment1") != 222 {
		t.Error("segment 1 till not updated correctly")
	}

	segmentStorage.Remove("segment1")
	if segmentStorage.Get("segment1") != nil || segmentStorage.Till("segment1") != -1 {
		t.Error("Segment 1 and it's till value should have been removed")
		t.Error(segmentStorage.Get("segment1"))
		t.Error(segmentStorage.Till("segment1"))
	}
}

func TestImpressionStorage(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	impressionStorage := NewRedisImpressionStorage("localhost", 6379, 1, "", "testPrefix", "instance123", "go-test", logger)

	impressionStorage.Put("feature1", &dtos.ImpressionDTO{
		BucketingKey: "abc",
		ChangeNumber: 123,
		KeyName:      "key1",
		Label:        "label1",
		Time:         111,
		Treatment:    "on",
	})
	impressionStorage.Put("feature1", &dtos.ImpressionDTO{
		BucketingKey: "abc",
		ChangeNumber: 123,
		KeyName:      "key2",
		Label:        "label1",
		Time:         111,
		Treatment:    "off",
	})
	impressionStorage.Put("feature2", &dtos.ImpressionDTO{
		BucketingKey: "abc",
		ChangeNumber: 123,
		KeyName:      "key1",
		Label:        "label1",
		Time:         111,
		Treatment:    "off",
	})
	impressionStorage.Put("feature2", &dtos.ImpressionDTO{
		BucketingKey: "abc",
		ChangeNumber: 123,
		KeyName:      "key2",
		Label:        "label1",
		Time:         111,
		Treatment:    "on",
	})

	impressions := impressionStorage.PopAll()

	if len(impressions) != 2 {
		t.Error("Incorrect number of features with impressions fetched")
	}

	var feature1, feature2 dtos.ImpressionsDTO
	if impressions[0].TestName == "feature1" && impressions[1].TestName == "feature2" {
		feature1 = impressions[0]
		feature2 = impressions[1]
	} else if impressions[1].TestName == "feature1" && impressions[0].TestName == "feature2" {
		feature1 = impressions[1]
		feature2 = impressions[0]
	} else {
		t.Error("Incorrect impression testnames!")
		return
	}

	if len(feature1.KeyImpressions) != 2 {
		t.Error("Incorrect number of impressions fetched for feature1")
	}

	if len(feature2.KeyImpressions) != 2 {
		t.Error("Incorrect number of impressions fetched for feature2")
	}
}
