package matchers

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
	"reflect"
	"testing"
)

func TestEqualToSetMatcher(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "setdata"
	dto := &dtos.MatcherDTO{
		MatcherType: "EQUAL_TO_SET",
		Whitelist: &dtos.WhitelistMatcherDataDTO{
			Whitelist: []string{"one", "two", "three", "four"},
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
	if matcherType != "*matchers.EqualToSetMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.EqualToSetMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", map[string]interface{}{"setdata": []string{"one", "two", "three", "four"}}, nil) {
		t.Error("Matcher an equal set")
	}

	if matcher.Match("asd", map[string]interface{}{"setdata": []string{"one", "two", "three", "four", "five"}}, nil) {
		t.Error("Matcher should NOT match a superset")
	}

	if matcher.Match("asd", map[string]interface{}{"setdata": []string{"one", "two", "three"}}, nil) {
		t.Error("Matcher should not match a subset")
	}
}
