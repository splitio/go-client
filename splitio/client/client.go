package client

import (
	"errors"
	"runtime/debug"
	"time"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-client/splitio/engine/evaluator"
	"github.com/splitio/go-client/splitio/impressionListener"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-client/splitio/util/metrics"

	"github.com/splitio/go-toolkit/asynctask"
	"github.com/splitio/go-toolkit/logging"
)

// SplitClient is the entry-point of the split SDK.
type SplitClient struct {
	apikey             string
	cfg                *conf.SplitSdkConfig
	logger             logging.LoggerInterface
	loggerConfig       logging.LoggerOptions
	evaluator          evaluator.Interface
	sync               *sdkSync
	impressions        storage.ImpressionStorageProducer
	metrics            storage.MetricsStorageProducer
	events             storage.EventStorageProducer
	validator          inputValidation
	factory            *SplitFactory
	impressionListener *impressionlistener.WrapperImpressionListener
	metadata           dtos.QueueStoredMachineMetadataDTO
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

// TreatmentResult struct that includes the Treatment evaluation with the corresponding Config
type TreatmentResult struct {
	Treatment string  `json:"treatment"`
	Config    *string `json:"config"`
}

// doTreatmentCall retrieves treatments of an specific feature with configurations object if it is present
// for a certain key and set of attributes
func (c *SplitClient) doTreatmentCall(key interface{}, feature string, attributes map[string]interface{}, operation string, metricsLabel string) (t TreatmentResult) {
	controlTreatment := TreatmentResult{
		Treatment: evaluator.Control,
		Config:    nil,
	}

	// Set up a guard deferred function to recover if the SDK starts panicking
	defer func() {
		if r := recover(); r != nil {
			// At this point we'll only trust that the logger isn't panicking trust
			// that the logger isn't panicking
			c.logger.Error(
				"SDK is panicking with the following error", r, "\n",
				string(debug.Stack()), "\n",
				"Returning CONTROL", "\n")
			t = controlTreatment
		}
	}()

	if c.IsDestroyed() {
		c.logger.Error("Client has already been destroyed - no calls possible")
		return controlTreatment
	}

	matchingKey, bucketingKey, err := c.validator.ValidateTreatmentKey(key, operation)
	if err != nil {
		c.logger.Error(err.Error())
		return controlTreatment
	}

	feature, err = c.validator.ValidateFeatureName(feature, operation)
	if err != nil {
		c.logger.Error(err.Error())
		return controlTreatment
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

		var impression = dtos.ImpressionDTO{
			FeatureName:  feature,
			BucketingKey: impressionBucketingKey,
			ChangeNumber: evaluationResult.SplitChangeNumber,
			KeyName:      matchingKey,
			Label:        label,
			Treatment:    evaluationResult.Treatment,
			Time:         time.Now().Unix() * 1000, // Convert standard timestamp to java's ms timestamps
		}

		toStore := []dtos.ImpressionDTO{impression}
		c.impressions.LogImpressions(toStore)

		// Custom Impression Listener
		if c.impressionListener != nil {
			for _, dataToSend := range toStore {
				c.impressionListener.SendDataToClient(dataToSend, attributes, c.metadata)
			}
		}
	} else {
		c.logger.Warning("No impression storage set in client. Not sending impressions!")
	}

	// Store latency
	bucket := metrics.Bucket(evaluationResult.EvaluationTimeNs)
	c.metrics.IncLatency(metricsLabel, bucket)

	return TreatmentResult{
		Treatment: evaluationResult.Treatment,
		Config:    evaluationResult.Config,
	}
}

// Treatment implements the main functionality of split. Retrieve treatments of a specific feature
// for a certain key and set of attributes
func (c *SplitClient) Treatment(key interface{}, feature string, attributes map[string]interface{}) string {
	return c.doTreatmentCall(key, feature, attributes, "Treatment", "sdk.getTreatment").Treatment
}

// TreatmentWithConfig implements the main functionality of split. Retrieves the treatment of a specific feature with
// the corresponding configuration if it is present
func (c *SplitClient) TreatmentWithConfig(key interface{}, feature string, attributes map[string]interface{}) TreatmentResult {
	return c.doTreatmentCall(key, feature, attributes, "TreatmentWithConfig", "sdk.getTreatmentWithConfig")
}

// Generates control treatments
func (c *SplitClient) generateControlTreatments(features []string, operation string) map[string]TreatmentResult {
	treatments := make(map[string]TreatmentResult)
	filtered, err := c.validator.ValidateFeatureNames(features, operation)
	if err != nil {
		return treatments
	}
	for _, feature := range filtered {
		treatments[feature] = TreatmentResult{
			Treatment: evaluator.Control,
			Config:    nil,
		}
	}
	return treatments
}

// doTreatmentsCall retrieves treatments of an specific array of features with configurations object if it is present
// for a certain key and set of attributes
func (c *SplitClient) doTreatmentsCall(key interface{}, features []string, attributes map[string]interface{}, operation string, metricsLabel string) (t map[string]TreatmentResult) {
	treatments := make(map[string]TreatmentResult)

	// Set up a guard deferred function to recover if the SDK starts panicking
	defer func() {
		if r := recover(); r != nil {
			// At this point we'll only trust that the logger isn't panicking trust
			// that the logger isn't panicking
			c.logger.Error(
				"SDK is panicking with the following error", r, "\n",
				string(debug.Stack()), "\n")
			t = treatments
		}
	}()

	if c.IsDestroyed() {
		c.logger.Error("Client has already been destroyed - no calls possible")
		return c.generateControlTreatments(features, operation)
	}

	before := time.Now()
	var bulkImpressions []dtos.ImpressionDTO

	matchingKey, bucketingKey, err := c.validator.ValidateTreatmentKey(key, operation)
	if err != nil {
		c.logger.Error(err.Error())
		return c.generateControlTreatments(features, operation)
	}

	filteredFeatures, err := c.validator.ValidateFeatureNames(features, operation)
	if err != nil {
		c.logger.Error(err.Error())
		return map[string]TreatmentResult{}
	}

	for _, feature := range filteredFeatures {
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
			var impression = dtos.ImpressionDTO{
				FeatureName:  feature,
				BucketingKey: impressionBucketingKey,
				ChangeNumber: evaluationResult.SplitChangeNumber,
				KeyName:      matchingKey,
				Label:        label,
				Treatment:    evaluationResult.Treatment,
				Time:         time.Now().Unix() * 1000, // Convert standard timestamp to java's ms timestamps
			}
			bulkImpressions = append(bulkImpressions, impression)
		} else {
			c.logger.Warning("No impression storage set in client. Not sending impressions!")
		}

		treatments[feature] = TreatmentResult{
			Treatment: evaluationResult.Treatment,
			Config:    evaluationResult.Config,
		}
	}

	// Custom Impression Listener
	if c.impressionListener != nil {
		for _, dataToSend := range bulkImpressions {
			c.impressionListener.SendDataToClient(dataToSend, attributes, c.metadata)
		}
	}

	// Store latency
	bucket := metrics.Bucket(time.Now().Sub(before).Nanoseconds())
	c.metrics.IncLatency(metricsLabel, bucket)

	c.impressions.LogImpressions(bulkImpressions)

	return treatments
}

// Treatments evaluates multiple featers for a single user and set of attributes at once
func (c *SplitClient) Treatments(key interface{}, features []string, attributes map[string]interface{}) map[string]string {
	treatments := map[string]string{}
	result := c.doTreatmentsCall(key, features, attributes, "Treatments", "sdk.getTreatments")
	for feature, treatmentResult := range result {
		treatments[feature] = treatmentResult.Treatment
	}
	return treatments
}

// TreatmentsWithConfig evaluates multiple featers for a single user and set of attributes at once and returns configurations
func (c *SplitClient) TreatmentsWithConfig(key interface{}, features []string, attributes map[string]interface{}) map[string]TreatmentResult {
	return c.doTreatmentsCall(key, features, attributes, "TreatmentsWithConfig", "sdk.getTreatmentsWithConfig")
}

// IsDestroyed returns true if tbe client has been destroyed
func (c *SplitClient) IsDestroyed() bool {
	return c.factory.IsDestroyed()

}

// Destroy stops all async tasks and clears all storages
func (c *SplitClient) Destroy() {
	c.factory.Destroy()

	if c.cfg.OperationMode == "redis-consumer" {
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

	if c.IsDestroyed() {
		c.logger.Error("Client has already been destroyed - no calls possible")
		return errors.New("Client has already been destroyed - no calls possible")
	}

	key, trafficType, eventType, value, err := c.validator.ValidateTrackInputs(key, trafficType, eventType, value)
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}

	err = c.events.Push(dtos.EventDTO{
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

// BlockUntilReady Calls BlockUntilReady on factory to block client on readiness
func (c *SplitClient) BlockUntilReady(timer int) error {
	return c.factory.BlockUntilReady(timer)
}
