package matchers

import (
	"github.com/splitio/go-client/splitio/storage"
)

// InSegmentMatcher matches if the key passed is in the segment which the matcher was constructed with
type InSegmentMatcher struct {
	Matcher
	segmentName string
}

// Match returns true if the key is in the matcher's segment
func (m *InSegmentMatcher) Match(key string, attributes map[string]interface{}, bucketingKey *string) bool {
	segmentStorage, ok := m.Context.Dependency("segmentStorage").(storage.SegmentStorage)
	if !ok {
		return false
	}

	segment := segmentStorage.Get(m.segmentName)
	if segment == nil {
		return false
	}

	return segment.Has(key)
}

// NewInSegmentMatcher instantiates a new InSegmentMatcher
func NewInSegmentMatcher(negate bool, segmentName string, attributeName *string) *InSegmentMatcher {
	return &InSegmentMatcher{
		Matcher: Matcher{
			negate:        negate,
			attributeName: attributeName,
		},
		segmentName: segmentName,
	}
}
