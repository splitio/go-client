package matchers

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
	"reflect"
	"testing"
)

func TestEndsWith(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "value"
	dto := &dtos.MatcherDTO{
		MatcherType: "ENDS_WITH",
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
	if matcherType != "*matchers.EndsWithMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.EndsWithMatcher and was %s", matcherType)
	}

	if matcher.Match("asd", map[string]interface{}{"value": "zzz"}, nil) {
		t.Errorf("string without any of the suffixes shouldn't match")
	}

	if matcher.Match("asd", map[string]interface{}{"value": ""}, nil) {
		t.Errorf("empty string shouldn't match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": "ppabc"}, nil) {
		t.Errorf("string containing one of the suffixes should match")
	}

	if matcher.Match("asd", map[string]interface{}{"value": "abcdefghimklsad"}, nil) {
		t.Errorf("string containing some substrings but not as suffixes should not match")
	}
}
