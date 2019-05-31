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
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/storage/mutexqueue"
	"github.com/splitio/go-toolkit/logging"
)

type impressionRecord struct {
	KeyName      string `json:"keyName"`
	Treatment    string `json:"treatment"`
	Time         int64  `json:"time"`
	ChangeNumber int64  `json:"changeNumber"`
	Label        string `json:"label"`
	BucketingKey string `json:"bucketingKey,omitempty"`
}

type impressionsRecord struct {
	TestName       string             `json:"testName"`
	KeyImpressions []impressionRecord `json:"keyImpressions"`
}

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

		var impressions []impressionsRecord
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

	impressionStorage := mutexqueue.NewMQImpressionsStorage(200, make(chan string, 1), logger)

	impressionTask := NewRecordImpressionsTask(
		impressionStorage,
		impressionRecorder,
		1,
		"go-0.1",
		"192.168.0.123",
		"machine1",
		logger,
		100,
	)

	impressionTask.Start()

	if !impressionTask.IsRunning() {
		t.Error("Impression recording task should be running")
	}

	imp1 := storage.Impression{
		FeatureName:  "feature1",
		BucketingKey: "123",
		ChangeNumber: 456,
		KeyName:      "key1",
		Time:         123,
		Treatment:    "on",
	}

	impressionStorage.LogImpressions([]storage.Impression{imp1})

	imp2 := storage.Impression{
		FeatureName:  "feature1",
		BucketingKey: "123",
		ChangeNumber: 456,
		KeyName:      "key2",
		Time:         124,
		Treatment:    "off",
	}

	impressionStorage.LogImpressions([]storage.Impression{imp2})

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

func (r *mockRecorder) Record(i []storage.Impression, s string, m string, m2 string) error {
	return nil
}

type impressionRecorderMock struct {
	iterations int
}

func (i *impressionRecorderMock) Record(
	impressions []storage.Impression,
	sdkVersion string,
	machineIP string,
	machineName string,
) error {
	i.iterations++
	return nil
}

func TestImpressionsFlushWhenTaskIsStopped(t *testing.T) {
	logger := logging.NewLogger(nil)
	imp1 := storage.Impression{
		FeatureName:  "feature1",
		BucketingKey: "123",
		ChangeNumber: 456,
		KeyName:      "key1",
		Time:         123,
		Treatment:    "on",
	}
	impressionStorage := mutexqueue.NewMQImpressionsStorage(200, make(chan string, 1), logger)
	impressionStorage.LogImpressions([]storage.Impression{imp1})
	impressionStorage.LogImpressions([]storage.Impression{imp1})
	impressionStorage.LogImpressions([]storage.Impression{imp1})
	impressionStorage.LogImpressions([]storage.Impression{imp1})
	impressionRecorder := &impressionRecorderMock{}
	impressionTask := NewRecordImpressionsTask(
		impressionStorage,
		impressionRecorder,
		100,
		"aa",
		"123.123.123.123",
		"123-123-123-123",
		logger,
		100,
	)

	impressionTask.Start()
	time.Sleep(time.Second * 2)

	if impressionRecorder.iterations != 1 {
		t.Error("Impressions should already have been flushed once")
	}

	// Add more impressions so that they can be flushed when Stop() is called
	imp2 := storage.Impression{
		FeatureName:  "feature2",
		BucketingKey: "123",
		ChangeNumber: 456,
		KeyName:      "key1",
		Time:         123,
		Treatment:    "on",
	}
	impressionStorage.LogImpressions([]storage.Impression{imp2})
	impressionStorage.LogImpressions([]storage.Impression{imp2})
	impressionStorage.LogImpressions([]storage.Impression{imp2})
	impressionStorage.LogImpressions([]storage.Impression{imp2})
	impressionTask.Stop()
	time.Sleep(time.Second * 2)
	if impressionRecorder.iterations != 2 {
		t.Error("Impression Task should have ran twice")
	}
}
