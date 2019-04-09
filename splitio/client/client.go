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
	Treatment string
	Config    *string
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
			BucketingKey: impressionBucketingKey,
			ChangeNumber: evaluationResult.SplitChangeNumber,
			KeyName:      matchingKey,
			Label:        label,
			Treatment:    evaluationResult.Treatment,
			Time:         time.Now().Unix() * 1000, // Convert standard timestamp to java's ms timestamps
		}

		keyImpressions := []dtos.ImpressionDTO{impression}
		toStore := []dtos.ImpressionsDTO{dtos.ImpressionsDTO{
			TestName:       feature,
			KeyImpressions: keyImpressions,
		}}

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
func (c *SplitClient) Treatment(key interface{}, feature string, attributes map[string]interface{}) (ret string) {
	return c.doTreatmentCall(key, feature, attributes, "Treatment", "sdk.getTreatment").Treatment
}

// TreatmentWithConfig implements the main functionality of split. Retrieves the treatment of a specific feature with
// the corresponding configuration if it is present
func (c *SplitClient) TreatmentWithConfig(key interface{}, feature string, attributes map[string]interface{}) TreatmentResult {
	return c.doTreatmentCall(key, feature, attributes, "TreatmentWithConfig", "sdk.getTreatmentWithConfig")
}

// Treatments evaluates multiple featers for a single user and set of attributes at once
func (c *SplitClient) Treatments(key interface{}, features []string, attributes map[string]interface{}) map[string]string {
	treatments := make(map[string]string)

	if c.IsDestroyed() {
		c.logger.Error("Client has already been destroyed - no calls possible")
		return c.validator.GenerateControlTreatments(features)
	}

	before := time.Now()
	var bulkImpressions []dtos.ImpressionsDTO

	matchingKey, bucketingKey, err := c.validator.ValidateTreatmentKey(key, "Treatments")
	if err != nil {
		c.logger.Error(err.Error())
		return c.validator.GenerateControlTreatments(features)
	}

	filteredFeatures, err := c.validator.ValidateFeatureNames(features)
	if err != nil {
		c.logger.Error(err.Error())
		return map[string]string{}
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
				BucketingKey: impressionBucketingKey,
				ChangeNumber: evaluationResult.SplitChangeNumber,
				KeyName:      matchingKey,
				Label:        label,
				Treatment:    evaluationResult.Treatment,
				Time:         time.Now().Unix() * 1000, // Convert standard timestamp to java's ms timestamps
			}
			keyImpressions := []dtos.ImpressionDTO{impression}
			bulkImpressions = append(bulkImpressions, dtos.ImpressionsDTO{
				TestName:       feature,
				KeyImpressions: keyImpressions,
			})
		} else {
			c.logger.Warning("No impression storage set in client. Not sending impressions!")
		}

		treatments[feature] = evaluationResult.Treatment
	}

	// Custom Impression Listener
	if c.impressionListener != nil {
		for _, dataToSend := range bulkImpressions {
			c.impressionListener.SendDataToClient(dataToSend, attributes, c.metadata)
		}
	}

	// Store latency
	bucket := metrics.Bucket(time.Now().Sub(before).Nanoseconds())
	c.metrics.IncLatency("sdk.getTreatments", bucket)

	c.impressions.LogImpressions(bulkImpressions)

	return treatments
}

// IsDestroyed returns true if tbe client has been destroyed
func (c *SplitClient) IsDestroyed() bool {
	return c.factory.IsDestroyed()

}

// Destroy stops all async tasks and clears all storages
func (c *SplitClient) Destroy() {
	c.factory.Destroy()

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
