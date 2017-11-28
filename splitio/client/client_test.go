package client

import (
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-toolkit/logging"
	"testing"
)

type mockEvaluator struct{}

func (e *mockEvaluator) Evaluate(
	key string,
	bucketingKey *string,
	feature string,
	attributes map[string]interface{},
) *evaluator.Result {
	switch feature {
	case "feature":
		return &evaluator.Result{
			EvaluationTimeNs:  0,
			Label:             "aLabel",
			SplitChangeNumber: 123,
			Treatment:         "TreatmentA",
		}
	default:
		return &evaluator.Result{
			EvaluationTimeNs:  0,
			Label:             "exception",
			SplitChangeNumber: 123,
			Treatment:         "control",
		}
	}
}

func TestClientGetTreatment(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	client := SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexmap.NewMMImpressionStorage(),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
	}

	client.Treatment("key", "feature", nil)

	impression := client.impressions.PopAll()[0].KeyImpressions[0]
	if impression.Label != "aLabel" {
		t.Error("Impression should have label when labelsEnabled is true")
	}

	client.cfg.LabelsEnabled = false
	client.Treatment("key", "feature2", nil)

	impression = client.impressions.PopAll()[0].KeyImpressions[0]
	if impression.Label != "" {
		t.Error("Impression should have label when labelsEnabled is true")
	}

}

func TestTreatments(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	client := SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexmap.NewMMImpressionStorage(),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
	}

	res := client.Treatments("user1", []string{"feature", "notFeature"}, nil)

	featureRes, ok := res["feature"]
	if !ok || featureRes != "TreatmentA" {
		t.Error("Incorrect result for \"feature\"")
	}

	notFeatureRes, ok := res["notFeature"]
	if !ok || notFeatureRes != "control" {
		t.Error("Incorrect result for \"notFeature\"")
	}

}
