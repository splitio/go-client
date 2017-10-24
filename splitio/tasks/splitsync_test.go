package tasks

import (
	"encoding/json"
	"github.com/splitio/go-client/splitio/service/api"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-client/splitio/util/logging"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSplitSyncTask(t *testing.T) {

	mockedSplit1 := dtos.SplitDTO{Name: "split1", Killed: false}
	mockedSplit2 := dtos.SplitDTO{Name: "split2", Killed: true}

	reqestReceived := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/splits" && r.Method != "GET" {
			t.Error("Invalid request. Should be GET to /splits")
		}
		reqestReceived = true

		splitChanges := dtos.SplitChangesDTO{
			Splits: []dtos.SplitDTO{mockedSplit1, mockedSplit2},
		}

		raw, err := json.Marshal(splitChanges)
		if err != nil {
			t.Error("Error building json")
			return
		}

		w.Write(raw)
	}))
	defer ts.Close()

	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitFetcher := api.NewHTTPSplitFetcher(
		&configuration.SplitSdkConfig{
			Apikey: "123",
			Advanced: &configuration.AdvancedConfig{
				EventsURL: ts.URL,
				SdkURL:    ts.URL,
			},
		},
		logger,
	)

	splitStorage := storage.NewMMSplitStorage()

	splitTask := NewFetchSplitsTask(
		splitStorage,
		splitFetcher,
		1000,
		logger,
	)

	splitTask.Start()

	if !splitTask.IsRunning() {
		t.Error("Split fetching task should be running")
	}

	time.Sleep(time.Second * 2)

	if !reqestReceived {
		t.Error("Request not received")
	}

	splitTask.Stop()

	time.Sleep(time.Second * 2)

	s1 := splitStorage.Get("split1")
	if s1 == nil || s1.Name != "split1" || s1.Killed != false {
		t.Error("split1 stored/retrieved incorrectly")
	}

	s2 := splitStorage.Get("split2")
	if s2 == nil || s2.Name != "split2" || s2.Killed != true {
		t.Error("split2 stored/retrieved incorrectly")
		t.Error(s2)
	}

	if splitTask.IsRunning() {
		t.Error("Task should be stopped")
	}
}
