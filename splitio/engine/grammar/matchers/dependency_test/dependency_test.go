package matchers

import (
	"github.com/splitio/go-client/splitio/engine"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/engine/grammar/matchers"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/util/logging"
	"github.com/splitio/go-toolkit/injection"
	"reflect"
	"testing"
)

func TestDependencyMatcher(t *testing.T) {
	attrName := "value"
	dto := &dtos.MatcherDTO{
		MatcherType: "IN_SPLIT_TREATMENT",
		Dependency: &dtos.DependencyMatcherDataDTO{
			Split:      "feature1",
			Treatments: []string{"on"},
		},
		KeySelector: &dtos.KeySelectorDTO{
			Attribute: &attrName,
		},
	}

	splitStorage := storage.NewMMSplitStorage()
	splitStorage.PutMany(&[]dtos.SplitDTO{
		dtos.SplitDTO{
			Name: "feature1",
			Conditions: []dtos.ConditionDTO{
				dtos.ConditionDTO{
					ConditionType: "WHITELIST",
					Label:         "SomeLabel",
					MatcherGroup: dtos.MatcherGroupDTO{
						Combiner: "AND",
						Matchers: []dtos.MatcherDTO{
							dtos.MatcherDTO{
								MatcherType: "ALL_KEYS",
							},
						},
					},
					Partitions: []dtos.PartitionDTO{
						dtos.PartitionDTO{
							Size:      100,
							Treatment: "on",
						},
					},
				},
			},
		},
		dtos.SplitDTO{
			Name: "feature2",
			Conditions: []dtos.ConditionDTO{
				dtos.ConditionDTO{
					ConditionType: "WHITELIST",
					Label:         "SomeLabel",
					MatcherGroup: dtos.MatcherGroupDTO{
						Combiner: "AND",
						Matchers: []dtos.MatcherDTO{
							dtos.MatcherDTO{
								MatcherType: "WHITELIST",
								Whitelist: &dtos.WhitelistMatcherDataDTO{
									Whitelist: []string{"VAL1"},
								},
							},
						},
					},
					Partitions: []dtos.PartitionDTO{
						dtos.PartitionDTO{
							Size:      100,
							Treatment: "on",
						},
					},
				},
			},
			DefaultTreatment: "off",
		},
	}, 1)
	segmentStorage := storage.NewMMSegmentStorage()

	ctx := injection.NewContext()
	ctx.AddDependency(
		"evaluator",
		evaluator.NewEvaluator(
			splitStorage,
			segmentStorage,
			engine.Engine{Logger: logging.NewLogger(&logging.LoggerOptions{})},
		),
	)

	matcher, err := matchers.BuildMatcher(dto, ctx)
	if err != nil {
		t.Error("There should be no errors when building the matcher")
		t.Error(err)
	}

	matcherType := reflect.TypeOf(matcher).String()
	if matcherType != "*matchers.DependencyMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.DependencyMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", map[string]interface{}{"value": "caca"}, nil) {
		t.Errorf("depends on all keys. should match!")
	}

	dto = &dtos.MatcherDTO{
		MatcherType: "IN_SPLIT_TREATMENT",
		Dependency: &dtos.DependencyMatcherDataDTO{
			Split:      "feature2",
			Treatments: []string{"on"},
		},
		KeySelector: &dtos.KeySelectorDTO{
			Attribute: &attrName,
		},
	}
	matcher, err = matchers.BuildMatcher(dto, ctx)
	if err != nil {
		t.Error("There should be no errors when building the matcher")
		t.Error(err)
	}

	if !matcher.Match("VAL1", map[string]interface{}{}, nil) {
		t.Errorf("depends on whitelist with VAL1. Should match")
	}

	if matcher.Match("VAL2", map[string]interface{}{}, nil) {
		t.Errorf("depends on whitelist with VAL1. passign VAL2 should fail")
	}
}
