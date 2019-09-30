package engine

import (
	"encoding/csv"
	"io"
	"math"
	"os"
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
	treatment, _ := eng.DoEvaluation(split, "aaaaaaklmnbv", "aaaaaaklmnbv", nil)

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
	treatment, _ := eng.DoEvaluation(split, "aaaaaaklmnbv", "aaaaaaklmnbv", nil)

	if *treatment != "default" {
		t.Error("It should return default treatment.")
	}
}

type TreatmentResult struct {
	Key    string
	Result string
}

func parseCSV(file string) ([]TreatmentResult, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvr := csv.NewReader(f)

	var results []TreatmentResult
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return results, err
		}

		results = append(results, TreatmentResult{

			Key:    row[0],
			Result: row[1],
		})
	}
}

func TestTest(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitDTO := dtos.SplitDTO{
		Algo:                  2,
		ChangeNumber:          1550099287313,
		DefaultTreatment:      "on",
		Killed:                false,
		Name:                  "real_split",
		Seed:                  764645059,
		Status:                "ACTIVE",
		TrafficAllocation:     100,
		TrafficAllocationSeed: -1757484928,
		TrafficTypeName:       "user",
		Conditions: []dtos.ConditionDTO{
			{
				ConditionType: "ROLLOUT",
				Label:         "default rule",
				MatcherGroup: dtos.MatcherGroupDTO{
					Combiner: "AND",
					Matchers: []dtos.MatcherDTO{
						{
							KeySelector: &dtos.KeySelectorDTO{
								Attribute:   nil,
								TrafficType: "user",
							},
							MatcherType:        "ALL_KEYS",
							Negate:             false,
							UserDefinedSegment: nil,
							Whitelist:          nil,
							UnaryNumeric:       nil,
							Between:            nil,
							Boolean:            nil,
							String:             nil,
						},
					},
				},
				Partitions: []dtos.PartitionDTO{
					{
						Size:      50,
						Treatment: "on",
					},
					{
						Size:      50,
						Treatment: "off",
					},
				},
			},
		},
	}

	treatmentsResults, err := parseCSV("../../testdata/expected-treatments.csv")
	if err != nil {
		t.Error(err)
	}

	if len(treatmentsResults) == 0 {
		t.Error("Data was not added for testing consistency")
	}

	split := grammar.NewSplit(&splitDTO, nil, logger)

	eng := Engine{}
	eng.logger = logger

	for _, tr := range treatmentsResults {
		treatment, _ := eng.DoEvaluation(split, tr.Key, tr.Key, nil)

		if *treatment != tr.Result {
			t.Error("Checking expected treatment " + tr.Result + " for key: " + tr.Key)
		}
	}
}
