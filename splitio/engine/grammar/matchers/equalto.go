package matchers

import (
	"github.com/splitio/go-client/splitio/engine/grammar/matchers/datatypes"
)

// EqualToMatcher will match if two numbers or two datetimes are equal
type EqualToMatcher struct {
	Matcher
	ComparisonDataType string
	ComparisonValue    interface{}
}

// Match will match if the comparisonValue is equal to the matchingValue
func (m *EqualToMatcher) Match(key string, attributes map[string]interface{}, bucketingKey *string) bool {

	matchingRaw, err := m.matchingKey(key, attributes)
	if err != nil {
		return false
	}

	matchingValue, okMatching := matchingRaw.(int64)
	comparisonValue, okComparison := m.ComparisonValue.(int64)
	if !okMatching || !okComparison {
		return false
	}

	switch m.ComparisonDataType {
	case datatypes.Number:
		return matchingValue == comparisonValue
	case datatypes.Datetime:
		matchingValue = datatypes.ZeroTimeTS(matchingValue)
		comparisonValue = datatypes.ZeroTimeTS(comparisonValue)
		return matchingValue == comparisonValue
	default:
		return false
	}
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
