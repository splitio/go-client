package matchers

import (
	"github.com/splitio/go-client/splitio/engine/grammar/matchers/datatypes"
)

// GreaterThanOrEqualToMatcher will match if two numbers or two datetimes are equal
type GreaterThanOrEqualToMatcher struct {
	Matcher
	ComparisonDataType string
	ComparisonValue    interface{}
}

// Match will match if the comparisonValue is greater than or equal to the matchingValue
func (m *GreaterThanOrEqualToMatcher) Match(key string, attributes map[string]interface{}, bucketingKey *string) bool {

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
	case datatypes.Datetime:
		matchingValue = datatypes.ZeroTimeTS(matchingValue)
		comparisonValue = datatypes.ZeroTimeTS(comparisonValue)
	default:
		return false
	}

	return matchingValue >= comparisonValue
}

// NewGreaterThanOrEqualToMatcher returns a pointer to a new instance of GreaterThanOrEqualToMatcher
func NewGreaterThanOrEqualToMatcher(negate bool, cmpVal int64, cmpType string, attributeName *string) *GreaterThanOrEqualToMatcher {
	return &GreaterThanOrEqualToMatcher{
		Matcher: Matcher{
			negate:        negate,
			attributeName: attributeName,
		},
		ComparisonValue:    cmpVal,
		ComparisonDataType: cmpType,
	}
}
