package evaluator

import (
	"time"

	"github.com/splitio/go-client/splitio/engine"
	"github.com/splitio/go-client/splitio/engine/evaluator/impressionlabels"
	"github.com/splitio/go-client/splitio/engine/grammar"
	"github.com/splitio/go-client/splitio/storage"
)

// Result represents the result of an evaluation, including the resulting treatment, the label for the impression,
// the latency and error if any
type Result struct {
	Treatment         string
	Label             string
	Latency           int64
	SplitChangeNumber int64
}

// Evaluator struct is the main evaluator
type Evaluator struct {
	splitStorage   storage.SplitStorage
	segmentStorage storage.SegmentStorage
	eng            engine.Engine
}

// NewEvaluator lala
func NewEvaluator(
	splitStorage storage.SplitStorage,
	segmentStorage storage.SegmentStorage,
	eng engine.Engine,
) *Evaluator {
	return &Evaluator{
		splitStorage:   splitStorage,
		segmentStorage: segmentStorage,
		eng:            eng,
	}
}

// Evaluate returns a struct with the resulting treatment and extra information for the impression
func (e *Evaluator) Evaluate(key string, bucketingKey string, feature string, attributes map[string]interface{}) *Result {
	splitDto := e.splitStorage.Get(feature)
	if splitDto == nil {
		return &Result{Treatment: "CONTROL", Label: impressionlabels.SplitNotFound}
	}

	split := grammar.NewSplit(splitDto)

	if split.Killed() {
		return &Result{
			Treatment:         split.DefaultTreatment(),
			Label:             impressionlabels.Killed,
			SplitChangeNumber: split.ChangeNumber(),
		}
	}

	// TODO: SEBA METRICAS
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
		Latency:           after.Sub(before).Nanoseconds() / 1000,
		SplitChangeNumber: split.ChangeNumber(),
	}
}
