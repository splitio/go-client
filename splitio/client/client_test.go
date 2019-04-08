package client

import (
	"github.com/splitio/go-client/splitio/service/dtos"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"

	"io/ioutil"
	"os"
	"testing"
	"time"
)

type mockEvaluator struct{}
type mockEvents struct{}
type mockEventsPanic struct{}

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
			Treatment:         evaluator.Control,
		}
	}
}

func (e *mockEventsPanic) Evaluate(
	key string,
	bucketingKey *string,
	feature string,
	attributes map[string]interface{},
) *evaluator.Result {
	panic("Testing panicking")
}

func (s *mockEvents) Push(event dtos.EventDTO) error { return nil }

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

	impressions := client.impressions.(storage.ImpressionStorage)
	impression := impressions.PopAll()[0].KeyImpressions[0]
	if impression.Label != "aLabel" {
		t.Error("Impression should have label when labelsEnabled is true")
	}

	client.cfg.LabelsEnabled = false
	client.Treatment("key", "feature2", nil)

	impression = impressions.PopAll()[0].KeyImpressions[0]
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
	if !ok || notFeatureRes != evaluator.Control {
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

func TestClientGetTreatmentConsideringValidationInputs(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	client := SplitClient{
		cfg:         cfg,
		evaluator:   &mockEvaluator{},
		impressions: mutexmap.NewMMImpressionStorage(),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
		validator:   inputValidation{logger: logger},
	}

	feature1 := client.Treatment(nil, "feature", nil)
	if feature1 != "control" {
		t.Error("Feature1 retrieved incorrectly")
	}

	feature2 := client.Treatment(true, "feature", nil)
	if feature2 != "control" {
		t.Error("Feature2 retrieved incorrectly")
	}

	feature3 := client.Treatment(123, "feature", nil)
	if feature3 != "TreatmentA" {
		t.Error("Feature3 retrieved incorrectly")
	}

	feature4 := client.Treatment("key", "feature", nil)
	if feature4 != "TreatmentA" {
		t.Error("Feature4 retrieved incorrectly")
	}

	var key = &Key{
		MatchingKey:  "key",
		BucketingKey: "bucketing",
	}

	feature5 := client.Treatment(key, "feature", nil)
	if feature5 != "TreatmentA" {
		t.Error("Feature5 retrieved incorrectly")
	}
}

func TestClientPanicking(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	client := SplitClient{
		cfg:         cfg,
		evaluator:   &mockEventsPanic{},
		events:      &mockEvents{},
		impressions: mutexmap.NewMMImpressionStorage(),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
	}

	treatment := client.Treatment("key", "some", nil)
	if treatment != "control" {
		t.Error("treatment retrieved incorrectly")
	}
}

func TestClientDestroy(t *testing.T) {
	logger := logging.NewLogger(nil)

	resSplits := 0
	stoppedSplit := false
	resSegments := 0
	stoppedSegments := false
	resImpressions := 0
	stoppedImpressions := false
	resGauge := 0
	stoppedGauge := false
	resCounters := 0
	stoppedCounters := false
	resLatencies := 0
	stoppedLatencies := false

	splitSync := func(l logging.LoggerInterface) error { resSplits++; return nil }
	splitStop := func(l logging.LoggerInterface) { stoppedSplit = true }
	segmentSync := func(l logging.LoggerInterface) error { resSegments++; return nil }
	segmentStop := func(l logging.LoggerInterface) { stoppedSegments = true }
	impressionSync := func(l logging.LoggerInterface) error { resImpressions++; return nil }
	impressionStop := func(l logging.LoggerInterface) { stoppedImpressions = true }
	gaugeSync := func(l logging.LoggerInterface) error { resGauge++; return nil }
	gaugeStop := func(l logging.LoggerInterface) { stoppedGauge = true }
	counterSync := func(l logging.LoggerInterface) error { resCounters++; return nil }
	counterStop := func(l logging.LoggerInterface) { stoppedCounters = true }
	latencySync := func(l logging.LoggerInterface) error { resLatencies++; return nil }
	latencyStop := func(l logging.LoggerInterface) { stoppedLatencies = true }

	splitTask := asynctask.NewAsyncTask("splits", splitSync, 100, nil, splitStop, logger)
	segmentsTask := asynctask.NewAsyncTask("segments", segmentSync, 100, nil, segmentStop, logger)
	impressionsTask := asynctask.NewAsyncTask("impressions", impressionSync, 100, nil, impressionStop, logger)
	gaugesTask := asynctask.NewAsyncTask("gauges", gaugeSync, 100, nil, gaugeStop, logger)
	countersTask := asynctask.NewAsyncTask("counters", counterSync, 100, nil, counterStop, logger)
	latenciesTask := asynctask.NewAsyncTask("latencies", latencySync, 100, nil, latencyStop, logger)

	splitTask.Start()
	segmentsTask.Start()
	impressionsTask.Start()
	gaugesTask.Start()
	countersTask.Start()
	latenciesTask.Start()

	client := SplitClient{
		cfg:    &conf.SplitSdkConfig{},
		logger: logger,
		sync: &sdkSync{
			countersSync:   countersTask,
			gaugeSync:      gaugesTask,
			impressionSync: impressionsTask,
			latenciesSync:  latenciesTask,
			segmentSync:    segmentsTask,
			splitSync:      splitTask,
		},
	}

	factory := SplitFactory{
		client: &client,
	}

	client.factory = &factory

	time.Sleep(1 * time.Second)
	client.Destroy()
	time.Sleep(1 * time.Second)

	if splitTask.IsRunning() {
		t.Error("split task should be stopped")
	}

	if segmentsTask.IsRunning() {
		t.Error("segment task should be stopped")
	}

	if impressionsTask.IsRunning() {
		t.Error("impression task should be stopped")
	}

	if gaugesTask.IsRunning() {
		t.Error("gauges task should be stopped")
	}

	if countersTask.IsRunning() {
		t.Error("counters task should be stopped")
	}

	if latenciesTask.IsRunning() {
		t.Error("latencies task should be stopped")
	}

	// -----

	if resSplits != 1 {
		t.Error("Splits should have run once")
	}

	if resSegments != 1 {
		t.Error("Segments should have run once")
	}

	if resImpressions != 1 {
		t.Error("Impressions should have run once")
	}

	if resGauge != 1 {
		t.Error("Gauge should have run once")
	}

	if resCounters != 1 {
		t.Error("Conters should have run once")
	}

	if resLatencies != 1 {
		t.Error("Latencies should have run once")
	}

	if !client.IsDestroyed() {
		t.Error("Client should be destroyed")
	}

	if client.Treatment("key", "feature", nil) != evaluator.Control {
		t.Error("Single .Treatment() call should return control")
	}

	if !stoppedCounters {
		t.Error("Counters shoud be stopped")
	}

	if !stoppedGauge {
		t.Error("Gauge shoud be stopped")
	}

	if !stoppedImpressions {
		t.Error("Impressions shoud be stopped")
	}

	if !stoppedLatencies {
		t.Error("Latencies shoud be stopped")
	}

	if !stoppedSegments {
		t.Error("Segments shoud be stopped")
	}

	if !stoppedSplit {
		t.Error("Split shoud be stopped")
	}

	treatments := client.Treatments("key", []string{"feature1", "feature2", "feature3"}, nil)
	if len(treatments) != 3 {
		t.Error("Should return 3 treatments.")
	}

	if treatments["feature1"] != evaluator.Control {
		t.Error("Wrong treatment result")
	}

	if treatments["feature2"] != evaluator.Control {
		t.Error("Wrong treatment result")
	}

	if treatments["feature3"] != evaluator.Control {
		t.Error("Wrong treatment result")
	}
}

var valid = &dtos.SplitDTO{
	Algo:                  2,
	ChangeNumber:          1494593336752,
	DefaultTreatment:      "off",
	Killed:                false,
	Name:                  "valid",
	Seed:                  -1992295819,
	Status:                "ACTIVE",
	TrafficAllocation:     100,
	TrafficAllocationSeed: -285565213,
	TrafficTypeName:       "user",
	Configurations:        map[string]string{"on": "{\"color\": \"blue\",\"size\": 13}"},
	Conditions: []dtos.ConditionDTO{
		{
			ConditionType: "ROLLOUT",
			Label:         "default rule",
			MatcherGroup: dtos.MatcherGroupDTO{
				Combiner: "AND",
				Matchers: []dtos.MatcherDTO{
					{
						KeySelector: &dtos.KeySelectorDTO{
							TrafficType: "user",
							Attribute:   nil,
						},
						MatcherType: "IN_SEGMENT",
						Whitelist:   nil,
						Negate:      false,
						UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
							SegmentName: "employees",
						},
					},
				},
			},
			Partitions: []dtos.PartitionDTO{
				{
					Size:      100,
					Treatment: "on",
				},
			},
		},
	},
}
var killed = &dtos.SplitDTO{
	Algo:                  2,
	ChangeNumber:          1494593336752,
	DefaultTreatment:      "defTreatment",
	Killed:                true,
	Name:                  "killed",
	Seed:                  -1992295819,
	Status:                "ACTIVE",
	TrafficAllocation:     100,
	TrafficAllocationSeed: -285565213,
	TrafficTypeName:       "user",
	Configurations:        map[string]string{"defTreatment": "{\"color\": \"orange\",\"size\": 15}"},
	Conditions: []dtos.ConditionDTO{
		{
			ConditionType: "ROLLOUT",
			Label:         "default rule",
			MatcherGroup: dtos.MatcherGroupDTO{
				Combiner: "AND",
				Matchers: []dtos.MatcherDTO{
					{
						KeySelector: &dtos.KeySelectorDTO{
							TrafficType: "user",
							Attribute:   nil,
						},
						MatcherType: "IN_SEGMENT",
						Whitelist:   nil,
						Negate:      false,
						UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
							SegmentName: "employees",
						},
					},
				},
			},
			Partitions: []dtos.PartitionDTO{
				{
					Size:      100,
					Treatment: "off",
				},
			},
		},
	},
}

type mockStorage struct{}

func (s *mockStorage) Get(
	feature string,
) *dtos.SplitDTO {
	switch feature {
	default:
	case "valid":
		return valid
	case "killed":
		return killed
	}
	return nil
}
func (s *mockStorage) GetAll() []dtos.SplitDTO            { return make([]dtos.SplitDTO, 0) }
func (s *mockStorage) SegmentNames() *set.ThreadUnsafeSet { return nil }
func (s *mockStorage) SplitNames() []string               { return make([]string, 0) }

type mockSegmentStorage struct{}

func (i *mockSegmentStorage) Get(feature string) *set.ThreadUnsafeSet {
	switch feature {
	default:
	case "employees":
		return set.NewSet("user1")
	}
	return nil
}

func isInvalidImpression(client SplitClient, key string, feature string, treatment string) bool {
	impressions := client.impressions.(storage.ImpressionStorage)
	impression := impressions.PopAll()[0]
	name := impression.TestName
	i := impression.KeyImpressions[0]

	if name != feature || i.KeyName != key || treatment != i.Treatment {
		return true
	}
	return false
}
func TestClient(t *testing.T) {
	cfg := conf.Default()
	cfg.LabelsEnabled = true
	logger := logging.NewLogger(nil)

	evaluator := evaluator.NewEvaluator(
		&mockStorage{},
		&mockSegmentStorage{},
		nil,
		logger,
	)

	client := SplitClient{
		cfg:         cfg,
		evaluator:   evaluator,
		impressions: mutexmap.NewMMImpressionStorage(),
		logger:      logger,
		metrics:     mutexmap.NewMMMetricsStorage(),
	}

	// Assertions Treatment
	if client.Treatment("user1", "valid", nil) != "on" {
		t.Error("Unexpected Treatment Result")
	}
	if isInvalidImpression(client, "user1", "valid", "on") {
		t.Error("Wrong impression saved")
	}

	if client.Treatment("invalid", "valid", nil) != "off" {
		t.Error("Unexpected Treatment Result")
	}
	if isInvalidImpression(client, "invalid", "valid", "off") {
		t.Error("Wrong impression saved")
	}

	if client.Treatment("invalid", "invalid", nil) != "control" {
		t.Error("Unexpected Treatment Result")
	}
	if isInvalidImpression(client, "invalid", "invalid", "control") {
		t.Error("Wrong impression saved")
	}

	if client.Treatment("invalid", "killed", nil) != "defTreatment" {
		t.Error("Unexpected Treatment Result")
	}
	if isInvalidImpression(client, "invalid", "killed", "defTreatment") {
		t.Error("Wrong impression saved")
	}

	// Assertion Treatments
	treatments := client.Treatments("user1", []string{"valid", "invalid", "killed"}, nil)
	if treatments["invalid"] != "control" {
		t.Error("Unexpected treatment result")
	}
	if treatments["killed"] != "defTreatment" {
		t.Error("Unexpected treatment result")
	}
	if treatments["valid"] != "on" {
		t.Error("Unexpected treatment result")
	}

	// Assertion TreatmentWithConfig
	result := client.TreatmentWithConfig("user1", "valid", nil)
	if result.Treatment != "on" {
		t.Error("Unexpected Treatment Result")
	}
	if result.Config != "{\"color\": \"blue\",\"size\": 13}" {
		t.Error("Unexpected Config Result")
	}
	if isInvalidImpression(client, "user1", "valid", "on") {
		t.Error("Wrong impression saved")
	}

	result = client.TreatmentWithConfig("invalid", "valid", nil)
	if result.Treatment != "off" {
		t.Error("Unexpected Treatment Result")
	}
	if result.Config != nil {
		t.Error("Unexpected Config Result")
	}
	if isInvalidImpression(client, "invalid", "valid", result.Treatment) {
		t.Error("Wrong impression saved")
	}

	result = client.TreatmentWithConfig("invalid", "invalid", nil)
	if result.Treatment != "control" {
		t.Error("Unexpected Treatment Result")
	}
	if result.Config != nil {
		t.Error("Unexpected Config Result")
	}
	if isInvalidImpression(client, "invalid", "invalid", result.Treatment) {
		t.Error("Wrong impression saved")
	}

	result = client.TreatmentWithConfig("invalid", "killed", nil)
	if result.Treatment != "defTreatment" {
		t.Error("Unexpected Treatment Result")
	}
	if result.Config != "{\"color\": \"orange\",\"size\": 15}" {
		t.Error("Unexpected Config Result")
	}
	if isInvalidImpression(client, "invalid", "killed", result.Treatment) {
		t.Error("Wrong impression saved")
	}
}
