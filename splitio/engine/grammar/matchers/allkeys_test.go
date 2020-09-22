package matchers

import (
	"reflect"
	"testing"

	"github.com/splitio/go-split-commons/v2/dtos"
	"github.com/splitio/go-toolkit/v3/logging"
)

func TestAllKeysMatcher(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	dto := &dtos.MatcherDTO{
		MatcherType: "ALL_KEYS",
	}

	matcher, err := BuildMatcher(dto, nil, logger)
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
