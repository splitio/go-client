package matchers

import (
	"fmt"
	"reflect"
	"regexp"
)

// RegexMatcher matches if the supplied key matches the split's regex
type RegexMatcher struct {
	Matcher
	regex string
}

// Match returns true if the supplied key matches the split's regex
func (m *RegexMatcher) Match(key string, attributes map[string]interface{}, bucketingKey *string) bool {
	matchingKey, err := m.matchingKey(key, attributes)
	if err != nil {
		m.logger.Error("Error parsing matching key: ", err)
		return false
	}

	conv, ok := matchingKey.(string)
	if !ok {
		m.logger.Error(
			"Incorrect type. Expected string and recieved ",
			reflect.TypeOf(matchingKey).String(),
		)
		return false
	}

	re := regexp.MustCompile(m.regex)
	return re.MatchString(conv)
}

// NewRegexMatcher returns a new instance to a RegexMatcher
func NewRegexMatcher(negate bool, regex string, attributeName *string) *RegexMatcher {
	return &RegexMatcher{
		Matcher: Matcher{
			negate:        negate,
			attributeName: attributeName,
		},
		regex: regex,
	}
}
