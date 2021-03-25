package matchers

import (
	"reflect"
	"testing"

	"github.com/splitio/go-split-commons/v3/dtos"
	"github.com/splitio/go-split-commons/v3/storage/inmemory/mutexmap"
	"github.com/splitio/go-toolkit/v4/datastructures/set"
	"github.com/splitio/go-toolkit/v4/injection"
	"github.com/splitio/go-toolkit/v4/logging"
)

func TestInSegmentMatcher(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	dto := &dtos.MatcherDTO{
		MatcherType: "IN_SEGMENT",
		UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
			SegmentName: "segmentito",
		},
	}

	segmentKeys := set.NewSet()
	segmentKeys.Add("item1", "item2")

	segmentStorage := mutexmap.NewMMSegmentStorage()
	segmentStorage.Update("segmentito", segmentKeys, set.NewSet(), 123)

	ctx := injection.NewContext()
	ctx.AddDependency("segmentStorage", segmentStorage)

	matcher, err := BuildMatcher(dto, ctx, logger)
	if err != nil {
		t.Error("There should be no errors when building the matcher")
		t.Error(err)
	}

	matcherType := reflect.TypeOf(matcher).String()
	if matcherType != "*matchers.InSegmentMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.InSegmentMatcher and was %s", matcherType)
	}

	if !matcher.Match("item1", nil, nil) {
		t.Error("Should match a key present in the segment")
	}

	if matcher.Match("item7", nil, nil) {
		t.Error("Should not match a key not present in the segment")
	}

	segmentStorage.Update("segmentito", set.NewSet(), segmentKeys, 123)
	if matcher.Match("item1", nil, nil) {
		t.Error("Should return false for a nonexistent segment")
	}
}
