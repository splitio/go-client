// Package api contains all functions and dtos Split APIs
package api

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-client/splitio/util/logging"
)

func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	logger := logging.NewLogger(&logging.LoggerOptions{})
	httpClient := NewHTTPClient(&configuration.SplitSdkConfig{}, ts.URL, logger)
	txt, errg := httpClient.Get("/")
	if errg != nil {
		t.Error(errg)
	}

	if string(txt) != "Hello, client\n" {
		t.Error("Given message failed ")
	}
}

func TestGetGZIP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Encoding", "gzip")

		gzw := gzip.NewWriter(w)
		defer gzw.Close()
		fmt.Fprintln(gzw, "Hello, client")
	}))
	defer ts.Close()

	logger := logging.NewLogger(&logging.LoggerOptions{})
	httpClient := NewHTTPClient(&configuration.SplitSdkConfig{}, ts.URL, logger)
	txt, errg := httpClient.Get("/")
	if errg != nil {
		t.Error(errg)
	}

	if string(txt) != "Hello, client\n" {
		t.Error("Given message failed ")
	}
}

func TestPost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	logger := logging.NewLogger(&logging.LoggerOptions{})
	httpClient := NewHTTPClient(&configuration.SplitSdkConfig{}, ts.URL, logger)
	httpClient.AddHeader("someHeader", "HeaderValue")
	errp := httpClient.Post("/", []byte("some text"))
	if errp != nil {
		t.Error(errp)
	}
}

func TestHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	logger := logging.NewLogger(&logging.LoggerOptions{})
	httpClient := NewHTTPClient(&configuration.SplitSdkConfig{}, ts.URL, logger)
	httpClient.AddHeader("someHeader", "HeaderValue")
	_, ok1 := httpClient.headers["someHeader"]
	if !ok1 {
		t.Error("Header could not be added")
	}

	httpClient.ResetHeaders()
	_, ok2 := httpClient.headers["someHeader"]
	if ok2 {
		t.Error("Reset Header fails")
	}
}
