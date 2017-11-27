package api

import (
	"bytes"
	"encoding/json"
	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/util/configuration"
	"github.com/splitio/go-toolkit/logging"
	"strconv"
)

type httpFetcherBase struct {
	client *HTTPClient
	logger logging.LoggerInterface
}

func (h *httpFetcherBase) fetchRaw(url string, since int64) ([]byte, error) {
	var bufferQuery bytes.Buffer
	bufferQuery.WriteString(url)

	if since >= -1 {
		bufferQuery.WriteString("?since=")
		bufferQuery.WriteString(strconv.FormatInt(since, 10))
	}
	data, err := h.client.Get(bufferQuery.String())
	if err != nil {
		h.logger.Error("Error fetching raw data for url: ", url)
		return nil, err
	}
	return data, nil
}

// HTTPSplitFetcher struct is responsible for fetching splits from the backend via HTTP protocol
type HTTPSplitFetcher struct {
	httpFetcherBase
}

// NewHTTPSplitFetcher instantiates and return an HTTPSplitFetcher
func NewHTTPSplitFetcher(
	apikey string,
	cfg *configuration.SplitSdkConfig,
	logger logging.LoggerInterface,
) *HTTPSplitFetcher {
	sdkURL, _ := getUrls(cfg.Advanced)
	return &HTTPSplitFetcher{
		httpFetcherBase: httpFetcherBase{
			client: NewHTTPClient(apikey, cfg, sdkURL, splitio.Version, logger),
			logger: logger,
		},
	}
}

// Fetch makes an http call to the split backend and returns the list of updated splits
func (f *HTTPSplitFetcher) Fetch(since int64) (*dtos.SplitChangesDTO, error) {
	data, err := f.fetchRaw("/splitChanges", since)
	if err != nil {
		f.logger.Error("Error fetching split changes ", err)
		return nil, err
	}

	var splitChangesDto dtos.SplitChangesDTO
	err = json.Unmarshal(data, &splitChangesDto)
	if err != nil {
		f.logger.Error("Error parsing split changes JSON ", err)
		return nil, err
	}

	// RAW DATA --------------
	var objmap map[string]*json.RawMessage
	if err = json.Unmarshal(data, &objmap); err != nil {
		f.logger.Error(err)
		return nil, err
	}

	if err = json.Unmarshal(*objmap["splits"], &splitChangesDto.RawSplits); err != nil {
		f.logger.Error(err)
		return nil, err
	}
	//-------------------------
	return &splitChangesDto, nil
}

// HTTPSegmentFetcher struct is responsible for fetching segment by name from the API via HTTP method
type HTTPSegmentFetcher struct {
	httpFetcherBase
}

// NewHTTPSegmentFetcher instantiates and returns a new HTTPSegmentFetcher.
func NewHTTPSegmentFetcher(
	apikey string,
	cfg *configuration.SplitSdkConfig,
	logger logging.LoggerInterface,
) *HTTPSegmentFetcher {
	sdkURL, _ := getUrls(cfg.Advanced)
	return &HTTPSegmentFetcher{
		httpFetcherBase: httpFetcherBase{
			client: NewHTTPClient(apikey, cfg, sdkURL, splitio.Version, logger),
			logger: logger,
		},
	}
}

// Fetch issues a GET request to the split backend and returns the contents of a particular segment
func (f *HTTPSegmentFetcher) Fetch(segmentName string, since int64) (*dtos.SegmentChangesDTO, error) {
	var bufferQuery bytes.Buffer
	bufferQuery.WriteString("/segmentChanges/")
	bufferQuery.WriteString(segmentName)

	data, err := f.fetchRaw(bufferQuery.String(), since)
	if err != nil {
		//		f.logger.Error("Error fetching segment changes ", err)
		return nil, err
	}
	var segmentChangesDto dtos.SegmentChangesDTO
	err = json.Unmarshal(data, &segmentChangesDto)
	if err != nil {
		f.logger.Error("Error parsing segment changes JSON for segment ", segmentName, err)
		return nil, err
	}

	return &segmentChangesDto, nil
}
