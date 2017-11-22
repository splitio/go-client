// Package api contains all functions and dtos Split APIs
package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-toolkit/logging"
)

var splitsMock, _ = ioutil.ReadFile("../../../testdata/splits_mock.json")
var splitMock, _ = ioutil.ReadFile("../../../testdata/split_mock.json")
var segmentMock, _ = ioutil.ReadFile("../../../testdata/segment_mock.json")

func TestSpitChangesFetch(t *testing.T) {

	logger := logging.NewLogger(&logging.LoggerOptions{})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, fmt.Sprintf(string(splitsMock), splitMock))
	}))
	defer ts.Close()

	splitFetcher := NewHTTPSplitFetcher(
		&configuration.SplitSdkConfig{
			Advanced: &configuration.AdvancedConfig{
				EventsURL: ts.URL,
				SdkURL:    ts.URL,
			},
		},
		logger,
	)

	splitChangesDTO, err := splitFetcher.Fetch(-1)
	if err != nil {
		t.Error(err)
	}

	if splitChangesDTO.Till != 1491244291288 ||
		splitChangesDTO.Splits[0].Name != "DEMO_MURMUR2" {
		t.Error("DTO mal formed")
	}
}

func TestSpitChangesFetchHTTPError(t *testing.T) {

	logger := logging.NewLogger(&logging.LoggerOptions{})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
	}))
	defer ts.Close()

	splitFetcher := NewHTTPSplitFetcher(
		&configuration.SplitSdkConfig{
			Advanced: &configuration.AdvancedConfig{
				EventsURL: ts.URL,
				SdkURL:    ts.URL,
			},
		},
		logger,
	)

	_, err := splitFetcher.Fetch(-1)
	if err == nil {
		t.Error("Error expected but not found")
	}
}

func TestSegmentChangesFetch(t *testing.T) {

	logger := logging.NewLogger(&logging.LoggerOptions{})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, fmt.Sprintf(string(segmentMock)))
	}))
	defer ts.Close()

	segmentFetcher := NewHTTPSegmentFetcher(
		&configuration.SplitSdkConfig{
			Advanced: &configuration.AdvancedConfig{
				EventsURL: ts.URL,
				SdkURL:    ts.URL,
			},
		},
		logger,
	)

	segmentFetched, err := segmentFetcher.Fetch("employees", -1)
	if err != nil {
		t.Error("Error fetching segment", err)
		return
	}
	if segmentFetched.Name != "employees" {
		t.Error("Fetched segment mal-formed")
	}
}

func TestSegmentChangesFetchHTTPError(t *testing.T) {

	logger := logging.NewLogger(&logging.LoggerOptions{})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
	}))
	defer ts.Close()

	segmentFetcher := NewHTTPSegmentFetcher(
		&configuration.SplitSdkConfig{
			Advanced: &configuration.AdvancedConfig{
				EventsURL: ts.URL,
				SdkURL:    ts.URL,
			},
		},
		logger,
	)

	_, err := segmentFetcher.Fetch("employees", -1)
	if err == nil {
		t.Error("Error expected but not found")
	}
}
