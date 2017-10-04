package engine

import (
	"github.com/splitio/go-client/splitio/engine/evaluator/impressionlabels"
	"github.com/splitio/go-client/splitio/engine/grammar"
	"github.com/splitio/go-client/splitio/util/logging"
)

// Engine lala
type Engine struct {
	Logger logging.LoggerInterface
}

// DoEvaluation lala
func (e *Engine) DoEvaluation(
	split *grammar.Split,
	key string,
	bucketingKey string,
	attributes map[string]interface{},
) (*string, string) {

	if bucketingKey == "" {
		bucketingKey = key
	}

	inRollOut := false
	for _, condition := range split.Conditions() {
		if !inRollOut && condition.ConditionType() == grammar.ConditionTypeRollout {
			if split.TrafficAllocation() < 100 {
				bucket := e.calculateBucket(split.Algo(), bucketingKey, split.TrafficAllocationSeed())
				if bucket >= split.TrafficAllocation() {
					defaultTreatment := split.DefaultTreatment()
					return &defaultTreatment, impressionlabels.NotInSplit
				}
				inRollOut = true
			}
		}
		if condition.Matches(key, attributes) {
			bucket := e.calculateBucket(split.Algo(), bucketingKey, split.TrafficAllocationSeed())
			treatment := condition.CalculateTreatment(bucket)
			return treatment, condition.Label()
		}
	}
	return nil, impressionlabels.NoConditionMatched
}

func (e *Engine) calculateBucket(algo int, bucketingKey string, seed int64) int {
	// TODO
	return 0
}
