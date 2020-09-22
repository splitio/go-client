package dependencytests

// This tests are stored here to break circular dependencies between evaluator
// & matchers as referenced directly within this test cases.

import (
	"reflect"
	"testing"

	"github.com/splitio/go-client/v6/splitio/engine"
	"github.com/splitio/go-client/v6/splitio/engine/evaluator"
	"github.com/splitio/go-client/v6/splitio/engine/grammar/matchers"
	"github.com/splitio/go-split-commons/v2/dtos"
	"github.com/splitio/go-split-commons/v2/storage/mutexmap"
	"github.com/splitio/go-toolkit/v3/injection"
	"github.com/splitio/go-toolkit/v3/logging"
)

type mockEvaluator struct {
	t                    *testing.T
	expectedBucketingKey string
}

func (e *mockEvaluator) EvaluateDependency(
	key string,
	bucketingKey *string,
	feature string,
	attributes map[string]interface{},
) string {
	var ok bool
	switch e.expectedBucketingKey {
	case "":
		if bucketingKey == nil {
			ok = true
		}
	default:
		if e.expectedBucketingKey == *bucketingKey {
			ok = true
		}
	}

	if !ok {
		e.t.Error("Incorrect bucketing key received")
	}

	return "on"
}

func TestDependencyMatcher(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
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

	splitStorage := mutexmap.NewMMSplitStorage()
	splitStorage.PutMany([]dtos.SplitDTO{
		{
			Name: "feature1",
			Conditions: []dtos.ConditionDTO{
				{
					ConditionType: "WHITELIST",
					Label:         "SomeLabel",
					MatcherGroup: dtos.MatcherGroupDTO{
						Combiner: "AND",
						Matchers: []dtos.MatcherDTO{
							{
								MatcherType: "ALL_KEYS",
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
		},
		{
			Name: "feature2",
			Conditions: []dtos.ConditionDTO{
				{
					ConditionType: "WHITELIST",
					Label:         "SomeLabel",
					MatcherGroup: dtos.MatcherGroupDTO{
						Combiner: "AND",
						Matchers: []dtos.MatcherDTO{
							{
								MatcherType: "WHITELIST",
								Whitelist: &dtos.WhitelistMatcherDataDTO{
									Whitelist: []string{"VAL1"},
								},
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
			DefaultTreatment: "off",
		},
	}, 1)
	segmentStorage := mutexmap.NewMMSegmentStorage()

	ctx := injection.NewContext()
	ctx.AddDependency(
		"evaluator",
		evaluator.NewEvaluator(
			splitStorage,
			segmentStorage,
			engine.NewEngine(logger),
			logger,
		),
	)

	matcher, err := matchers.BuildMatcher(dto, ctx, logger)
	if err != nil {
		t.Error("There should be no errors when building the matcher")
		t.Error(err)
	}

	matcherType := reflect.TypeOf(matcher).String()
	if matcherType != "*matchers.DependencyMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.DependencyMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", map[string]interface{}{"value": "something"}, nil) {
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
	matcher, err = matchers.BuildMatcher(dto, ctx, logger)
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

func TestDependencyMatcherWithBucketingKey(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
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

	splitStorage := mutexmap.NewMMSplitStorage()
	splitStorage.PutMany([]dtos.SplitDTO{
		{
			Name: "feature1",
			Conditions: []dtos.ConditionDTO{
				{
					ConditionType: "WHITELIST",
					Label:         "SomeLabel",
					MatcherGroup: dtos.MatcherGroupDTO{
						Combiner: "AND",
						Matchers: []dtos.MatcherDTO{
							{
								MatcherType: "ALL_KEYS",
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
		},
		{
			Name: "feature2",
			Conditions: []dtos.ConditionDTO{
				{
					ConditionType: "WHITELIST",
					Label:         "SomeLabel",
					MatcherGroup: dtos.MatcherGroupDTO{
						Combiner: "AND",
						Matchers: []dtos.MatcherDTO{
							{
								MatcherType: "WHITELIST",
								Whitelist: &dtos.WhitelistMatcherDataDTO{
									Whitelist: []string{"VAL1"},
								},
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
			DefaultTreatment: "off",
		},
	}, 1)

	ctx := injection.NewContext()
	ctx.AddDependency("evaluator", &mockEvaluator{expectedBucketingKey: "bucketingKey_1", t: t})

	matcher, err := matchers.BuildMatcher(dto, ctx, logger)
	if err != nil {
		t.Error("There should be no errors when building the matcher")
		t.Error(err)
	}

	bucketingKey := "bucketingKey_1"
	matcher.Match("asd", map[string]interface{}{"value": "something"}, &bucketingKey)

	ctx.AddDependency("evaluator", &mockEvaluator{expectedBucketingKey: "", t: t})
	matcher.Match("asd", map[string]interface{}{"value": "something"}, nil)

}
