package matchers

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"reflect"
	"testing"
)

func TestWhitelistMatcherAttr(t *testing.T) {
	attrName := "value"
	dto := &dtos.MatcherDTO{
		MatcherType: "WHITELIST",
		Whitelist: &dtos.WhitelistMatcherDataDTO{
			Whitelist: []string{"aaa", "bbb"},
		},
		KeySelector: &dtos.KeySelectorDTO{
			Attribute: &attrName,
		},
	}

	matcher, err := BuildMatcher(dto, nil)
	if err != nil {
		t.Error("There should be no errors when building the matcher")
		t.Error(err)
	}

	matcherType := reflect.TypeOf(matcher).String()
	if matcherType != "*matchers.WhitelistMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.WhitelistMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", map[string]interface{}{"value": "aaa"}, nil) {
		t.Error("Item in whitelist should match")
	}

	if matcher.Match("asd", map[string]interface{}{"value": "ccc"}, nil) {
		t.Error("Item NOT in whitelist should NOT match")
	}
}

func TestWhitelistMatcherKey(t *testing.T) {
	dto := &dtos.MatcherDTO{
		MatcherType: "WHITELIST",
		Whitelist: &dtos.WhitelistMatcherDataDTO{
			Whitelist: []string{"aaa", "bbb"},
		},
	}

	matcher, err := BuildMatcher(dto, nil)
	if err != nil {
		t.Error("There should be no errors when building the matcher")
		t.Error(err)
	}

	matcherType := reflect.TypeOf(matcher).String()
	if matcherType != "*matchers.WhitelistMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.WhitelistMatcher and was %s", matcherType)
	}

	if !matcher.Match("aaa", map[string]interface{}{"value": 1}, nil) {
		t.Error("Item in whitelist should match")
	}

	if matcher.Match("asd", map[string]interface{}{"value": 2}, nil) {
		t.Error("Item NOT in whitelist should NOT match")
	}
}
