package client

import (
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"
)

// SplitClient is the entry-point of the split SDK.
type SplitClient struct {
	apikey       string
	logger       logging.LoggerInterface
	loggerConfig logging.LoggerOptions
	evaluator    *evaluator.Evaluator
	sync         sdkSync
}

type sdkSync struct {
	splitSync      *asynctask.AsyncTask
	segmentSync    *asynctask.AsyncTask
	impressionSync *asynctask.AsyncTask
	gaugeSync      *asynctask.AsyncTask
	countersSync   *asynctask.AsyncTask
	latenciesSync  *asynctask.AsyncTask
}

// Treatment implements the main functionality of split. Retrieve treatments of a specific feature
// for a certain key and set of attributes
func (c *SplitClient) Treatment(key string, feature string, attributes *map[string]interface{}) string {
	return "CONTROL"
}
