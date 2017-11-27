package client

import (
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-client/splitio/util/configuration"
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
	return &evaluator.Result{
		EvaluationTimeNs:  0,
		Label:             "aLabel",
		SplitChangeNumber: 123,
		Treatment:         "TreatmentA",
	}
}

func TestClientGetTreatment(t *testing.T) {
	cfg := configuration.Default()
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
