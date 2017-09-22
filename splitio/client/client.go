package client

import (
	"github.com/splitio/go-client/splitio/util/logging"
)

// SplitClient ...
// SDK Client to get the treatment of a specific feature for a certain key and set of attributes
type SplitClient struct {
	Apikey       string
	Logger       logging.LoggerInterface
	LoggerConfig logging.LoggerOptions
}

// GetTreatment ...
// main functionality of split. Retrieve treatment of a specific feature
// for a certain key and set of attributes
func (c *SplitClient) GetTreatment(key string, feature string, attributes *map[string]interface{}) string {
	return "CONTROL"
}
