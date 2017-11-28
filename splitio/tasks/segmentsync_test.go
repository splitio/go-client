package tasks

import (
	"encoding/json"
	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/service/api"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage/mutexmap"
	"github.com/splitio/go-toolkit/logging"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSegmentSyncTask(t *testing.T) {

	addedS1 := []string{"item1", "item2", "item3", "item4"}
	addedS2 := []string{"item5", "item6", "item7", "item8"}

	s1RequestReceieved := false
	s2RequestReceieved := false
	var toReturn []string
	var name string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/segmentChanges/s1":
			s1RequestReceieved = true
			toReturn = addedS1
			name = "s1"
		case "/segmentChanges/s2":
			s2RequestReceieved = true
			toReturn = addedS2
			name = "s2"
		default:
			t.Errorf("Invalid URL %s", r.URL.Path)
		}

		segmentChanges := dtos.SegmentChangesDTO{
			Added:   toReturn,
			Name:    name,
			Removed: []string{},
			Since:   123,
			Till:    123,
		}

		raw, err := json.Marshal(segmentChanges)
		if err != nil {
			t.Error("Error building json")
			return
		}

		w.Write(raw)
	}))
	defer ts.Close()

	logger := logging.NewLogger(&logging.LoggerOptions{})
	segmentFetcher := api.NewHTTPSegmentFetcher(
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
	splitStorage.PutMany([]dtos.SplitDTO{
		{
			Name: "split1",
			Conditions: []dtos.ConditionDTO{
				{
					ConditionType: "WHITELIST",
					Label:         "Cond1",
					MatcherGroup: dtos.MatcherGroupDTO{
						Combiner: "AND",
						Matchers: []dtos.MatcherDTO{
							{
								UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
									SegmentName: "s1",
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "split2",
			Conditions: []dtos.ConditionDTO{
				{
					ConditionType: "WHITELIST",
					Label:         "Cond1",
					MatcherGroup: dtos.MatcherGroupDTO{
						Combiner: "AND",
						Matchers: []dtos.MatcherDTO{
							{
								UserDefinedSegment: &dtos.UserDefinedSegmentMatcherDataDTO{
									SegmentName: "s2",
								},
							},
						},
					},
				},
			},
		},
	}, 123)

	segmentStorage := mutexmap.NewMMSegmentStorage()

	readyChannel := make(chan string)
	segmentTask := NewFetchSegmentsTask(
		splitStorage,
		segmentStorage,
		segmentFetcher,
		1,
		5,
		100,
		logger,
		readyChannel,
	)

	segmentTask.Start()

	if !segmentTask.IsRunning() {
		t.Error("Split fetching task should be running")
	}

	select {
	case msg := <-readyChannel:
		if msg != "SEGMENTS_READY" {
			t.Error("Incorrect msg receieved")
			return
		}
	case <-time.After(3 * time.Second):
		t.Error("SEGMENTS_READY signal not received")
		return
	}

	if !s1RequestReceieved || !s2RequestReceieved {
		t.Error("Request not received")
	}

	segmentTask.Stop()

	time.Sleep(time.Second * 5)

	// By now, the segment fetching task should have retrieved and stored segments s1 and s2
	s1 := segmentStorage.Get("s1")
	if s1 == nil || !s1.Has("item1") {
		t.Error("Segment S1 stored/retrieved incorrectly")
	}

	s2 := segmentStorage.Get("s2")
	if s2 == nil || !s2.Has("item5") {
		t.Error("Segment S2 stored/retrieved incorrectly")
	}

	if segmentTask.IsRunning() {
		t.Error("Task should be stopped")
	}
}
