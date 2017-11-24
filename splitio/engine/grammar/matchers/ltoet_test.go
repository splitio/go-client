package matchers

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
	"reflect"
	"testing"
)

func TestLessThanOrEqualToMatcherInt(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "value"
	dto := &dtos.MatcherDTO{
		MatcherType: "LESS_THAN_OR_EQUAL_TO",
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
	if matcherType != "*matchers.LessThanOrEqualToMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.LessThanOrEqualToMatcher and was %s", matcherType)
	}

	if !matcher.Match("asd", map[string]interface{}{"value": 100}, nil) {
		t.Error("Equal should match")
	}

	if matcher.Match("asd", map[string]interface{}{"value": 500}, nil) {
		t.Error("Greater should not match")
	}

	if !matcher.Match("asd", map[string]interface{}{"value": 50}, nil) {
		t.Error("Lower should match")
	}
}

func TestLessThanOrEqualToMatcherDatetime(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "value"
	dto := &dtos.MatcherDTO{
		MatcherType: "LESS_THAN_OR_EQUAL_TO",
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
	if matcherType != "*matchers.LessThanOrEqualToMatcher" {
		t.Errorf("Incorrect matcher constructed. Should be *matchers.LessThanOrEqualToMatcher and was %s", matcherType)
	}

	attributes := make(map[string]interface{})
	attributes["value"] = int64(960293532000)
	if !matcher.Match("asd", attributes, nil) {
		t.Error("Equal should match")
	}

	attributes["value"] = int64(1275782400000)
	if matcher.Match("asd", attributes, nil) {
		t.Error("Greater should not match")
	}

	attributes["value"] = int64(293532000000)
	if !matcher.Match("asd", attributes, nil) {
		t.Error("Lower should match")
	}
}
