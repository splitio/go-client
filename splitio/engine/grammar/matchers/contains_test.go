package matchers

import (
	"reflect"
	"testing"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
)

func TestContainsString(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "value"
	dto := &dtos.MatcherDTO{
		MatcherType: "CONTAINS_STRING",
		Whitelist: &dtos.WhitelistMatcherDataDTO{
			Whitelist: []string{"abc", "def", "ghi"},
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
	if matcherType != "*matchers.ContainsStringMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.ContainsStringMatcher and was %s", matcherType)
	}

	if matcher.Match("asd", map[string]interface{}{"value": "zzz"}, nil) {
		t.Errorf("string without any of the substrings shouldn't match")
	}

	if matcher.Match("asd", map[string]interface{}{"value": ""}, nil) {
		t.Errorf("empty string shouldn't match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": "ppabc"}, nil) {
		t.Errorf("string containing one of the substrings should match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": "abcdefghimklsad"}, nil) {
		t.Errorf("string containing all of the substrings should match")
	}

}
