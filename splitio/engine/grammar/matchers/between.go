package matchers

import (
	"github.com/splitio/go-client/splitio/engine/grammar/matchers/datatypes"
)

// BetweenMatcher will match if two numbers or two datetimes are equal
type BetweenMatcher struct {
	Matcher
	ComparisonDataType   string
	LowerComparisonValue interface{}
	UpperComparisonValue interface{}
}

// Match will match if the matchingValue is between lowerComparisonValue and upperComparisonValue
func (m *BetweenMatcher) Match(key string, attributes map[string]interface{}, bucketingKey *string) bool {

	matchingRaw, err := m.matchingKey(key, attributes)
	if err != nil {
		return false
	}

	matchingValue, okMatching := matchingRaw.(int64)
	comparisonLower, okComparisonLower := m.LowerComparisonValue.(int64)
	comparisonUpper, okComparisonUpper := m.UpperComparisonValue.(int64)
	if !okMatching || !okComparisonLower || !okComparisonUpper {
		return false
	}

	switch m.ComparisonDataType {
	case datatypes.Number:
	case datatypes.Datetime:
		matchingValue = datatypes.ZeroTimeTS(matchingValue)
		comparisonLower = datatypes.ZeroTimeTS(comparisonLower)
		comparisonUpper = datatypes.ZeroTimeTS(comparisonUpper)
	default:
		return false
	}

	return matchingValue >= comparisonLower && matchingValue <= comparisonUpper
}

// NewBetweenMatcher returns a pointer to a new instance of BetweenMatcher
func NewBetweenMatcher(negate bool, lower int64, upper int64, cmpType string, attributeName *string) *BetweenMatcher {
	return &BetweenMatcher{
		Matcher: Matcher{
			negate:        negate,
			attributeName: attributeName,
		},
		LowerComparisonValue: lower,
		UpperComparisonValue: upper,
		ComparisonDataType:   cmpType,
	}
}
