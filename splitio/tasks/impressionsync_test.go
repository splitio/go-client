package tasks

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/service/api"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-toolkit/logging"
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

type mockRecorder struct{}

func (r *mockRecorder) Record(i []dtos.ImpressionsDTO, s string, m string, m2 string) error {
	return nil
}

type impressionRecorderMock struct {
	iterations int
}

func (i *impressionRecorderMock) Record(
	impressions []dtos.ImpressionsDTO,
	sdkVersion string,
	machineIP string,
	machineName string,
) error {
	i.iterations++
	return nil
}

func TestImpressionsFlushWhenTaskIsStopped(t *testing.T) {
	logger := logging.NewLogger(nil)
	impressionStorage := mutexmap.NewMMImpressionStorage()
	impressionStorage.Put("feature1", &dtos.ImpressionDTO{})
	impressionStorage.Put("feature1", &dtos.ImpressionDTO{})
	impressionStorage.Put("feature1", &dtos.ImpressionDTO{})
	impressionStorage.Put("feature1", &dtos.ImpressionDTO{})
	impressionRecorder := &impressionRecorderMock{}
	impressionTask := NewRecordImpressionsTask(
		impressionStorage,
		impressionRecorder,
		100,
		"aa",
		"123.123.123.123",
		"123-123-123-123",
		logger,
	)

	impressionTask.Start()
	time.Sleep(time.Second * 2)

	if impressionRecorder.iterations != 1 {
		t.Error("Impressions should already have been flushed once")
	}

	// Add more impressions so that they can be flushed when Stop() is called
	impressionStorage.Put("feature2", &dtos.ImpressionDTO{})
	impressionStorage.Put("feature2", &dtos.ImpressionDTO{})
	impressionStorage.Put("feature2", &dtos.ImpressionDTO{})
	impressionStorage.Put("feature2", &dtos.ImpressionDTO{})
	impressionTask.Stop()
	time.Sleep(time.Second * 2)
	if impressionRecorder.iterations != 2 {
		t.Error("Impression Task should have ran twice")
	}
}
