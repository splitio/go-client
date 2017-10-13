package matchers

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"reflect"
	"testing"
)

func TestAllKeysMatcher(t *testing.T) {
	dto := &dtos.MatcherDTO{
		MatcherType: "ALL_KEYS",
	}

	matcher, err := BuildMatcher(dto, nil)
	if err != nil {
		t.Error("There should be no errors when building the matcher")
	}

	matcherType := reflect.TypeOf(matcher).String()
	if matcherType != "*matchers.AllKeysMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.AllKeysMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", nil, nil) {
		t.Error("Matcher should match ANY string")
	}
}
