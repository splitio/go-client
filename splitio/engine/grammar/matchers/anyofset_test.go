package matchers

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
	"reflect"
	"testing"
)

func TestAnyOfSetMatcher(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "setdata"
	dto := &dtos.MatcherDTO{
		MatcherType: "CONTAINS_ANY_OF_SET",
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
	if matcherType != "*matchers.ContainsAnyOfSetMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.ContainsAnyOfSetMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", map[string]interface{}{"setdata": []string{"one", "two", "three", "four"}}, nil) {
		t.Error("Matcher should match an equal set")
	}

	if !matcher.Match("asd", map[string]interface{}{"setdata": []string{"one", "two", "three", "four", "five"}}, nil) {
		t.Error("Matcher should match a superset")
	}

	if matcher.Match("asd", map[string]interface{}{"setdata": []string{}}, nil) {
		t.Error("Matcher should NOT match an empty set")
	}

	if !matcher.Match("asd", map[string]interface{}{"setdata": []string{"one", "two", "three"}}, nil) {
		t.Error("Matcher should match a subset")
	}

	if matcher.Match("asd", map[string]interface{}{"setdata": []string{"five", "six"}}, nil) {
		t.Error("Matcher hsould not match a non-intersecting set")
	}

}
