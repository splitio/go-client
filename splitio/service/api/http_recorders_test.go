// Package api contains all functions and dtos Split APIs
package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-client/splitio/util/logging"
)

func TestPostImpressions(t *testing.T) {

	logger := logging.NewLogger(&logging.LoggerOptions{
		ErrorWriter:  ioutil.Discard,
		CommonWriter: ioutil.Discard,
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sdkVersion := r.Header.Get("SplitSDKVersion")
		sdkMachine := r.Header.Get("SplitSDKMachineIP")

		if sdkVersion != "test-1.0.0" {
			t.Error("SDK Version HEADER not match")
		}

		if sdkMachine != "127.0.0.1" {
			t.Error("SDK Machine HEADER not match")
		}

		sdkMachineName := r.Header.Get("SplitSDKMachineName")
		if sdkMachineName != "SOME_MACHINE_NAME" {
			t.Error("SDK Machine Name HEADER not match", sdkMachineName)
		}

		rBody, _ := ioutil.ReadAll(r.Body)
		//fmt.Println(string(rBody))
		var impressionsInPost []dtos.ImpressionsDTO
		err := json.Unmarshal(rBody, &impressionsInPost)
		if err != nil {
			t.Error(err)
			return
		}

		if impressionsInPost[0].TestName != "some_test" ||
			impressionsInPost[0].KeyImpressions[0].KeyName != "some_key_1" ||
			impressionsInPost[0].KeyImpressions[1].KeyName != "some_key_2" {
			t.Error("Posted impressions arrived mal-formed")
		}

		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	os.Setenv(envSdkURLNamespace, ts.URL)
	os.Setenv(envEventsURLNamespace, ts.URL)

	imp1 := dtos.ImpressionDTO{
		KeyName:      "some_key_1",
		Treatment:    "on",
		Time:         1234567890,
		ChangeNumber: 9876543210,
		Label:        "some_label_1",
		BucketingKey: "some_bucket_key_1",
	}
	imp2 := dtos.ImpressionDTO{
		KeyName:      "some_key_2",
		Treatment:    "off",
		Time:         1234567890,
		ChangeNumber: 9876543210,
		Label:        "some_label_2",
		BucketingKey: "some_bucket_key_2",
	}

	keyImpressions := make([]dtos.ImpressionDTO, 0)
	keyImpressions = append(keyImpressions, imp1, imp2)
	impressionsTest := dtos.ImpressionsDTO{
		TestName:       "some_test",
		KeyImpressions: keyImpressions,
	}

	impressions := make([]dtos.ImpressionsDTO, 0)
	impressions = append(impressions, impressionsTest)

	impressionRecorder := NewHTTPImpressionRecorder(
		&configuration.SplitSdkConfig{},
		logger,
	)
	err2 := impressionRecorder.Record(impressions, "test-1.0.0", "127.0.0.1", "SOME_MACHINE_NAME")
	if err2 != nil {
		t.Error(err2)
	}
}

func TestPostMetricsLatency(t *testing.T) {

	logger := logging.NewLogger(&logging.LoggerOptions{
		ErrorWriter:  ioutil.Discard,
		CommonWriter: ioutil.Discard,
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sdkVersion := r.Header.Get("SplitSDKVersion")
		sdkMachine := r.Header.Get("SplitSDKMachineIP")

		if sdkVersion != "test-1.0.0" {
			t.Error("SDK Version HEADER not match")
		}

		if sdkMachine != "127.0.0.1" {
			t.Error("SDK Machine HEADER not match")
		}

		sdkMachineName := r.Header.Get("SplitSDKMachineName")
		if sdkMachineName != "ip-127-0-0-1" {
			t.Error("SDK Machine Name HEADER not match", sdkMachineName)
		}

		rBody, _ := ioutil.ReadAll(r.Body)
		var latenciesInPost []dtos.LatenciesDTO
		err := json.Unmarshal(rBody, &latenciesInPost)
		if err != nil {
			t.Error(err)
			return
		}

		if latenciesInPost[0].MetricName != "some_metric_name" ||
			latenciesInPost[0].Latencies[5] != 1234567890 {
			t.Error("Latencies arrived mal-formed")
		}

		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	os.Setenv(envSdkURLNamespace, ts.URL)
	os.Setenv(envEventsURLNamespace, ts.URL)

	var latencyValues = make([]int64, 23) //23 maximun number of buckets
	latencyValues[5] = 1234567890
	var latencies []dtos.LatenciesDTO
	latencies = append(latencies, dtos.LatenciesDTO{MetricName: "some_metric_name", Latencies: latencyValues})

	metricsRecorder := NewHTTPMetricsRecorder(
		&configuration.SplitSdkConfig{},
		logger,
	)
	err2 := metricsRecorder.RecordLatencies(latencies, "test-1.0.0", "127.0.0.1", "ip-127-0-0-1")
	if err2 != nil {
		t.Error(err2)
	}
}

func TestPostMetricsCounters(t *testing.T) {

	logger := logging.NewLogger(&logging.LoggerOptions{
		ErrorWriter:  ioutil.Discard,
		CommonWriter: ioutil.Discard,
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sdkVersion := r.Header.Get("SplitSDKVersion")
		sdkMachine := r.Header.Get("SplitSDKMachineIP")

		if sdkVersion != "test-1.0.0" {
			t.Error("SDK Version HEADER not match")
		}

		if sdkMachine != "127.0.0.1" {
			t.Error("SDK Machine HEADER not match")
		}

		sdkMachineName := r.Header.Get("SplitSDKMachineName")
		if sdkMachineName != "ip-127-0-0-1" {
			t.Error("SDK Machine Name HEADER not match", sdkMachineName)
		}

		rBody, _ := ioutil.ReadAll(r.Body)
		var countersInPost []dtos.CounterDTO
		err := json.Unmarshal(rBody, &countersInPost)
		if err != nil {
			t.Error(err)
			return
		}

		if countersInPost[0].MetricName != "counter_1" ||
			countersInPost[0].Count != 111 ||
			countersInPost[1].MetricName != "counter_2" ||
			countersInPost[1].Count != 222 {
			t.Error("Counters arrived mal-formed")
		}

		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	os.Setenv(envSdkURLNamespace, ts.URL)
	os.Setenv(envEventsURLNamespace, ts.URL)

	var counters []dtos.CounterDTO
	counters = append(
		counters,
		dtos.CounterDTO{MetricName: "counter_1", Count: 111},
		dtos.CounterDTO{MetricName: "counter_2", Count: 222},
	)

	metricsRecorder := NewHTTPMetricsRecorder(
		&configuration.SplitSdkConfig{},
		logger,
	)
	err2 := metricsRecorder.RecordCounters(counters, "test-1.0.0", "127.0.0.1", "ip-127-0-0-1")
	if err2 != nil {
		t.Error(err2)
	}
}

func TestPostMetricsGauge(t *testing.T) {

	logger := logging.NewLogger(&logging.LoggerOptions{
		ErrorWriter:  ioutil.Discard,
		CommonWriter: ioutil.Discard,
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sdkVersion := r.Header.Get("SplitSDKVersion")
		sdkMachine := r.Header.Get("SplitSDKMachineIP")

		if sdkVersion != "test-1.0.0" {
			t.Error("SDK Version HEADER not match")
		}

		if sdkMachine != "127.0.0.1" {
			t.Error("SDK Machine HEADER not match")
		}

		sdkMachineName := r.Header.Get("SplitSDKMachineName")
		if sdkMachineName != "ip-127-0-0-1" {
			t.Error("SDK Machine Name HEADER not match", sdkMachineName)
		}

		rBody, _ := ioutil.ReadAll(r.Body)
		var gaugesInPost dtos.GaugeDTO
		err := json.Unmarshal(rBody, &gaugesInPost)
		if err != nil {
			t.Error(err)
			return
		}

		if gaugesInPost.MetricName != "gauge_1" ||
			gaugesInPost.Gauge != 111.1 {
			t.Error("Gauges arrived mal-formed")
		}

		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	os.Setenv(envSdkURLNamespace, ts.URL)
	os.Setenv(envEventsURLNamespace, ts.URL)

	var gauge dtos.GaugeDTO
	gauge = dtos.GaugeDTO{MetricName: "gauge_1", Gauge: 111.1}

	metricsRecorder := NewHTTPMetricsRecorder(
		&configuration.SplitSdkConfig{},
		logger,
	)
	err2 := metricsRecorder.RecordGauge(gauge, "test-1.0.0", "127.0.0.1", "ip-127-0-0-1")
	if err2 != nil {
		t.Error(err2)
	}

}