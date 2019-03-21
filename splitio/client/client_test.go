package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/splitio/go-client/splitio/service/dtos"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-toolkit/asynctask"
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

	factory := SplitFactory{
		client: &client,
	}
	factory.status.Store(SdkReady)
	client.factory = &factory

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

	factory := SplitFactory{
		client: &client,
	}
	factory.status.Store(SdkReady)
	client.factory = &factory

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
	client.BlockUntilReady(1)

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

	factory := SplitFactory{
		client: &client,
	}
	factory.status.Store(SdkReady)
	client.factory = &factory

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

func TestBlockUntilReadyWrongTimerPassed(t *testing.T) {
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
	err = client.BlockUntilReady(-1)
	expected := "SDK Initialization: timer must be positive number"
	if err != nil && err.Error() != expected {
		t.Error("Error was expected")
	}

	manager := factory.Manager()
	err = manager.BlockUntilReady(-1)
	if err != nil && err.Error() != expected {
		t.Error("Error was expected")
	}
}

func TestBlockUntilReadyStatusLoclahost(t *testing.T) {
	file, err := ioutil.TempFile("", "splitio_tests")
	if err != nil {
		t.Error("Couldn't create temporary file for localhost client tests: ", err)
		return
	}

	file.Write([]byte("feature1 on\n"))
	file.Sync()

	sdkConf := conf.Default()
	sdkConf.SplitFile = file.Name()

	factory, _ := NewSplitFactory("localhost", sdkConf)

	client := factory.Client()
	manager := factory.Manager()

	if len(manager.SplitNames()) != 0 {
		t.Error("It should not return splits")
	}

	if client.factory.IsReady() {
		t.Error("Client should not be ready")
	}

	err = client.Track("something", "something", "something", nil)
	if err == nil {
		t.Error("It should return error")
	}

	if client.Treatment("something", "something", nil) != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	if client.Treatment("something", "something", nil) != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	features := []string{"something"}
	result := client.Treatments("something", features, nil)
	if result["something"] != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	err = client.BlockUntilReady(1)
	if err != nil {
		t.Error("Error was not expected")
	}

	if !client.factory.IsReady() {
		t.Error("Client should be ready")
	}

	if !manager.factory.IsReady() {
		t.Error("Manager should be ready")
	}

	err = client.Track("something", "something", "something", nil)
	if err != nil {
		t.Error("It should not return error")
	}

	feature1 := client.Treatment("asd", "feature1", nil)
	if feature1 != "on" {
		t.Error("Feature1 retrieved incorrectly")
	}

	if manager.SplitNames()[0] != "feature1" {
		t.Error("It should return splits")
	}
}

func TestBlockUntilReadyStatusLoclahostOnDestroy(t *testing.T) {
	file, err := ioutil.TempFile("", "splitio_tests")
	if err != nil {
		t.Error("Couldn't create temporary file for localhost client tests: ", err)
		return
	}

	file.Write([]byte("feature1 on\n"))
	file.Sync()

	sdkConf := conf.Default()
	sdkConf.SplitFile = file.Name()

	factory, _ := NewSplitFactory("localhost", sdkConf)

	client := factory.Client()
	manager := factory.Manager()

	if len(manager.SplitNames()) != 0 {
		t.Error("It should not return splits")
	}

	if client.factory.IsReady() {
		t.Error("Client should not be ready")
	}

	err = client.BlockUntilReady(1)
	if err != nil {
		t.Error("Error was not expected")
	}

	if !client.factory.IsReady() {
		t.Error("Client should be ready")
	}

	if !manager.factory.IsReady() {
		t.Error("Manager should be ready")
	}

	client.Destroy()

	if !client.factory.IsDestroyed() {
		t.Error("Client should be destroyed")
	}

	if !manager.factory.IsDestroyed() {
		t.Error("Manager should be destroyed")
	}

	err = manager.BlockUntilReady(1)
	expected := "SDK Initialization: Client is destroyed"
	if err == nil || err.Error() != expected {
		t.Error("It should return an error")
	}
}

func TestBlockUntilReadyRedis(t *testing.T) {
	sdkConf := conf.Default()
	sdkConf.OperationMode = "redis-consumer"

	factory, _ := NewSplitFactory("something", sdkConf)

	if !factory.IsReady() {
		t.Error("Factory should be ready immediately")
	}

	client := factory.Client()
	if !client.factory.IsReady() {
		t.Error("Client should be ready immediately")
	}
	err := client.BlockUntilReady(1)
	if err != nil {
		t.Error("Error was not expected")
	}

	manager := factory.Manager()
	if !manager.factory.IsReady() {
		t.Error("Manager should be ready immediately")
	}
	err = manager.BlockUntilReady(1)
	if err != nil {
		t.Error("Error was not expected")
	}
}

func TestBlockUntilReadyInMemoryError(t *testing.T) {
	sdkConf := conf.Default()

	factory, _ := NewSplitFactory("something", sdkConf)

	if factory.IsReady() {
		t.Error("Factory should not be ready")
	}

	client := factory.Client()
	if client.factory.IsReady() {
		t.Error("Client should not be ready")
	}

	if client.Treatment("something", "something", nil) != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	if client.Treatment("something", "something", nil) != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	features := []string{"something"}
	result := client.Treatments("something", features, nil)
	if result["something"] != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	err := client.Track("something", "something", "something", nil)
	if err == nil {
		t.Error("It should return error")
	}

	expected := "Client Instantiation: Client is not ready yet"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	err = client.BlockUntilReady(1)
	if err == nil {
		t.Error("It should return error")
	}

	if err != nil && err.Error() != "SDK Initialization failed" {
		t.Error("Wrong error")
	}
}

func TestBlockUntilReadyInMemory(t *testing.T) {
	mockedSplit1 := dtos.SplitDTO{
		Algo:                  2,
		ChangeNumber:          123,
		DefaultTreatment:      "default",
		Killed:                false,
		Name:                  "split",
		Seed:                  1234,
		Status:                "ACTIVE",
		TrafficAllocation:     1,
		TrafficAllocationSeed: -1667452163,
		TrafficTypeName:       "tt1",
		Conditions: []dtos.ConditionDTO{
			{
				ConditionType: "ROLLOUT",
				Label:         "in segment all",
				MatcherGroup: dtos.MatcherGroupDTO{
					Combiner: "AND",
					Matchers: []dtos.MatcherDTO{
						{
							MatcherType:        "ALL_KEYS",
							Whitelist:          nil,
							Negate:             false,
							UserDefinedSegment: nil,
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
	mockedSplit2 := dtos.SplitDTO{Name: "split2", Killed: true, Status: "ACTIVE"}
	mockedSplit3 := dtos.SplitDTO{Name: "split3", Killed: true, Status: "INACTIVE"}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(4 * time.Second)
		if r.URL.Path != "/splits" && r.Method != "GET" {
			t.Error("Invalid request. Should be GET to /splits")
		}

		splitChanges := dtos.SplitChangesDTO{
			Splits: []dtos.SplitDTO{mockedSplit1, mockedSplit2, mockedSplit3},
			Since:  3,
			Till:   3,
		}

		raw, err := json.Marshal(splitChanges)
		if err != nil {
			t.Error("Error building json")
			return
		}

		w.Write(raw)
	}))
	defer ts.Close()

	segmentMock, _ := ioutil.ReadFile("../../testdata/segment_mock.json")

	tss := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		fmt.Fprintln(w, fmt.Sprintf(string(segmentMock)))
	}))
	defer tss.Close()

	sdkConf := conf.Default()
	sdkConf.Advanced.EventsURL = tss.URL
	sdkConf.Advanced.SdkURL = ts.URL

	factory, _ := NewSplitFactory("something", sdkConf)

	if factory.IsReady() {
		t.Error("Factory should not be ready")
	}

	client := factory.Client()
	if client.factory.IsReady() {
		t.Error("Client should not be ready")
	}

	manager := factory.Manager()
	if manager.factory.IsReady() {
		t.Error("Manager should not be ready")
	}

	if len(manager.SplitNames()) != 0 {
		t.Error("It should not return splits")
	}

	if client.Treatment("something", "something", nil) != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	if client.Treatment("something", "something", nil) != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	features := []string{"something"}
	result := client.Treatments("something", features, nil)
	if result["something"] != evaluator.Control {
		t.Error("Wrong evaluation")
	}

	err := client.Track("something", "something", "something", nil)
	if err == nil {
		t.Error("It should return error")
	}

	expected := "Client Instantiation: Client is not ready yet"
	if err != nil && err.Error() != expected {
		t.Error("Wrong error")
	}

	err = client.BlockUntilReady(1)
	if err == nil {
		t.Error("It should return error")
	}

	expected2 := "SDK Initialization: time of 1 exceeded"
	if err != nil && err.Error() != expected2 {
		t.Error("Wrong message error")
	}

	if client.factory.IsReady() {
		t.Error("Client should not be ready")
	}

	err = manager.BlockUntilReady(2)
	if err == nil {
		t.Error("It should return error")
	}

	expected2 = "SDK Initialization: time of 2 exceeded"
	if err != nil && err.Error() != expected2 {
		t.Error("Wrong message error")
	}

	err = client.BlockUntilReady(2)
	if err != nil && err.Error() != expected2 {
		t.Error("Wrong message error")
	}

	if !client.factory.IsReady() || !manager.factory.IsReady() {
		t.Error("Both client and manager should be ready")
	}

	if len(manager.SplitNames()) != 2 {
		t.Error("It should return Splits")
	}

	if client.Treatment("aaaaaaklmnbv", "split", nil) != "on" {
		t.Error("Treatment error")
	}
}
