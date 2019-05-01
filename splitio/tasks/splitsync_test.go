package tasks

import (
	"encoding/json"
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

func TestSplitSyncTask(t *testing.T) {
	mockedSplit1 := dtos.SplitDTO{Name: "split1", Killed: false, Status: "ACTIVE", TrafficTypeName: "one"}
	mockedSplit2 := dtos.SplitDTO{Name: "split2", Killed: true, Status: "ACTIVE", TrafficTypeName: "two"}
	mockedSplit3 := dtos.SplitDTO{Name: "split3", Killed: true, Status: "INACTIVE", TrafficTypeName: "one"}
	reqestReceived := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/splits" && r.Method != "GET" {
			t.Error("Invalid request. Should be GET to /splits")
		}
		reqestReceived = true

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

	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitFetcher := api.NewHTTPSplitFetcher(
		"",
		&conf.SplitSdkConfig{
			Advanced: conf.AdvancedConfig{
				EventsURL: ts.URL,
				SdkURL:    ts.URL,
			},
		},
		logger,
	)

	splitStorage := mutexmap.NewMMSplitStorage()
	splitStorage.PutMany([]dtos.SplitDTO{}, -1)

	readyChannel := make(chan string)
	splitTask := NewFetchSplitsTask(
		splitStorage,
		splitFetcher,
		3,
		logger,
		readyChannel,
	)

	splitTask.Start()

	if !splitTask.IsRunning() {
		t.Error("Split fetching task should be running")
	}

	select {
	case msg := <-readyChannel:
		if msg != "SPLITS_READY" {
			t.Error("Incorrect msg receieved")
			return
		}
		break
	case <-time.After(3 * time.Second):
		t.Error("SPLITS_READY signal not received")
		return
	}

	if !reqestReceived {
		t.Error("Request not received")
	}

	splitTask.Stop()

	time.Sleep(time.Second * 10)

	s1 := splitStorage.Get("split1")
	if s1 == nil || s1.Name != "split1" || s1.Killed {
		t.Error("split1 stored/retrieved incorrectly")
		t.Error(s1)
	}

	if !splitStorage.TrafficTypeExists("one") {
		t.Error("It should exists")
	}

	s2 := splitStorage.Get("split2")
	if s2 == nil || s2.Name != "split2" || !s2.Killed {
		t.Error("split2 stored/retrieved incorrectly")
		t.Error(s2)
	}

	s3 := splitStorage.Get("split3")
	if s3 != nil {
		t.Error("split3 should have been removed")
	}

	if !splitStorage.TrafficTypeExists("two") {
		t.Error("It should exists")
	}

	if splitTask.IsRunning() {
		t.Error("Task should be stopped")
	}
}

func TestSplitSyncTaskStatus(t *testing.T) {
	mockedSplit1 := dtos.SplitDTO{Name: "split1", Killed: false, Status: "ACTIVE", TrafficTypeName: "one"}
	mockedSplit2 := dtos.SplitDTO{Name: "split2", Killed: true, Status: "ACTIVE", TrafficTypeName: "two"}
	mockedSplit3 := dtos.SplitDTO{Name: "split3", Killed: true, Status: "INACTIVE", TrafficTypeName: "one"}
	mockedSplit4 := dtos.SplitDTO{Name: "split1", Killed: true, Status: "INACTIVE", TrafficTypeName: "one"}
	mockedSplit5 := dtos.SplitDTO{Name: "split4", Killed: false, Status: "ACTIVE", TrafficTypeName: "two"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/splits" && r.Method != "GET" {
			t.Error("Invalid request. Should be GET to /splits")
		}

		splitChanges := dtos.SplitChangesDTO{
			Splits: []dtos.SplitDTO{mockedSplit4, mockedSplit5},
			Since:  4,
			Till:   4,
		}

		raw, err := json.Marshal(splitChanges)
		if err != nil {
			t.Error("Error building json")
			return
		}

		w.Write(raw)
	}))
	defer ts2.Close()

	logger := logging.NewLogger(&logging.LoggerOptions{})
	splitFetcher := api.NewHTTPSplitFetcher(
		"",
		&conf.SplitSdkConfig{
			Advanced: conf.AdvancedConfig{
				EventsURL: ts.URL,
				SdkURL:    ts.URL,
			},
		},
		logger,
	)

	splitStorage := mutexmap.NewMMSplitStorage()
	splitStorage.PutMany([]dtos.SplitDTO{{}}, -1)

	updateSplits(splitStorage, splitFetcher)

	if !splitStorage.TrafficTypeExists("one") {
		t.Error("It should exists")
	}

	if !splitStorage.TrafficTypeExists("two") {
		t.Error("It should exists")
	}

	splitFetcher2 := api.NewHTTPSplitFetcher(
		"",
		&conf.SplitSdkConfig{
			Advanced: conf.AdvancedConfig{
				EventsURL: ts2.URL,
				SdkURL:    ts2.URL,
			},
		},
		logger,
	)

	updateSplits(splitStorage, splitFetcher2)

	s1 := splitStorage.Get("split1")
	if s1 != nil {
		t.Error("split1 should have been removed")
	}

	s2 := splitStorage.Get("split2")
	if s2 == nil || s2.Name != "split2" || !s2.Killed {
		t.Error("split2 stored/retrieved incorrectly")
		t.Error(s2)
	}

	s4 := splitStorage.Get("split4")
	if s4 == nil || s4.Name != "split4" || s4.Killed {
		t.Error("split4 stored/retrieved incorrectly")
		t.Error(s4)
	}

	s3 := splitStorage.Get("split3")
	if s3 != nil {
		t.Error("split3 should have been removed")
	}

	if splitStorage.TrafficTypeExists("one") {
		t.Error("It should not exists")
	}

	if !splitStorage.TrafficTypeExists("two") {
		t.Error("It should exists")
	}
}
