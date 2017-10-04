package evaluator

import (
	"time"

	"github.com/splitio/go-client/splitio/engine"
	"github.com/splitio/go-client/splitio/engine/grammar"
	"github.com/splitio/go-client/splitio/storage"
)

// Result represents the result of an evaluation, including the resulting treatment, the label for the impression,
// the latency and error if any
type Result struct {
	Treatment         string
	Label             string
	Latency           int
	Error             error
	SplitChangeNumber int64
}

// Evaluator struct is the main evaluator
type Evaluator struct {
	splitStorage   storage.SplitStorage
	segmentStorage storage.SegmentStorage
	eng            engine.Engine
}

// NewEvaluator lala
func NewEvaluator(splitStorage storage.SplitStorage, segmentStorage storage.SegmentStorage, eng engine.Engine) {
	return &Evaluator{
		splitStorage:   splitStorage,
		segmentStorage: segmentStorage,
		eng:            eng,
	}
}

// Evaluate returns the appropriate treatment together with an error indicating if something went wrong
func (e *Evaluator) Evaluate(key string, bucketingKey string, feature string, attributes map[string]interface{}) *Result {
	dto, found := e.splitStorage.Get(feature)
	if !found {
		return &Result{Treatment: "CONTROL", Label: SplitNotFound}
	}

	split := grammar.NewSplit(dto)

	if split.Killed {
		return &Result{Treatment: split.DefaultTreatment, Label: Killed, SplitChangeNumber: split.ChangeNumber}
	}

	before := time.Now()
	treatment, label, err := Engine.DoEvaluation(split, key, bucketingKey, attributes)
	after := time.Now()

	if treatment == nil {
		treatment = split.DefaultTreatment
		label = NoConditionMatched
	}

	return &Result{
		Treatment:         treatment,
		Label:             label,
		Latency:           after.Sub(before).Nanoseconds() / 1000,
		SplitChangeNumber: split.ChangeNumber,
	}
}
