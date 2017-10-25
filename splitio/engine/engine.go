package engine

import (
	"math"

	"github.com/splitio/go-client/splitio/engine/evaluator/impressionlabels"
	"github.com/splitio/go-client/splitio/engine/grammar"
	"github.com/splitio/go-client/splitio/engine/hash"
	"github.com/splitio/go-toolkit/logging"
)

// Engine struct is responsible for cheking if any of the conditions of the split matches,
// performing traffic allocation, calculating the bucket and returning the appropriate treatment
type Engine struct {
	Logger logging.LoggerInterface
}

// DoEvaluation performs the main evaluation agains each condition
func (e *Engine) DoEvaluation(
	split *grammar.Split,
	key string,
	bucketingKey *string,
	attributes map[string]interface{},
) (*string, string) {

	if bucketingKey == nil {
		bucketingKey = &key
	}

	inRollOut := false
	for _, condition := range split.Conditions() {
		if !inRollOut && condition.ConditionType() == grammar.ConditionTypeRollout {
			if split.TrafficAllocation() < 100 {
				bucket := e.calculateBucket(split.Algo(), *bucketingKey, split.TrafficAllocationSeed())
				if bucket >= split.TrafficAllocation() {
					defaultTreatment := split.DefaultTreatment()
					return &defaultTreatment, impressionlabels.NotInSplit
				}
				inRollOut = true
			}
		}
		if condition.Matches(key, attributes) {
			bucket := e.calculateBucket(split.Algo(), *bucketingKey, split.TrafficAllocationSeed())
			treatment := condition.CalculateTreatment(bucket)
			return treatment, condition.Label()
		}
	}
	return nil, impressionlabels.NoConditionMatched
}

func (e *Engine) calculateBucket(algo int, bucketingKey string, seed int64) int {
	var hashedKey uint32
	switch algo {
	case grammar.SplitAlgoMurmur:
		hashedKey = hash.Murmur3_32([]byte(bucketingKey), uint32(seed))
	case grammar.SplitAlgoLegacy:
		fallthrough
	default:
		hashedKey = hash.Legacy([]byte(bucketingKey), uint32(seed))
	}

	return int(math.Abs(float64(hashedKey%100)) + 1)

}
