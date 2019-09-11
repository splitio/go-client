package matchers

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/injection"
	"github.com/splitio/go-toolkit/logging"
	"reflect"
	"testing"
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
	segmentStorage.Put("segmentito", segmentKeys, 123)

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

	segmentStorage.Remove("segmentito")
	if matcher.Match("item1", nil, nil) {
		t.Error("Should return false for a nonexistent segment")
	}
}
