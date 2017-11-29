package client

import (
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-toolkit/logging"
	"io/ioutil"
	"os"
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

func TestLocalhostMode(t *testing.T) {
	file, err := ioutil.TempFile("", "splitio_tests")
	if err != nil {
		t.Error("Couldn't create temporary file for localhost client tests: ", err)
		return
	}

	file.Write([]byte("feature1 on\n"))
	file.Write([]byte("feature2 off\n"))
	file.Sync()

	sdkConf := conf.Default()
	sdkConf.SplitFile = file.Name()
	factory, _ := NewSplitFactory("localhost", sdkConf)
	client := factory.Client()

	if client.cfg.OperationMode != "localhost" {
		t.Error("Localhost operation mode should be set when received apikey is 'localhost'")
	}

	feature1 := client.Treatment("asd", "feature1", nil)
	if feature1 != "on" {
		t.Error("Feature1 retrieved incorrectly")
	}

	feature2 := client.Treatment("asd", "feature2", nil)
	if feature2 != "off" {
		t.Error("Feature2 retrieved incorrectly")
	}

	file.Close()
	os.Remove(file.Name())
}
