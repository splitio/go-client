package engine

import (
	"math"
	"testing"

	"github.com/splitio/go-client/splitio/engine/grammar"
	"github.com/splitio/go-client/splitio/engine/hash"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
)

func TestProperHashFunctionIsUsed(t *testing.T) {
	eng := Engine{}

	murmurHash := hash.Murmur3_32([]byte("SOME_TEST"), 12345)
	murmurBucket := int(math.Abs(float64(murmurHash%100)) + 1)
	if murmurBucket != eng.calculateBucket(2, "SOME_TEST", 12345) {
		t.Error("Incorrect hash!")
	}

	legacyHash := hash.Legacy([]byte("SOME_TEST"), 12345)
	legacyBucket := int(math.Abs(float64(legacyHash%100)) + 1)
	if legacyBucket != eng.calculateBucket(1, "SOME_TEST", 12345) {
		t.Error("Incorrect hash!")
	}
}

func TestProperHashFunctionIsUsedWithConstants(t *testing.T) {
	eng := Engine{}

	murmurHash := hash.Murmur3_32([]byte("SOME_TEST"), 12345)
	murmurBucket := int(math.Abs(float64(murmurHash%100)) + 1)
	if murmurBucket != eng.calculateBucket(grammar.SplitAlgoMurmur, "SOME_TEST", 12345) {
		t.Error("Incorrect hash!")
	}

	legacyHash := hash.Legacy([]byte("SOME_TEST"), 12345)
	legacyBucket := int(math.Abs(float64(legacyHash%100)) + 1)
	if legacyBucket != eng.calculateBucket(grammar.SplitAlgoLegacy, "SOME_TEST", 12345) {
		t.Error("Incorrect hash!")
	}
}

func TestTreatmentOnTrafficAllocation1(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitDTO := dtos.SplitDTO{
		Algo:                  2,
		ChangeNumber:          123,
		DefaultTreatment:      "default",
		Killed:                false,
		Name:                  "split",
		Seed:                  1234,
		Status:                "ACTIVE",
		TrafficAllocation:     1,
		TrafficAllocationSeed: -1667452163,
		TrafficTypeName:       "tt1",
		Conditions: []dtos.ConditionDTO{
			{
				ConditionType: "ROLLOUT",
				Label:         "in segment all",
				MatcherGroup: dtos.MatcherGroupDTO{
					Combiner: "AND",
					Matchers: []dtos.MatcherDTO{
						{
							MatcherType:        "ALL_KEYS",
							Whitelist:          nil,
							Negate:             false,
							UserDefinedSegment: nil,
						},
					},
				},
				Partitions: []dtos.PartitionDTO{
					{
						Size:      100,
						Treatment: "on",
					},
				},
			},
		},
	}

	split := grammar.NewSplit(&splitDTO, nil, logger)

	eng := Engine{}
	eng.logger = logger
	treatment, _ := eng.DoEvaluation(split, "aaaaaaklmnbv", nil, nil)

	if *treatment == "default" {
		t.Error("It should not return default treatment.")
	}
}

func TestTreatmentOnTrafficAllocation99(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitDTO := dtos.SplitDTO{
		Algo:                  1,
		ChangeNumber:          123,
		DefaultTreatment:      "default",
		Killed:                false,
		Name:                  "split",
		Seed:                  1234,
		Status:                "ACTIVE",
		TrafficAllocation:     99,
		TrafficAllocationSeed: -1667452111935,
		TrafficTypeName:       "tt1",
		Conditions: []dtos.ConditionDTO{
			{
				ConditionType: "ROLLOUT",
				Label:         "in segment all",
				MatcherGroup: dtos.MatcherGroupDTO{
					Combiner: "AND",
					Matchers: []dtos.MatcherDTO{
						{
							MatcherType:        "ALL_KEYS",
							Whitelist:          nil,
							Negate:             false,
							UserDefinedSegment: nil,
						},
					},
				},
				Partitions: []dtos.PartitionDTO{
					{
						Size:      100,
						Treatment: "on",
					},
				},
			},
		},
	}

	split := grammar.NewSplit(&splitDTO, nil, logger)

	eng := Engine{}
	eng.logger = logger
	treatment, _ := eng.DoEvaluation(split, "aaaaaaklmnbv", nil, nil)

	if *treatment != "default" {
		t.Error("It should return default treatment.")
	}
}
