package grammar

import (
	"errors"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/injection"
)

const (
	// MatcherTypeAllKeys string value
	MatcherTypeAllKeys = "ALL_KEYS"
	// MatcherTypeInSegment string value
	MatcherTypeInSegment = "IN_SEGMENT"
	// MatcherTypeWhiteist string value
	MatcherTypeWhiteist = "WHITELIST"
	// MatcherTypeEqualTo string value
	MatcherTypeEqualTo = "EQUAL_TO"
	// MatcherTypeGreaterThanOrEqualTo string value
	MatcherTypeGreaterThanOrEqualTo = "GREATER_THAN_OR_EQUAL_TO"
	// MatcherTypeLessThanOrEqualTo string value
	MatcherTypeLessThanOrEqualTo = "LESS_THAN_OR_EQUAL_TO"
	// MatcherTypeBetween string value
	MatcherTypeBetween = "BETWEEN"
	// MatcherTypeEqualToSet string value
	MatcherTypeEqualToSet = "EQUAL_TO_SET"
	// MatcherTypePartOfSet string value
	MatcherTypePartOfSet = "PART_OF_SET"
	// MatcherTypeContainsAllOfSet string value
	MatcherTypeContainsAllOfSet = "CONTAINS_ALL_OF_SET"
	// MatcherTypeContainsAnyOfSet string value
	MatcherTypeContainsAnyOfSet = "CONTAINS_ANY_OF_SET"
	// MatcherTypeStartsWith string value
	MatcherTypeStartsWith = "STARTS_WITH"
	// MatcherTypeEndsWith string value
	MatcherTypeEndsWith = "ENDS_WITH"
	// MatcherTypeContainsString string value
	MatcherTypeContainsString = "CONTAINS_STRING"
	// MatcherTypeInSplitTreatment string value
	MatcherTypeInSplitTreatment = "IN_SPLIT_TREATMENT"
	// MatcherTypeEqualToBoolean string value
	MatcherTypeEqualToBoolean = "EQUAL_TO_BOOLEAN"
	// MatcherTypeMatchesString string value
	MatcherTypeMatchesString = "MATCHES_STRING"
)

// MatcherInterface should be implemented by all matchers
type MatcherInterface interface {
	Match(key string, attributes map[string]interface{}, bucketingKey *string) bool
	Negate() bool
	base() *Matcher // This method is used to return the embedded matcher when iterating over interfaces
}

// Matcher struct with added logic that wraps around a DTO
type Matcher struct {
	*injection.Context
	negate bool
}

// Negate returns whether this mather is negated or not
func (m *Matcher) Negate() bool {
	return m.negate
}

// matcher returns the matcher instance embbeded in structs
func (m *Matcher) base() *Matcher {
	return m
}

// BuildMatcher constructs the appropriate matcher based on the MatcherType attribute of the dto
func BuildMatcher(dto *dtos.MatcherDTO, ctx *injection.Context) (MatcherInterface, error) {
	var matcher MatcherInterface
	switch dto.MatcherType {
	case MatcherTypeAllKeys:
		matcher = &AllKeysMatcher{Matcher: Matcher{negate: dto.Negate}}
	default:
		return nil, errors.New("Matcher not found")
	}

	if ctx != nil {
		ctx.Inject(matcher.base())
	}

	return matcher, nil
}

// AllKeysMatcher matches any given key and set of attributes
type AllKeysMatcher struct {
	Matcher
	asd int
}

// Match implementation for AllKeysMatcher
func (m AllKeysMatcher) Match(key string, attributes map[string]interface{}, bucketingKey *string) bool {
	return true
}
