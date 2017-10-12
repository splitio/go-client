package matchers

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
	// MatcherTypeWhitelist string value
	MatcherTypeWhitelist = "WHITELIST"
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
	matchingKey(key string, attributes map[string]interface{}) (interface{}, error)
}

// Matcher struct with added logic that wraps around a DTO
type Matcher struct {
	*injection.Context
	negate        bool
	attributeName *string
}

// Negate returns whether this mather is negated or not
func (m *Matcher) Negate() bool {
	return m.negate
}

func (m *Matcher) matchingKey(key string, attributes map[string]interface{}) (interface{}, error) {
	if m.attributeName == nil {
		return key, nil
	}

	// Reaching this point means WE NEED attributes
	if attributes == nil {
		return nil, errors.New("Attribute required but no attributes provided")
	}

	attrValue, found := attributes[*m.attributeName]
	if !found {
		return nil, errors.New("Attribute required but not present in provided attribute map")
	}

	return attrValue, nil
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
		matcher = NewAllKeysMatcher(dto.Negate)

	case MatcherTypeEqualTo:
		matcher = NewEqualToMatcher(
			dto.Negate,
			dto.UnaryNumeric.Value,
			dto.UnaryNumeric.DataType,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeInSegment:
		matcher = NewInSegmentMatcher(
			dto.Negate,
			dto.UserDefinedSegment.SegmentName,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeWhitelist:
		matcher = NewWhitelistMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeGreaterThanOrEqualTo:
		matcher = NewGreaterThanOrEqualToMatcher(
			dto.Negate,
			dto.UnaryNumeric.Value,
			dto.UnaryNumeric.DataType,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeLessThanOrEqualTo:
		matcher = NewLessThanOrEqualToMatcher(
			dto.Negate,
			dto.UnaryNumeric.Value,
			dto.UnaryNumeric.DataType,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeBetween:
		matcher = NewBetweenMatcher(
			dto.Negate,
			dto.Between.Start,
			dto.Between.End,
			dto.Between.DataType,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeEqualToSet:
		matcher = NewEqualToSetMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			dto.KeySelector.Attribute,
		)

	case MatcherTypePartOfSet:
		matcher = NewPartOfSetMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeContainsAllOfSet:
		matcher = NewContainsAllOfSetMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeContainsAnyOfSet:
		matcher = NewContainsAnyOfSetMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeStartsWith:
		matcher = NewStartsWithMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeEndsWith:
		matcher = NewEndsWithMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeContainsString:
		matcher = NewContainsStringMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			dto.KeySelector.Attribute,
		)

	case MatcherTypeInSplitTreatment:
		matcher = NewDependencyMatcher(
			dto.Negate,
			dto.Dependency.Split,
			dto.Dependency.Treatments,
		)

	case MatcherTypeEqualToBoolean:
		matcher = NewBooleanMatcher(
			dto.Negate,
			dto.Boolean,
			dto.KeySelector.Attribute,
		)

	default:
		return nil, errors.New("Matcher not found")
	}

	if ctx != nil {
		ctx.Inject(matcher.base())
	}

	return matcher, nil
}
