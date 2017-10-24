package api

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-client/splitio/util/logging"
)

const prodSdkURL = "https://sdk.split.io/api"
const prodEventsURL = "https://events.split.io/api"

func getUrls(cfg *configuration.AdvancedConfig) (sdkURL string, eventsURL string) {
	if cfg != nil && cfg.SdkURL != "" {
		sdkURL = cfg.SdkURL
	} else {
		sdkURL = prodSdkURL
	}

	if cfg != nil && cfg.EventsURL != "" {
		eventsURL = cfg.EventsURL
	} else {
		eventsURL = prodEventsURL
	}
	return sdkURL, eventsURL
}

// HTTPClient structure to wrap up the net/http.Client
type HTTPClient struct {
	url        string
	httpClient *http.Client
	headers    map[string]string
	logger     logging.LoggerInterface
	apikey     string
}

// NewHTTPClient instance of HttpClient
func NewHTTPClient(
	cfg *configuration.SplitSdkConfig,
	endpoint string,
	logger logging.LoggerInterface,
) *HTTPClient {
	client := &http.Client{Timeout: time.Duration(cfg.HTTPTimeout) * time.Second}
	return &HTTPClient{
		url:        endpoint,
		httpClient: client,
		headers:    make(map[string]string),
		logger:     logger,
		apikey:     cfg.Apikey,
	}
}

// Get method is a get call to an url
func (c *HTTPClient) Get(service string) ([]byte, error) {

	serviceURL := c.url + service
	c.logger.Debug("[GET] ", serviceURL)
	req, _ := http.NewRequest("GET", serviceURL, nil)

	authorization := c.apikey
	c.logger.Debug("Authorization [ApiKey]: ", logging.ObfuscateAPIKey(authorization))

	req.Header.Add("Authorization", "Bearer "+authorization)
	req.Header.Add("SplitSDKVersion", "go-0.0.1")
	req.Header.Add("User-Agent", "SplitIO-GO-AGENT/0.1")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Error requesting data to API: ", req.URL.String(), err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	// Check that the server actually sent compressed data
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		c.logger.Error(err.Error())
		return nil, err
	}

	c.logger.Debug("[RESPONSE_BODY]", string(body), "[END_RESPONSE_BODY]")

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return body, nil
	}

	return nil, fmt.Errorf("GET method: Status Code: %d - %s", resp.StatusCode, resp.Status)
}

// Post performs a HTTP POST request
func (c *HTTPClient) Post(service string, body []byte) error {

	serviceURL := c.url + service
	c.logger.Debug("[POST] ", serviceURL)
	req, _ := http.NewRequest("POST", serviceURL, bytes.NewBuffer(body))
	//****************
	req.Close = true // To prevent EOF error when connection is closed
	//****************
	authorization := c.apikey
	c.logger.Debug("Authorization [ApiKey]: ", logging.ObfuscateAPIKey(authorization))

	req.Header.Add("Authorization", "Bearer "+authorization)
	//SplitSDKVersion added by poster tasks
	req.Header.Add("User-Agent", "SplitIO-GO-SDK/1.0")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Content-Type", "application/json")

	for headerName, headerValue := range c.headers {
		req.Header.Add(headerName, headerValue)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Error requesting data to API: ", req.URL.String(), err.Error())
		return err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}

	c.logger.Debug("[RESPONSE_BODY]", string(respBody), "[END_RESPONSE_BODY]")

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("POST method: Status Code: %d - %s", resp.StatusCode, resp.Status)
}

// AddHeader adds header value to HTTP client
func (c *HTTPClient) AddHeader(name string, value string) {
	c.headers[name] = value
}

// ResetHeaders resets custom headers
func (c *HTTPClient) ResetHeaders() {
	c.headers = make(map[string]string)
}
