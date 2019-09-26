package api

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-toolkit/logging"
)

const prodSdkURL = "https://sdk.split.io/api"
const prodEventsURL = "https://events.split.io/api"
const defaultHTTPTimeout = 30

func getUrls(cfg *conf.AdvancedConfig) (sdkURL string, eventsURL string) {
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
	version    string
}

// NewHTTPClient instance of HttpClient
func NewHTTPClient(
	apikey string,
	cfg *conf.SplitSdkConfig,
	endpoint string,
	version string,
	logger logging.LoggerInterface,
) *HTTPClient {
	var timeout int
	if cfg.Advanced.HTTPTimeout != 0 {
		timeout = cfg.Advanced.HTTPTimeout
	} else {
		timeout = defaultHTTPTimeout
	}
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	return &HTTPClient{
		url:        endpoint,
		httpClient: client,
		logger:     logger,
		apikey:     apikey,
		version:    version,
	}
}

// Get method is a get call to an url
func (c *HTTPClient) Get(service string) ([]byte, error) {

	serviceURL := c.url + service
	c.logger.Debug("[GET] ", serviceURL)
	req, _ := http.NewRequest("GET", serviceURL, nil)

	authorization := c.apikey
	c.logger.Debug("Authorization [ApiKey]: ", logging.ObfuscateAPIKey(authorization))
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Content-Type", "application/json")

	c.logger.Debug(fmt.Sprintf("Headers: %v", req.Header))

	req.Header.Add("Authorization", "Bearer "+authorization)

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

	c.logger.Verbose("[RESPONSE_BODY]", string(body), "[END_RESPONSE_BODY]")

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return body, nil
	}

	return nil, fmt.Errorf("GET method: Status Code: %d - %s", resp.StatusCode, resp.Status)
}

// Post performs a HTTP POST request
func (c *HTTPClient) Post(service string, body []byte, headers map[string]string) error {

	serviceURL := c.url + service
	c.logger.Debug("[POST] ", serviceURL)
	req, _ := http.NewRequest("POST", serviceURL, bytes.NewBuffer(body))
	//****************
	req.Close = true // To prevent EOF error when connection is closed
	//****************
	authorization := c.apikey
	c.logger.Debug("Authorization [ApiKey]: ", logging.ObfuscateAPIKey(authorization))

	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Content-Type", "application/json")

	for headerName, headerValue := range headers {
		req.Header.Add(headerName, headerValue)
	}

	c.logger.Debug(fmt.Sprintf("Headers: %v", req.Header))

	req.Header.Add("Authorization", "Bearer "+authorization)

	c.logger.Verbose("[REQUEST_BODY]", string(body), "[END_REQUEST_BODY]")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Error posting data to API: ", req.URL.String(), err.Error())
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}

	c.logger.Verbose("[RESPONSE_BODY]", string(respBody), "[END_RESPONSE_BODY]")

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("POST method: Status Code: %d - %s", resp.StatusCode, resp.Status)
}

// ValidateApikey validates apikey
func ValidateApikey(apikey string, config conf.AdvancedConfig) error {
	sdkURL, _ := getUrls(&config)
	client := &http.Client{}

	req, _ := http.NewRequest("GET", sdkURL+"/segmentChanges/___TEST___?since=-1", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+apikey)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return errors.New("you passed a browser type apikey, please grab an apikey from the Split console that is of type sdk")
	}

	return nil
}
