package matchers

import (
	"github.com/splitio/go-toolkit/datastructures/set"
)

// ContainsAllOfSetMatcher matches if the set supplied to the getTreatment is a superset of the one in the split
type ContainsAllOfSetMatcher struct {
	Matcher
	comparisonSet *set.ThreadUnsafeSet
}

// Match returns true if the set provided is a superset of the one in the split
func (m *ContainsAllOfSetMatcher) Match(key string, attributes map[string]interface{}, bucketingKey *string) bool {
	matchingKey, err := m.matchingKey(key, attributes)
	if err != nil {
		return false
	}

	conv, ok := matchingKey.([]string)
	if !ok {
		return false
	}

	matchingSet := set.NewSet()
	for _, x := range conv {
		matchingSet.Add(x)
	}

	return matchingSet.IsSuperset(m.comparisonSet)

}

// NewContainsAllOfSetMatcher returns a pointer to a new instance of ContainsAllOfSetMatcher
func NewContainsAllOfSetMatcher(negate bool, setItems []string, attributeName *string) *ContainsAllOfSetMatcher {
	setObj := set.NewSet()
	for _, item := range setItems {
		setObj.Add(item)
	}

	return &ContainsAllOfSetMatcher{
		Matcher: Matcher{
			negate:        negate,
			attributeName: attributeName,
		},
		comparisonSet: setObj,
	}
}