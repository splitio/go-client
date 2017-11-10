package evaluator

import (
	"time"

	"github.com/splitio/go-client/splitio/engine"
	"github.com/splitio/go-client/splitio/engine/evaluator/impressionlabels"
	"github.com/splitio/go-client/splitio/engine/grammar"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-toolkit/injection"
	"github.com/splitio/go-toolkit/logging"
)

// Result represents the result of an evaluation, including the resulting treatment, the label for the impression,
// the latency and error if any
type Result struct {
	Treatment         string
	Label             string
	EvaluationTimeNs  int64
	SplitChangeNumber int64
}

// Evaluator struct is the main evaluator
type Evaluator struct {
	splitStorage   storage.SplitStorage
	segmentStorage storage.SegmentStorage
	eng            *engine.Engine
	logger         logging.LoggerInterface
}

// NewEvaluator instantiates an Evaluator struct and returns a reference to it
func NewEvaluator(
	splitStorage storage.SplitStorage,
	segmentStorage storage.SegmentStorage,
	eng *engine.Engine,
	logger logging.LoggerInterface,
) *Evaluator {
	return &Evaluator{
		splitStorage:   splitStorage,
		segmentStorage: segmentStorage,
		eng:            eng,
		logger:         logger,
	}
}

// Evaluate returns a struct with the resulting treatment and extra information for the impression
func (e *Evaluator) Evaluate(key string, bucketingKey *string, feature string, attributes map[string]interface{}) *Result {
	splitDto := e.splitStorage.Get(feature)
	if splitDto == nil {
		return &Result{Treatment: "control", Label: impressionlabels.SplitNotFound}
	}

	// TODO: Move this to NewEvaluator ?
	ctx := injection.NewContext()
	ctx.AddDependency("segmentStorage", e.segmentStorage)
	ctx.AddDependency("evaluator", e)

	split := grammar.NewSplit(splitDto, ctx, e.logger)

	if split.Killed() {
		return &Result{
			Treatment:         split.DefaultTreatment(),
			Label:             impressionlabels.Killed,
			SplitChangeNumber: split.ChangeNumber(),
		}
	}

	before := time.Now()
	treatment, label := e.eng.DoEvaluation(split, key, bucketingKey, attributes)
	after := time.Now()

	if treatment == nil {
		defaultTreatment := split.DefaultTreatment()
		treatment = &defaultTreatment
		label = impressionlabels.NoConditionMatched
	}

	return &Result{
		Treatment:         *treatment,
		Label:             label,
		EvaluationTimeNs:  after.Sub(before).Nanoseconds(),
		SplitChangeNumber: split.ChangeNumber(),
	}
}

// EvaluateDependency SHOULD ONLY BE USED by DependencyMatcher.
// It's used to break the dependency cycle between matchers and evaluators.
func (e *Evaluator) EvaluateDependency(key string, bucketingKey *string, feature string, attributes map[string]interface{}) string {
	res := e.Evaluate(key, bucketingKey, feature, attributes)
	return res.Treatment
}
