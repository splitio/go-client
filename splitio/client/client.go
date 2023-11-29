package client

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/splitio/go-client/v6/splitio/conf"
	impressionlistener "github.com/splitio/go-client/v6/splitio/impressionListener"

	"github.com/splitio/go-split-commons/v5/dtos"
	"github.com/splitio/go-split-commons/v5/engine/evaluator"
	"github.com/splitio/go-split-commons/v5/engine/evaluator/impressionlabels"
	"github.com/splitio/go-split-commons/v5/flagsets"
	"github.com/splitio/go-split-commons/v5/provisional"
	"github.com/splitio/go-split-commons/v5/storage"
	"github.com/splitio/go-split-commons/v5/telemetry"
	"github.com/splitio/go-toolkit/v5/logging"
)

const (
	treatment                      = "Treatment"
	treatments                     = "Treatments"
	treatmentsByFlagSet            = "TreatmentsByFlagSet"
	treatmentsByFlagSets           = "TreatmentsByFlahSets"
	treatmentWithConfig            = "TreatmentWithConfig"
	treatmentsWithConfig           = "TreatmentsWithConfig"
	treatmentsWithConfigByFlagSet  = "TreatmentsWithConfigByFlagSet"
	treatmentsWithConfigByFlagSets = "TrearmentsWithConfigByFlagSets"
)

// SplitClient is the entry-point of the split SDK.
type SplitClient struct {
	logger              logging.LoggerInterface
	evaluator           evaluator.Interface
	impressions         storage.ImpressionStorageProducer
	events              storage.EventStorageProducer
	validator           inputValidation
	factory             *SplitFactory
	impressionListener  *impressionlistener.WrapperImpressionListener
	impressionManager   provisional.ImpressionManager
	initTelemetry       storage.TelemetryConfigProducer
	evaluationTelemetry storage.TelemetryEvaluationProducer
	runtimeTelemetry    storage.TelemetryRuntimeProducer
	flagSetsFilter      flagsets.FlagSetFilter
}

// TreatmentResult struct that includes the Treatment evaluation with the corresponding Config
type TreatmentResult struct {
	Treatment string  `json:"treatment"`
	Config    *string `json:"config"`
}

// getEvaluationResult calls evaluation for one particular feature flag
func (c *SplitClient) getEvaluationResult(matchingKey string, bucketingKey *string, featureFlag string, attributes map[string]interface{}, operation string) *evaluator.Result {
	if c.isReady() {
		return c.evaluator.EvaluateFeature(matchingKey, bucketingKey, featureFlag, attributes)
	}

	c.logger.Warning(fmt.Sprintf("%s: the SDK is not ready, results may be incorrect for feature flag %s. Make sure to wait for SDK readiness before using this method", operation, featureFlag))
	c.initTelemetry.RecordNonReadyUsage()
	return &evaluator.Result{
		Treatment: evaluator.Control,
		Label:     impressionlabels.ClientNotReady,
		Config:    nil,
	}
}

// getEvaluationsResult calls evaluation for multiple treatments at once
func (c *SplitClient) getEvaluationsResult(matchingKey string, bucketingKey *string, featureFlags []string, attributes map[string]interface{}, operation string) evaluator.Results {
	if c.isReady() {
		return c.evaluator.EvaluateFeatures(matchingKey, bucketingKey, featureFlags, attributes)
	}
	featureFlagsToPrint := strings.Join(featureFlags, ", ")
	c.logger.Warning(fmt.Sprintf("%s: the SDK is not ready, results may be incorrect for feature flags %s. Make sure to wait for SDK readiness before using this method", operation, featureFlagsToPrint))
	c.initTelemetry.RecordNonReadyUsage()
	result := evaluator.Results{
		EvaluationTime: 0,
		Evaluations:    make(map[string]evaluator.Result),
	}
	for _, featureFlag := range featureFlags {
		result.Evaluations[featureFlag] = evaluator.Result{
			Treatment: evaluator.Control,
			Label:     impressionlabels.ClientNotReady,
			Config:    nil,
		}
	}
	return result
}

// createImpression creates impression to be stored and used by listener
func (c *SplitClient) createImpression(featureFlag string, bucketingKey *string, evaluationLabel string, matchingKey string, treatment string, changeNumber int64) dtos.Impression {
	var label string
	if c.factory.cfg.LabelsEnabled {
		label = evaluationLabel
	}

	impressionBucketingKey := ""
	if bucketingKey != nil {
		impressionBucketingKey = *bucketingKey
	}

	return dtos.Impression{
		FeatureName:  featureFlag,
		BucketingKey: impressionBucketingKey,
		ChangeNumber: changeNumber,
		KeyName:      matchingKey,
		Label:        label,
		Treatment:    treatment,
		Time:         time.Now().UTC().UnixNano() / int64(time.Millisecond), // Convert standard timestamp to java's ms timestamps
	}
}

// storeData stores impression, runs listener and stores metrics
func (c *SplitClient) storeData(impressions []dtos.Impression, attributes map[string]interface{}, metricsLabel string, evaluationTime time.Duration) {
	// Store impression
	if c.impressions != nil {
		forLog, forListener := c.impressionManager.ProcessImpressions(impressions)
		c.impressions.LogImpressions(forLog)

		// Custom Impression Listener
		if c.impressionListener != nil {
			c.impressionListener.SendDataToClient(forListener, attributes)
		}
	} else {
		c.logger.Warning("No impression storage set in client. Not sending impressions!")
	}

	// Store latency
	c.evaluationTelemetry.RecordLatency(metricsLabel, evaluationTime)
}

// doTreatmentCall retrieves treatments of an specific feature flag with configurations object if it is present for a certain key and set of attributes
func (c *SplitClient) doTreatmentCall(key interface{}, featureFlag string, attributes map[string]interface{}, operation string, metricsLabel string) (t TreatmentResult) {
	controlTreatment := TreatmentResult{
		Treatment: evaluator.Control,
		Config:    nil,
	}

	// Set up a guard deferred function to recover if the SDK starts panicking
	defer func() {
		if r := recover(); r != nil {
			// At this point we'll only trust that the logger isn't panicking trust
			// that the logger isn't panicking
			c.evaluationTelemetry.RecordException(metricsLabel)
			c.logger.Error(
				"SDK is panicking with the following error", r, "\n",
				string(debug.Stack()), "\n",
				"Returning CONTROL", "\n")
			t = controlTreatment
		}
	}()

	if c.isDestroyed() {
		c.logger.Error("Client has already been destroyed - no calls possible")
		return controlTreatment
	}

	matchingKey, bucketingKey, err := c.validator.ValidateTreatmentKey(key, operation)
	if err != nil {
		c.logger.Error(err.Error())
		return controlTreatment
	}

	featureFlag, err = c.validator.ValidateFeatureName(featureFlag, operation)
	if err != nil {
		c.logger.Error(err.Error())
		return controlTreatment
	}

	evaluationResult := c.getEvaluationResult(matchingKey, bucketingKey, featureFlag, attributes, operation)

	if !c.validator.IsSplitFound(evaluationResult.Label, featureFlag, operation) {
		return controlTreatment
	}

	c.storeData(
		[]dtos.Impression{c.createImpression(featureFlag, bucketingKey, evaluationResult.Label, matchingKey, evaluationResult.Treatment, evaluationResult.SplitChangeNumber)},
		attributes,
		metricsLabel,
		evaluationResult.EvaluationTime,
	)

	return TreatmentResult{
		Treatment: evaluationResult.Treatment,
		Config:    evaluationResult.Config,
	}
}

// Treatment implements the main functionality of split. Retrieve treatments of a specific feature flag
// for a certain key and set of attributes
func (c *SplitClient) Treatment(key interface{}, featureFlagName string, attributes map[string]interface{}) string {
	return c.doTreatmentCall(key, featureFlagName, attributes, treatment, telemetry.Treatment).Treatment
}

// TreatmentWithConfig implements the main functionality of split. Retrieves the treatment of a specific feature flag
// with the corresponding configuration if it is present
func (c *SplitClient) TreatmentWithConfig(key interface{}, featureFlagName string, attributes map[string]interface{}) TreatmentResult {
	return c.doTreatmentCall(key, featureFlagName, attributes, treatmentWithConfig, telemetry.TreatmentWithConfig)
}

// Generates control treatments
func (c *SplitClient) generateControlTreatments(featureFlagNames []string, operation string) map[string]TreatmentResult {
	treatments := make(map[string]TreatmentResult)
	filtered, err := c.validator.ValidateFeatureNames(featureFlagNames, operation)
	if err != nil {
		return treatments
	}
	for _, featureFlag := range filtered {
		treatments[featureFlag] = TreatmentResult{
			Treatment: evaluator.Control,
			Config:    nil,
		}
	}
	return treatments
}

func (c *SplitClient) processResult(result evaluator.Results, operation string, bucketingKey *string, matchingKey string, attributes map[string]interface{}, metricsLabel string) (t map[string]TreatmentResult) {
	var bulkImpressions []dtos.Impression
	treatments := make(map[string]TreatmentResult)
	for feature, evaluation := range result.Evaluations {
		if !c.validator.IsSplitFound(evaluation.Label, feature, operation) {
			treatments[feature] = TreatmentResult{
				Treatment: evaluator.Control,
				Config:    nil,
			}
		} else {
			bulkImpressions = append(bulkImpressions, c.createImpression(feature, bucketingKey, evaluation.Label, matchingKey, evaluation.Treatment, evaluation.SplitChangeNumber))

			treatments[feature] = TreatmentResult{
				Treatment: evaluation.Treatment,
				Config:    evaluation.Config,
			}
		}
	}
	c.storeData(bulkImpressions, attributes, metricsLabel, result.EvaluationTime)
	return treatments
}

// doTreatmentsCall retrieves treatments of an specific array of feature flag names with configurations object if it is present for a certain key and set of attributes
func (c *SplitClient) doTreatmentsCall(key interface{}, featureFlagNames []string, attributes map[string]interface{}, operation string, metricsLabel string) (t map[string]TreatmentResult) {
	treatments := make(map[string]TreatmentResult)

	// Set up a guard deferred function to recover if the SDK starts panicking
	defer func() {
		if r := recover(); r != nil {
			// At this point we'll only trust that the logger isn't panicking trust
			// that the logger isn't panicking
			c.evaluationTelemetry.RecordException(metricsLabel)
			c.logger.Error(
				"SDK is panicking with the following error", r, "\n",
				string(debug.Stack()), "\n")
			t = c.generateControlTreatments(featureFlagNames, operation)
		}
	}()

	if c.isDestroyed() {
		c.logger.Error("Client has already been destroyed - no calls possible")
		return c.generateControlTreatments(featureFlagNames, operation)
	}

	matchingKey, bucketingKey, err := c.validator.ValidateTreatmentKey(key, operation)
	if err != nil {
		c.logger.Error(err.Error())
		return c.generateControlTreatments(featureFlagNames, operation)
	}

	filteredFeatures, err := c.validator.ValidateFeatureNames(featureFlagNames, operation)
	if err != nil {
		c.logger.Error(err.Error())
		return map[string]TreatmentResult{}
	}

	evaluationsResult := c.getEvaluationsResult(matchingKey, bucketingKey, filteredFeatures, attributes, operation)

	treatments = c.processResult(evaluationsResult, operation, bucketingKey, matchingKey, attributes, metricsLabel)

	return treatments
}

// doTreatmentsCallByFlagSets retrieves treatments of a specific array of feature flag names, that belong to flag sets, with configurations object if it is present for a certain key and set of attributes
func (c *SplitClient) doTreatmentsCallByFlagSets(key interface{}, flagSets []string, attributes map[string]interface{}, operation string, metricsLabel string) (t map[string]TreatmentResult) {
	treatments := make(map[string]TreatmentResult)

	// Set up a guard deferred function to recover if the SDK starts panicking
	defer func() {
		if r := recover(); r != nil {
			// At this point we'll only trust that the logger isn't panicking trust
			// that the logger isn't panicking
			c.evaluationTelemetry.RecordException(metricsLabel)
			c.logger.Error(
				"SDK is panicking with the following error", r, "\n",
				string(debug.Stack()), "\n")
			t = treatments
		}
	}()

	if c.isDestroyed() {
		return treatments
	}

	matchingKey, bucketingKey, err := c.validator.ValidateTreatmentKey(key, operation)
	if err != nil {
		c.logger.Error(err.Error())
		return treatments
	}

	if c.isReady() {
		evaluationsResult := c.evaluator.EvaluateFeatureByFlagSets(matchingKey, bucketingKey, flagSets, attributes)
		treatments = c.processResult(evaluationsResult, operation, bucketingKey, matchingKey, attributes, metricsLabel)
	}
	return treatments
}

// Treatments evaluates multiple feature flag names for a single user and set of attributes at once
func (c *SplitClient) Treatments(key interface{}, featureFlagNames []string, attributes map[string]interface{}) map[string]string {
	treatmentsResult := map[string]string{}
	result := c.doTreatmentsCall(key, featureFlagNames, attributes, treatments, telemetry.Treatments)
	for feature, treatmentResult := range result {
		treatmentsResult[feature] = treatmentResult.Treatment
	}
	return treatmentsResult
}

func (c *SplitClient) validateSets(flagSets []string) []string {
	if len(flagSets) == 0 {
		c.logger.Warning("sets must be a non-empty array")
		return nil
	}
	flagSets, errs := flagsets.SanitizeMany(flagSets)
	if len(errs) != 0 {
		for _, err := range errs {
			if errType, ok := err.(*dtos.FlagSetValidatonError); ok {
				c.logger.Warning(errType.Message)
			}
		}
	}
	flagSets = c.filterSetsAreInConfig(flagSets)
	if len(flagSets) == 0 {
		return nil
	}
	return flagSets
}

// Treatments evaluate multiple feature flag names belonging to a flag set for a single user and a set of attributes at once
func (c *SplitClient) TreatmentsByFlagSet(key interface{}, flagSet string, attributes map[string]interface{}) map[string]string {
	treatmentsResult := map[string]string{}
	sets := c.validateSets([]string{flagSet})
	if sets == nil {
		return treatmentsResult
	}
	result := c.doTreatmentsCallByFlagSets(key, sets, attributes, treatmentsByFlagSet, telemetry.TreatmentsByFlagSet)
	for feature, treatmentResult := range result {
		treatmentsResult[feature] = treatmentResult.Treatment
	}
	return treatmentsResult
}

// Treatments evaluate multiple feature flag names belonging to flag sets for a single user and a set of attributes at once
func (c *SplitClient) TreatmentsByFlagSets(key interface{}, flagSets []string, attributes map[string]interface{}) map[string]string {
	treatmentsResult := map[string]string{}
	flagSets = c.validateSets(flagSets)
	if flagSets == nil {
		return treatmentsResult
	}
	result := c.doTreatmentsCallByFlagSets(key, flagSets, attributes, treatmentsByFlagSets, telemetry.TreatmentsByFlagSets)
	for feature, treatmentResult := range result {
		treatmentsResult[feature] = treatmentResult.Treatment
	}
	return treatmentsResult
}

func (c *SplitClient) filterSetsAreInConfig(flagSets []string) []string {
	toReturn := []string{}
	for _, flagSet := range flagSets {
		if !c.flagSetsFilter.IsPresent(flagSet) {
			c.logger.Warning(fmt.Sprintf("you passed %s which is not part of the configured FlagSetsFilter, ignoring Flag Set.", flagSet))
			continue
		}
		toReturn = append(toReturn, flagSet)
	}
	return toReturn
}

// TreatmentsWithConfig evaluates multiple feature flag names for a single user and set of attributes at once and returns configurations
func (c *SplitClient) TreatmentsWithConfig(key interface{}, featureFlagNames []string, attributes map[string]interface{}) map[string]TreatmentResult {
	return c.doTreatmentsCall(key, featureFlagNames, attributes, treatmentsWithConfig, telemetry.TreatmentsWithConfig)
}

// TreatmentsWithConfigByFlagSet evaluates multiple feature flag names belonging to a flag set for a single user and set of attributes at once and returns configurations
func (c *SplitClient) TreatmentsWithConfigByFlagSet(key interface{}, flagSet string, attributes map[string]interface{}) map[string]TreatmentResult {
	treatmentsResult := make(map[string]TreatmentResult)
	sets := c.validateSets([]string{flagSet})
	if sets == nil {
		return treatmentsResult
	}
	return c.doTreatmentsCallByFlagSets(key, sets, attributes, treatmentsWithConfigByFlagSet, telemetry.TreatmentsByFlagSets)
}

// TreatmentsWithConfigByFlagSet evaluates multiple feature flag names belonging to a flag sets for a single user and set of attributes at once and returns configurations
func (c *SplitClient) TreatmentsWithConfigByFlagSets(key interface{}, flagSets []string, attributes map[string]interface{}) map[string]TreatmentResult {
	treatmentsResult := make(map[string]TreatmentResult)
	flagSets = c.validateSets(flagSets)
	if flagSets == nil {
		return treatmentsResult
	}
	return c.doTreatmentsCallByFlagSets(key, flagSets, attributes, treatmentsWithConfigByFlagSets, telemetry.TreatmentsByFlagSets)
}

// isDestroyed returns true if the client has been destroyed
func (c *SplitClient) isDestroyed() bool {
	return c.factory.IsDestroyed()
}

// isReady returns true if the client is ready
func (c *SplitClient) isReady() bool {
	return c.factory.IsReady()
}

// Destroy the client and the underlying factory.
func (c *SplitClient) Destroy() {
	if !c.isDestroyed() {
		c.factory.Destroy()
	}
}

// Track an event and its custom value
func (c *SplitClient) Track(key string, trafficType string, eventType string, value interface{}, properties map[string]interface{}) (ret error) {
	defer func() {
		if r := recover(); r != nil {
			// At this point we'll only trust that the logger isn't panicking
			c.evaluationTelemetry.RecordException(telemetry.Track)
			c.logger.Error(
				"SDK is panicking with the following error", r, "\n",
				string(debug.Stack()), "\n",
			)
			ret = errors.New("Track is panicking. Please check logs")
		}
	}()

	if c.isDestroyed() {
		c.logger.Error("Client has already been destroyed - no calls possible")
		return errors.New("Client has already been destroyed - no calls possible")
	}

	if !c.isReady() {
		c.logger.Warning("Track: the SDK is not ready, results may be incorrect. Make sure to wait for SDK readiness before using this method")
		c.initTelemetry.RecordNonReadyUsage()
	}

	key, trafficType, eventType, value, err := c.validator.ValidateTrackInputs(
		key,
		trafficType,
		eventType,
		value,
		c.isReady() && c.factory.apikey != conf.Localhost,
	)
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}

	properties, size, err := c.validator.validateTrackProperties(properties)
	if err != nil {
		return err
	}

	err = c.events.Push(dtos.EventDTO{
		Key:             key,
		TrafficTypeName: trafficType,
		EventTypeID:     eventType,
		Value:           value,
		Timestamp:       time.Now().UTC().UnixNano() / int64(time.Millisecond), // Convert standard timestamp to java's ms timestamps
		Properties:      properties,
	}, size)

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
