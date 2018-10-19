package client

import (
	"errors"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/util/metrics"

	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"
)

// SplitClient is the entry-point of the split SDK.
type SplitClient struct {
	apikey    string
	destroyed struct {
		status bool
		mutex  sync.RWMutex
	}
	cfg          *conf.SplitSdkConfig
	logger       logging.LoggerInterface
	loggerConfig logging.LoggerOptions
	evaluator    evaluator.Interface
	sync         *sdkSync
	impressions  storage.ImpressionStorageProducer
	metrics      storage.MetricsStorageProducer
	events       storage.EventStorageProducer
	validator    inputValidation
}

type sdkSync struct {
	splitSync      *asynctask.AsyncTask
	segmentSync    *asynctask.AsyncTask
	impressionSync *asynctask.AsyncTask
	gaugeSync      *asynctask.AsyncTask
	countersSync   *asynctask.AsyncTask
	latenciesSync  *asynctask.AsyncTask
	eventsSync     *asynctask.AsyncTask
}

// Treatment implements the main functionality of split. Retrieve treatments of a specific feature
// for a certain key and set of attributes
func (c *SplitClient) Treatment(key interface{}, feature string, attributes map[string]interface{}) (ret string) {
	// Set up a guard deferred function to recover if the SDK starts panicking
	defer func() {
		if r := recover(); r != nil {
			// At this point we'll only trust that the logger isn't panicking trust
			// that the logger isn't panicking
			c.logger.Error(
				"SDK is panicking with the following error", r, "\n",
				string(debug.Stack()), "\n",
				"Returning CONTROL", "\n")
			ret = evaluator.Control
		}
	}()

	if c.IsDestroyed() {
		return evaluator.Control
	}

	matchingKey, bucketingKey, err := c.validator.ValidateTreatmentKey(key)
	if err != nil {
		c.logger.Error(err.Error())
		return evaluator.Control
	}

	var evaluationResult *evaluator.Result
	var impressionBucketingKey = ""
	if bucketingKey != nil {
		evaluationResult = c.evaluator.Evaluate(matchingKey, bucketingKey, feature, attributes)
		impressionBucketingKey = *bucketingKey
	} else {
		evaluationResult = c.evaluator.Evaluate(matchingKey, &matchingKey, feature, attributes)
	}

	// Store impression
	if c.impressions != nil {
		var label string
		if c.cfg.LabelsEnabled {
			label = evaluationResult.Label
		}
		c.impressions.Put(feature, &dtos.ImpressionDTO{
			BucketingKey: impressionBucketingKey,
			ChangeNumber: evaluationResult.SplitChangeNumber,
			KeyName:      matchingKey,
			Label:        label,
			Treatment:    evaluationResult.Treatment,
			Time:         time.Now().Unix() * 1000, // Convert standard timestamp to java's ms timestamps
		})
	} else {
		c.logger.Warning("No impression storage set in client. Not sending impressions!")
	}

	// Store latency
	bucket := metrics.Bucket(evaluationResult.EvaluationTimeNs)
	c.metrics.IncLatency("sdk.getTreatment", bucket)

	return evaluationResult.Treatment
}

// Treatments evaluates multiple featers for a single user and set of attributes at once
func (c *SplitClient) Treatments(key interface{}, features []string, attributes map[string]interface{}) map[string]string {
	treatments := make(map[string]string)

	if c.IsDestroyed() {
		return treatments
	}

	for _, feature := range features {
		if strings.TrimSpace(feature) != "" {
			treatments[feature] = c.Treatment(key, feature, attributes)
		}
	}
	return treatments
}

// IsDestroyed returns true if tbe client has been destroyed
func (c *SplitClient) IsDestroyed() bool {
	c.destroyed.mutex.RLock()
	defer c.destroyed.mutex.RUnlock()
	return c.destroyed.status
}

// Destroy stops all async tasks and clears all storages
func (c *SplitClient) Destroy() {
	c.destroyed.mutex.Lock()
	defer c.destroyed.mutex.Unlock()
	c.destroyed.status = true

	if c.cfg.OperationMode == "redis-consumer" || c.cfg.OperationMode == "localhost" {
		return
	}

	// Stop all tasks
	if c.sync.splitSync != nil {
		c.sync.splitSync.Stop()
	}
	if c.sync.segmentSync != nil {
		c.sync.segmentSync.Stop()
	}

	if c.sync.impressionSync != nil {
		c.sync.impressionSync.Stop()
	}
	if c.sync.gaugeSync != nil {
		c.sync.gaugeSync.Stop()
	}
	if c.sync.countersSync != nil {
		c.sync.countersSync.Stop()
	}
	if c.sync.latenciesSync != nil {
		c.sync.latenciesSync.Stop()
	}
}

// Track an event and its custom value
func (c *SplitClient) Track(key string, trafficType string, eventType string, value interface{}) (ret error) {

	defer func() {
		if r := recover(); r != nil {
			// At this point we'll only trust that the logger isn't panicking
			c.logger.Error(
				"SDK is panicking with the following error", r, "\n",
				string(debug.Stack()), "\n",
			)
			ret = errors.New("Track is panicking. Please check logs")
		}
		return
	}()

	key, trafficType, eventType, value, iErr := ValidateTrackInputs(key, trafficType, eventType, value)
	if iErr != nil {
		c.logger.Error(iErr.Error())
		return iErr
	}

	err := c.events.Push(dtos.EventDTO{
		Key:             key,
		TrafficTypeName: trafficType,
		EventTypeID:     eventType,
		Value:           value,
		Timestamp:       time.Now().UnixNano() / 1000000,
	})

	if err != nil {
		c.logger.Error("Error tracking event", err.Error())
		return err
	}

	return nil
}
