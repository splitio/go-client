package evaluator

import (
	"testing"

	"github.com/splitio/go-client/v6/splitio/conf"
	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/datastructures/set"
	"github.com/splitio/go-toolkit/v5/logging"
)

type mockStorage struct{}

var mysplittest = &dtos.SplitDTO{
	Algo:                  2,
	ChangeNumber:          1494593336752,
	DefaultTreatment:      "off",
	Killed:                false,
	Name:                  "mysplittest",
	Seed:                  -1992295819,
	Status:                "ACTIVE",
	TrafficAllocation:     100,
	TrafficAllocationSeed: -285565213,
	TrafficTypeName:       "user",
	Configurations:        make(map[string]string),
	Conditions: []dtos.ConditionDTO{
		{
			ConditionType: "ROLLOUT",
			Label:         "default rule",
			MatcherGroup: dtos.MatcherGroupDTO{
				Combiner: "AND",
				Matchers: []dtos.MatcherDTO{
					{
						KeySelector: &dtos.KeySelectorDTO{
							TrafficType: "user",
							Attribute:   nil,
						},
						MatcherType:        "ALL_KEYS",
						Whitelist:          nil,
						Negate:             false,
						UserDefinedSegment: nil,
					},
				},
			},
			Partitions: []dtos.PartitionDTO{
				{
					Size:      0,
					Treatment: "on",
				}, {
					Size:      100,
					Treatment: "off",
				},
			},
		},
	},
}

var mysplittest2 = &dtos.SplitDTO{
	Algo:                  2,
	ChangeNumber:          1494593336752,
	DefaultTreatment:      "off",
	Killed:                false,
	Name:                  "mysplittest2",
	Seed:                  -1992295819,
	Status:                "ACTIVE",
	TrafficAllocation:     100,
	TrafficAllocationSeed: -285565213,
	TrafficTypeName:       "user",
	Configurations:        map[string]string{"on": "{\"color\": \"blue\",\"size\": 13}"},
	Conditions: []dtos.ConditionDTO{
		{
			ConditionType: "ROLLOUT",
			Label:         "default rule",
			MatcherGroup: dtos.MatcherGroupDTO{
				Combiner: "AND",
				Matchers: []dtos.MatcherDTO{
					{
						KeySelector: &dtos.KeySelectorDTO{
							TrafficType: "user",
							Attribute:   nil,
						},
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
				}, {
					Size:      0,
					Treatment: "off",
				},
			},
		},
	},
}

var mysplittest3 = &dtos.SplitDTO{
	Algo:                  2,
	ChangeNumber:          1494593336752,
	DefaultTreatment:      "killed",
	Killed:                true,
	Name:                  "mysplittest3",
	Seed:                  -1992295819,
	Status:                "ACTIVE",
	TrafficAllocation:     100,
	TrafficAllocationSeed: -285565213,
	TrafficTypeName:       "user",
	Configurations:        map[string]string{"on": "{\"color\": \"blue\",\"size\": 13}"},
	Conditions: []dtos.ConditionDTO{
		{
			ConditionType: "ROLLOUT",
			Label:         "default rule",
			MatcherGroup: dtos.MatcherGroupDTO{
				Combiner: "AND",
				Matchers: []dtos.MatcherDTO{
					{
						KeySelector: &dtos.KeySelectorDTO{
							TrafficType: "user",
							Attribute:   nil,
						},
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
				}, {
					Size:      0,
					Treatment: "off",
				},
			},
		},
	},
}

var mysplittest4 = &dtos.SplitDTO{
	Algo:                  2,
	ChangeNumber:          1494593336752,
	DefaultTreatment:      "killed",
	Killed:                true,
	Name:                  "mysplittest4",
	Seed:                  -1992295819,
	Status:                "ACTIVE",
	TrafficAllocation:     100,
	TrafficAllocationSeed: -285565213,
	TrafficTypeName:       "user",
	Configurations:        map[string]string{"on": "{\"color\": \"blue\",\"size\": 13}", "killed": "{\"color\": \"orange\",\"size\": 13}"},
	Conditions: []dtos.ConditionDTO{
		{
			ConditionType: "ROLLOUT",
			Label:         "default rule",
			MatcherGroup: dtos.MatcherGroupDTO{
				Combiner: "AND",
				Matchers: []dtos.MatcherDTO{
					{
						KeySelector: &dtos.KeySelectorDTO{
							TrafficType: "user",
							Attribute:   nil,
						},
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
				}, {
					Size:      0,
					Treatment: "off",
				},
			},
		},
	},
}

func (s *mockStorage) Split(
	feature string,
) *dtos.SplitDTO {
	switch feature {
	default:
	case "mysplittest":
		return mysplittest
	case "mysplittest2":
		return mysplittest2
	case "mysplittest3":
		return mysplittest3
	case "mysplittest4":
		return mysplittest4
	}
	return nil
}
func (s *mockStorage) FetchMany(
	feature []string,
) map[string]*dtos.SplitDTO {
	splits := make(map[string]*dtos.SplitDTO)
	splits["mysplittest"] = mysplittest
	splits["mysplittest2"] = mysplittest2
	splits["mysplittest3"] = mysplittest3
	splits["mysplittest4"] = mysplittest4
	splits["mysplittest5"] = nil
	return splits
}
func (s *mockStorage) All() []dtos.SplitDTO                      { return make([]dtos.SplitDTO, 0) }
func (s *mockStorage) SegmentNames() *set.ThreadUnsafeSet        { return nil }
func (s *mockStorage) SplitNames() []string                      { return make([]string, 0) }
func (s *mockStorage) TrafficTypeExists(trafficType string) bool { return true }
func (s *mockStorage) ChangeNumber() (int64, error)              { return 0, nil }

func TestSplitWithoutConfigurations(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	evaluator := NewEvaluator(
		&mockStorage{},
		nil,
		nil,
		logger)

	key := "test"
	result := evaluator.EvaluateFeature(key, &key, "mysplittest", nil)

	if result.Treatment != "off" {
		t.Error("Wrong treatment result")
	}

	if result.Config != nil {
		t.Error("Unexpected configs")
	}
}

func TestSplitWithtConfigurations(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	evaluator := NewEvaluator(
		&mockStorage{},
		nil,
		nil,
		logger)

	key := "test"
	result := evaluator.EvaluateFeature(key, &key, "mysplittest2", nil)

	if result.Treatment != "on" {
		t.Error("Wrong treatment result")
	}

	if result.Config == nil && *result.Config != "{\"color\": \"blue\",\"size\": 13}" {
		t.Error("Unexpected configs")
	}
}

func TestSplitWithtConfigurationsButKilled(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	evaluator := NewEvaluator(
		&mockStorage{},
		nil,
		nil,
		logger)

	key := "test"
	result := evaluator.EvaluateFeature(key, &key, "mysplittest3", nil)

	if result.Treatment != "killed" {
		t.Error("Wrong treatment result")
	}

	if result.Config != nil {
		t.Error("Unexpected configs")
	}
}

func TestSplitWithConfigurationsButKilledWithConfigsOnDefault(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	evaluator := NewEvaluator(
		&mockStorage{},
		nil,
		nil,
		logger)

	key := "test"
	result := evaluator.EvaluateFeature(key, &key, "mysplittest4", nil)

	if result.Treatment != "killed" {
		t.Error("Wrong treatment result")
	}

	if result.Config == nil && *result.Config != "{\"color\": \"orange\",\"size\": 13}" {
		t.Error("Unexpected configs")
	}
}

func TestMultipleEvaluations(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	evaluator := NewEvaluator(
		&mockStorage{},
		nil,
		nil,
		logger)

	key := "test"
	splits := []string{"mysplittest", "mysplittest2", "mysplittest3", "mysplittest4", "mysplittest5"}
	result := evaluator.EvaluateFeatures(key, &key, splits, nil)

	if result.Evaluations["mysplittest"].Treatment != "off" {
		t.Error("Wrong treatment result")
	}
	if result.Evaluations["mysplittest"].Config != nil {
		t.Error("Unexpected configs")
	}

	if result.Evaluations["mysplittest2"].Treatment != "on" {
		t.Error("Wrong treatment result")
	}
	if result.Evaluations["mysplittest2"].Config == nil && *result.Evaluations["mysplittest2"].Config != "{\"color\": \"blue\",\"size\": 13}" {
		t.Error("Unexpected configs")
	}

	if result.Evaluations["mysplittest3"].Treatment != "killed" {
		t.Error("Wrong treatment result")
	}
	if result.Evaluations["mysplittest3"].Config != nil {
		t.Error("Unexpected configs")
	}

	if result.Evaluations["mysplittest4"].Treatment != "killed" {
		t.Error("Wrong treatment result")
	}
	if result.Evaluations["mysplittest4"].Config == nil && *result.Evaluations["mysplittest4"].Config != "{\"color\": \"orange\",\"size\": 13}" {
		t.Error("Unexpected configs")
	}

	if result.Evaluations["mysplittest5"].Treatment != "control" {
		t.Error("Wrong treatment result")
	}
	if result.Evaluations["mysplittest5"].Config != nil {
		t.Error("Unexpected configs")
	}

	if result.EvaluationTime <= 0 {
		t.Error("It should be greater than 0")
	}
}
