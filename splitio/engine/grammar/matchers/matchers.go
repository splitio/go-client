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

	var attributeName *string
	if dto.KeySelector != nil {
		attributeName = dto.KeySelector.Attribute
	}

	switch dto.MatcherType {
	case MatcherTypeAllKeys:
		matcher = NewAllKeysMatcher(dto.Negate)

	case MatcherTypeEqualTo:
		if dto.UnaryNumeric == nil {
			return nil, errors.New("UnaryNumeric is required for EQUAL_TO matcher type")
		}
		matcher = NewEqualToMatcher(
			dto.Negate,
			dto.UnaryNumeric.Value,
			dto.UnaryNumeric.DataType,
			attributeName,
		)

	case MatcherTypeInSegment:
		if dto.UserDefinedSegment == nil {
			return nil, errors.New("UserDefinedSegment is required for IN_SEGMENT matcher type")
		}
		matcher = NewInSegmentMatcher(
			dto.Negate,
			dto.UserDefinedSegment.SegmentName,
			attributeName,
		)

	case MatcherTypeWhitelist:
		if dto.Whitelist == nil {
			return nil, errors.New("Whitelist is required for WHITELIST matcher type")
		}
		matcher = NewWhitelistMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			attributeName,
		)

	case MatcherTypeGreaterThanOrEqualTo:
		if dto.UnaryNumeric == nil {
			return nil, errors.New("UnaryNumeric is required for GREATER_THAN_OR_EQUAL_TO matcher type")
		}
		matcher = NewGreaterThanOrEqualToMatcher(
			dto.Negate,
			dto.UnaryNumeric.Value,
			dto.UnaryNumeric.DataType,
			attributeName,
		)

	case MatcherTypeLessThanOrEqualTo:
		if dto.UnaryNumeric == nil {
			return nil, errors.New("UnaryNumeric is required for LESS_THAN_OR_EQUAL_TO matcher type")
		}
		matcher = NewLessThanOrEqualToMatcher(
			dto.Negate,
			dto.UnaryNumeric.Value,
			dto.UnaryNumeric.DataType,
			attributeName,
		)

	case MatcherTypeBetween:
		if dto.Between == nil {
			return nil, errors.New("Between is required for BETWEEN matcher type")
		}
		matcher = NewBetweenMatcher(
			dto.Negate,
			dto.Between.Start,
			dto.Between.End,
			dto.Between.DataType,
			attributeName,
		)

	case MatcherTypeEqualToSet:
		if dto.Whitelist == nil {
			return nil, errors.New("Whitelist is required for EQUAL_TO_SET matcher type")
		}
		matcher = NewEqualToSetMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			attributeName,
		)

	case MatcherTypePartOfSet:
		if dto.Whitelist == nil {
			return nil, errors.New("Whitelist is required for PART_OF_SET matcher type")
		}
		matcher = NewPartOfSetMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			attributeName,
		)

	case MatcherTypeContainsAllOfSet:
		if dto.Whitelist == nil {
			return nil, errors.New("Whitelist is required for CONTAINS_ALL_OF_SET matcher type")
		}
		matcher = NewContainsAllOfSetMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			attributeName,
		)

	case MatcherTypeContainsAnyOfSet:
		if dto.Whitelist == nil {
			return nil, errors.New("Whitelist is required for CONTAINS_ANY_OF_SET matcher type")
		}
		matcher = NewContainsAnyOfSetMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			attributeName,
		)

	case MatcherTypeStartsWith:
		if dto.Whitelist == nil {
			return nil, errors.New("Whitelist is required for STARTS_WITH matcher type")
		}
		matcher = NewStartsWithMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			attributeName,
		)

	case MatcherTypeEndsWith:
		if dto.Whitelist == nil {
			return nil, errors.New("Whitelist is required for ENDS_WITH matcher type")
		}
		matcher = NewEndsWithMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			attributeName,
		)

	case MatcherTypeContainsString:
		if dto.Whitelist == nil {
			return nil, errors.New("Whitelist is required for CONTAINS_STRING matcher type")
		}
		matcher = NewContainsStringMatcher(
			dto.Negate,
			dto.Whitelist.Whitelist,
			attributeName,
		)

	case MatcherTypeInSplitTreatment:
		if dto.Dependency == nil {
			return nil, errors.New("Dependency is required for IN_SPLIT_TREATMENT matcher type")
		}
		matcher = NewDependencyMatcher(
			dto.Negate,
			dto.Dependency.Split,
			dto.Dependency.Treatments,
		)

	case MatcherTypeEqualToBoolean:
		if dto.Boolean == nil {
			return nil, errors.New("Boolean is required for EQUAL_TO_BOOLEAN matcher type")
		}
		matcher = NewBooleanMatcher(
			dto.Negate,
			dto.Boolean,
			attributeName,
		)

	case MatcherTypeMatchesString:
		if dto.String == nil {
			return nil, errors.New("String is required for MATCHES_STRING matcher type")
		}
		matcher = NewRegexMatcher(
			dto.Negate,
			*dto.String,
			attributeName,
		)

	default:
		return nil, errors.New("Matcher not found")
	}

	if ctx != nil {
		ctx.Inject(matcher.base())
	}

	return matcher, nil
}
