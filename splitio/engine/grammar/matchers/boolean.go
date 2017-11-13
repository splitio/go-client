package matchers

import (
	"reflect"
	"strconv"
	"strings"
)

// BooleanMatcher returns true if the value supplied can be interpreted as a boolean and is equal to the one stored
type BooleanMatcher struct {
	Matcher
	value *bool
}

// Match returns true if the value supplied can be interpreted as a boolean and is equal to the one stored
func (m *BooleanMatcher) Match(key string, attributes map[string]interface{}, bucketingKey *string) bool {
	matchingKey, err := m.matchingKey(key, attributes)
	if err != nil {
		return false
	}

	var asBool bool
	var ok bool
	switch reflect.TypeOf(matchingKey).Kind() {
	case reflect.String:
		m.logger.Error("TRYING AS STRING")
		asStr, ok := matchingKey.(string)
		if !ok {
			m.logger.Error("NOT A STRING!")
			return false
		}
		logger.Error("STRING IS ", asStr)
		asBool, err = strconv.ParseBool(strings.ToLower(asStr))
		if err != nil {
			return false
		}
	case reflect.Bool:
		m.logger.Error("TRYING AS BOOLEAN")
		asBool, ok = matchingKey.(bool)
		if !ok {
			return false
		}
	default:
		m.logger.Error("CANNOT USE TYPE ", reflect.TypeOf(matchingKey).String())
		return false
	}

	return m.value != nil && *m.value == asBool
}

// NewBooleanMatcher instantiates a new BooleanMatcher
func NewBooleanMatcher(negate bool, value *bool, attributeName *string) *BooleanMatcher {
	return &BooleanMatcher{
		Matcher: Matcher{
			negate:        negate,
			attributeName: attributeName,
		},
		value: value,
	}
}
