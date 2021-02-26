package matchers

import (
	"reflect"
	"testing"

	"github.com/splitio/go-split-commons/dtos"
	"github.com/splitio/go-toolkit/logging"
)

func TestEqualToMatcherInt(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "value"
	dto := &dtos.MatcherDTO{
		MatcherType: "EQUAL_TO",
		UnaryNumeric: &dtos.UnaryNumericMatcherDataDTO{
			DataType: "NUMBER",
			Value:    int64(100),
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
	if matcherType != "*matchers.EqualToMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.EqualToMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", map[string]interface{}{"value": 100}, nil) {
		t.Error("Equal should match")
	}

	if matcher.Match("asd", map[string]interface{}{"value": 500}, nil) {
		t.Error("Greater should not match")
	}

	if matcher.Match("asd", map[string]interface{}{"value": 50}, nil) {
		t.Error("Lower should not match")
	}
}

func TestEqualToMatcherDatetime(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "value"
	dto := &dtos.MatcherDTO{
		MatcherType: "EQUAL_TO",
		UnaryNumeric: &dtos.UnaryNumericMatcherDataDTO{
			DataType: "DATETIME",
			Value:    int64(960293532000), // 06/06/2000
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
	if matcherType != "*matchers.EqualToMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.EqualToMatcher and was %s", matcherType)
	}

	attributes := make(map[string]interface{})
	attributes["value"] = int64(960293532)
	if !matcher.Match("asd", attributes, nil) {
		t.Error("Equal should match")
	}

	attributes["value"] = int64(1275782400)
	if matcher.Match("asd", attributes, nil) {
		t.Error("Greater should not match")
	}

	attributes["value"] = int64(293532000)
	if matcher.Match("asd", attributes, nil) {
		t.Error("Lower should not match")
	}
}
