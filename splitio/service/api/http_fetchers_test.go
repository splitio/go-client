// Package api contains all functions and dtos Split APIs
package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-client/splitio/util/logging"
)

var splitsMock, _ = ioutil.ReadFile("../../../testdata/splits_mock.json")
var splitMock, _ = ioutil.ReadFile("../../../testdata/split_mock.json")
var segmentMock, _ = ioutil.ReadFile("../../../testdata/segment_mock.json")

func TestSpitChangesFetch(t *testing.T) {

	logger := logging.NewLogger(&logging.LoggerOptions{
		CommonWriter: ioutil.Discard,
		ErrorWriter:  ioutil.Discard,
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, fmt.Sprintf(string(splitsMock), splitMock))
	}))
	defer ts.Close()

	os.Setenv(envSdkURLNamespace, ts.URL)
	os.Setenv(envEventsURLNamespace, ts.URL)

	splitFetcher := NewHTTPSplitFetcher(
		&configuration.SplitSdkConfig{},
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

	logger := logging.NewLogger(&logging.LoggerOptions{
		CommonWriter: ioutil.Discard,
		ErrorWriter:  ioutil.Discard,
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
	}))
	defer ts.Close()

	os.Setenv(envSdkURLNamespace, ts.URL)
	os.Setenv(envEventsURLNamespace, ts.URL)

	splitFetcher := NewHTTPSplitFetcher(
		&configuration.SplitSdkConfig{},
		logger,
	)

	_, err := splitFetcher.Fetch(-1)
	if err == nil {
		t.Error("Error expected but not found")
	}
}

func TestSegmentChangesFetch(t *testing.T) {

	logger := logging.NewLogger(&logging.LoggerOptions{
		CommonWriter: ioutil.Discard,
		ErrorWriter:  ioutil.Discard,
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, fmt.Sprintf(string(segmentMock)))
	}))
	defer ts.Close()

	os.Setenv(envSdkURLNamespace, ts.URL)
	os.Setenv(envEventsURLNamespace, ts.URL)

	segmentFetcher := NewHTTPSegmentFetcher(
		&configuration.SplitSdkConfig{},
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

	logger := logging.NewLogger(&logging.LoggerOptions{
		CommonWriter: ioutil.Discard,
		ErrorWriter:  ioutil.Discard,
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
	}))
	defer ts.Close()

	os.Setenv(envSdkURLNamespace, ts.URL)
	os.Setenv(envEventsURLNamespace, ts.URL)

	segmentFetcher := NewHTTPSegmentFetcher(
		&configuration.SplitSdkConfig{},
		logger,
	)

	_, err := segmentFetcher.Fetch("employees", -1)
	if err == nil {
		t.Error("Error expected but not found")
	}
}
