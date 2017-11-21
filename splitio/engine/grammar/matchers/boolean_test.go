package matchers

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
	"reflect"
	"testing"
)

func TestBooleanMatcherTrue(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "value"
	boolValue := true
	dto := &dtos.MatcherDTO{
		MatcherType: "EQUAL_TO_BOOLEAN",
		Boolean:     &boolValue,
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
	if matcherType != "*matchers.BooleanMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.BooleanMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", map[string]interface{}{"value": true}, nil) {
		t.Errorf("true bool should match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": "true"}, nil) {
		t.Errorf("\"true\" tringhould match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": "tRUe"}, nil) {
		t.Errorf("true string with mixed caps should match")
	}
}

func TestBooleanMatcherFalse(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "value"
	boolValue := false
	dto := &dtos.MatcherDTO{
		MatcherType: "EQUAL_TO_BOOLEAN",
		Boolean:     &boolValue,
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
	if matcherType != "*matchers.BooleanMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.BooleanMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", map[string]interface{}{"value": false}, nil) {
		t.Errorf("false bool should match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": "false"}, nil) {
		t.Errorf("\"false\" tringhould match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": "fALse"}, nil) {
		t.Errorf("fALse bool should match")
	}
}
