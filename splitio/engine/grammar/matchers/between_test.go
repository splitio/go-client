package matchers

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
	"reflect"
	"testing"
)

func TestBetweenMatcherInt(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "value"
	dto := &dtos.MatcherDTO{
		MatcherType: "BETWEEN",
		Between: &dtos.BetweenMatcherDataDTO{
			DataType: "NUMBER",
			Start:    int64(100),
			End:      int64(500),
		},
		KeySelector: &dtos.KeySelectorDTO{
			Attribute: &attrName,
		},
	}

	matcher, err := BuildMatcher(dto, nil, logger)
	if err != nil {
		t.Error("There should be no errors when building the matcher")
		t.Error(err)
	}

	matcherType := reflect.TypeOf(matcher).String()
	if matcherType != "*matchers.BetweenMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.BetweenMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", map[string]interface{}{"value": 100}, nil) {
		t.Error("Lower limit should match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": 500}, nil) {
		t.Error("Upper limit should match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": 250}, nil) {
		t.Error("Matcher should match an equal set")
	}

	if matcher.Match("asd", map[string]interface{}{"value": 99}, nil) {
		t.Error("Lower than lower limit should NOT match")
	}

	if matcher.Match("asd", map[string]interface{}{"value": 501}, nil) {
		t.Error("Upper than upper limit should NOT match")
	}
}

func TestBetweenMatcherDatetime(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "value"
	dto := &dtos.MatcherDTO{
		MatcherType: "BETWEEN",
		Between: &dtos.BetweenMatcherDataDTO{
			DataType: "DATETIME",
			Start:    int64(960293532000),  // 06/06/2000
			End:      int64(1275782400000), // 06/06/2010
		},
		KeySelector: &dtos.KeySelectorDTO{
			Attribute: &attrName,
		},
	}

	matcher, err := BuildMatcher(dto, nil, logger)
	if err != nil {
		t.Error("There should be no errors when building the matcher")
		t.Error(err)
	}

	matcherType := reflect.TypeOf(matcher).String()
	if matcherType != "*matchers.BetweenMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.BetweenMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", map[string]interface{}{"value": 960293532}, nil) {
		t.Error("Lower limit should match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": 1275782400}, nil) {
		t.Error("Upper limit should match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": 980293532}, nil) {
		t.Error("Should match between lower and upper")
	}

	if matcher.Match("asd", map[string]interface{}{"value": 900293532}, nil) {
		t.Error("Lower than lower limit should NOT match")
	}

	if matcher.Match("asd", map[string]interface{}{"value": 1375782400}, nil) {
		t.Error("Upper than upper limit should NOT match")
	}
}
