package matchers

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/injection"
	"reflect"
	"testing"
)

func TestMatcherConstruction(t *testing.T) {
	dto1 := dtos.MatcherDTO{
		Negate:      false,
		MatcherType: "ALL_KEYS",
		KeySelector: &dtos.KeySelectorDTO{
			Attribute:   nil,
			TrafficType: "something",
		},
	}

	matcher1, err := BuildMatcher(&dto1, nil)

	if err != nil {
		t.Error("Matcher construction shouldn't fail")
	}

	if reflect.TypeOf(matcher1).String() != "*matchers.AllKeysMatcher" {
		t.Errorf(
			"Incorrect matcher created, expected: \"*matchers.AllKeysMatcher\", received: \"%s\"",
			reflect.TypeOf(matcher1).String(),
		)
	}

	if matcher1.Negate() {
		t.Error("Matcher shouldn't be negated.")
	}

	if matcher1.base() != matcher1.(*AllKeysMatcher).base() {
		t.Error("base() should point to embedded base matcher struct")
	}

	dto2 := dtos.MatcherDTO{
		Negate:      true,
		MatcherType: "INVALID_MATCHER",
		KeySelector: &dtos.KeySelectorDTO{
			Attribute:   nil,
			TrafficType: "something",
		},
	}

	matcher2, err := BuildMatcher(&dto2, nil)

	if err == nil {
		t.Error("Matcher construction shoul have failed for invalid matcher")
	}

	if matcher2 != nil {
		t.Error("Builder should have returned nil as a result for an invalid matcher")
	}

	dto3 := dtos.MatcherDTO{
		Negate:      true,
		MatcherType: "ALL_KEYS",
		KeySelector: &dtos.KeySelectorDTO{
			Attribute:   nil,
			TrafficType: "something",
		},
	}
	ctx := injection.NewContext()
	ctx.AddDependency("key1", "sampleString")
	matcher3, err := BuildMatcher(&dto3, ctx)

	if err != nil {
		t.Error("There shouldn't have been any errors constructing the matcher")
	}

	if !matcher3.Negate() {
		t.Error("Matcher should be negated")
	}

	if matcher3.base().Context != ctx {
		t.Error("Context not properly received", matcher3.base().Context)
	}

	dep := matcher3.base().Dependency("key1")
	asString, ok := dep.(string)
	if !ok {
		t.Error("Conversion of string stored in context failed")
	}

	if asString != "sampleString" {
		t.Error("Recovered string doesn't match stored one")
	}
}
