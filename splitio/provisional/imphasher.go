package provisional

import (
	"fmt"
	"strings"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/provisional/hashing"
)

const hashKeyTemplate = "%s:%s:%s:%s:%d"

func unknownIfEmpty(s string) string {
	if len(strings.TrimSpace(s)) == 0 {
		return "UNKNOWN"
	}
	return s
}

// ImpressionHasher interface
type ImpressionHasher interface {
	Process(featureName string, keyImpression *dtos.ImpressionRecord) (int64, error)
}

// ImpressionHasherImpl implements the hasher interface, mapping certain fields to an int64
type ImpressionHasherImpl struct{}

// Process an impression and return the 64 LSBs of a murmur3-128 digest
func (h *ImpressionHasherImpl) Process(featureName string, keyImpression *dtos.ImpressionRecord) (int64, error) {
	if keyImpression == nil {
		return 0, fmt.Errorf("keyImpression cannot be nil")
	}

	toHash := fmt.Sprintf(hashKeyTemplate,
		unknownIfEmpty(keyImpression.KeyName),
		unknownIfEmpty(featureName),
		unknownIfEmpty(keyImpression.Treatment),
		unknownIfEmpty(keyImpression.Label),
		keyImpression.ChangeNumber)

	h1, _ := hashing.Sum128([]byte(toHash))
	return int64(h1), nil
}
