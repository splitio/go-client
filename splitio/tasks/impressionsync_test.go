package tasks

import (
	"encoding/json"
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/service/api"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-toolkit/logging"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestImpressionSyncTask(t *testing.T) {
	reqestReceived := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/impressions" && r.Method != "POST" {
			t.Error("Invalid request. Should be POST to /impressions")
		}
		reqestReceived = true

		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Error("Error reading body")
			return
		}

		var impressions []dtos.ImpressionsDTO
		err = json.Unmarshal(body, &impressions)
		if err != nil {
			t.Errorf("Error parsing json: %s", err)
			return
		}

		if len(impressions) != 1 {
			t.Error("Incorrect number of features")
			return
		}

		if impressions[0].TestName != "feature_1" && len(impressions[0].KeyImpressions) != 2 {
			t.Error("Incorrect impressions received")
		}

	}))
	defer ts.Close()

	logger := logging.NewLogger(&logging.LoggerOptions{})
	impressionRecorder := api.NewHTTPImpressionRecorder(
		"",
		&conf.SplitSdkConfig{
			Advanced: conf.AdvancedConfig{
				EventsURL: ts.URL,
				SdkURL:    ts.URL,
			},
		},
		logger,
	)

	impressionStorage := mutexmap.NewMMImpressionStorage()

	impressionTask := NewRecordImpressionsTask(
		impressionStorage,
		impressionRecorder,
		1,
		"go-0.1",
		"192.168.0.123",
		"machine1",
		nil,
		logger,
	)

	impressionTask.Start()

	if !impressionTask.IsRunning() {
		t.Error("Impression recording task should be running")
	}

	impressionStorage.Put("feature1", &dtos.ImpressionDTO{
		BucketingKey: "123",
		ChangeNumber: 456,
		KeyName:      "key1",
		Time:         123,
		Treatment:    "on",
	})

	impressionStorage.Put("feature1", &dtos.ImpressionDTO{
		BucketingKey: "123",
		ChangeNumber: 456,
		KeyName:      "key2",
		Time:         124,
		Treatment:    "off",
	})

	time.Sleep(time.Second * 10)

	if !reqestReceived {
		t.Error("Request not received")
	}

	impressionTask.Stop()

	time.Sleep(time.Second * 2)

	if impressionTask.IsRunning() {
		t.Error("Task should be stopped")
	}
}

type goodListener struct {
	status bool
}

func (l *goodListener) Notify(impressions []dtos.ImpressionsDTO) {
	imp := impressions[0].KeyImpressions[0]
	if impressions[0].TestName == "feature1" && imp.Treatment == "aTreatment" {
		l.status = true
	}
}

type badListener struct{}

func (l *badListener) Notify(impressions []dtos.ImpressionsDTO) {
	panic("some msg")
}

type mockRecorder struct{}

func (r *mockRecorder) Record(i []dtos.ImpressionsDTO, s string, m string, m2 string) error {
	return nil
}

func TestImpressionListener(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	gListener := goodListener{status: false}
	impStorage := mutexmap.NewMMImpressionStorage()
	impStorage.Put("feature1", &dtos.ImpressionDTO{
		BucketingKey: "aBucketingKey",
		ChangeNumber: 1,
		KeyName:      "aKey",
		Label:        "aLabel",
		Time:         1,
		Treatment:    "aTreatment",
	})

	submitImpressions(impStorage, &mockRecorder{}, "", "", "", &gListener, logger)
	time.Sleep(2 * time.Second)

	if !gListener.status {
		t.Error("Listener not called correctly")
	}

	bListener := badListener{}
	submitImpressions(impStorage, &mockRecorder{}, "", "", "", &bListener, logger)
	// Panic should be caught!
}
