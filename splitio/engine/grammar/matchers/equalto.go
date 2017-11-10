package matchers

import (
	"fmt"
	"github.com/splitio/go-client/splitio/engine/grammar/matchers/datatypes"
	"reflect"
)

// EqualToMatcher will match if two numbers or two datetimes are equal
type EqualToMatcher struct {
	Matcher
	ComparisonDataType string
	ComparisonValue    int64
}

// Match will match if the comparisonValue is equal to the matchingValue
func (m *EqualToMatcher) Match(key string, attributes map[string]interface{}, bucketingKey *string) bool {

	matchingRaw, err := m.matchingKey(key, attributes)
	if err != nil {
		m.logger.Error("Error parsing matching key", err)
		return false
	}

	matchingValue, ok := matchingRaw.(int64)
	if !ok {
		var asInt int
		asInt, ok = matchingRaw.(int)
		if ok {
			matchingValue = int64(asInt)
		}
	}
	if !ok {
		m.base().logger.Error(
			"Error type-asserting matching key to an int",
			fmt.Sprintf("%s is a %s\n", matchingRaw, reflect.TypeOf(matchingRaw).String()),
		)
		return false
	}

	var comparisonValue int64
	switch m.ComparisonDataType {
	case datatypes.Number:
		comparisonValue = m.ComparisonValue
	case datatypes.Datetime:
		matchingValue = datatypes.ZeroTimeTS(matchingValue)
		comparisonValue = datatypes.ZeroTimeTS(m.ComparisonValue)
	default:
		m.base().logger.Error(fmt.Sprintf("Invalid comparison type %s\n", m.ComparisonDataType))
		return false
	}
	return matchingValue == comparisonValue
}

// NewEqualToMatcher returns a pointer to a new instance of EqualToMatcher
func NewEqualToMatcher(negate bool, cmpVal int64, cmpType string, attributeName *string) *EqualToMatcher {
	return &EqualToMatcher{
		Matcher: Matcher{
			negate:        negate,
			attributeName: attributeName,
		},
		ComparisonValue:    cmpVal,
		ComparisonDataType: cmpType,
	}
}
