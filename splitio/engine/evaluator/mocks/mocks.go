package mocks

import "github.com/splitio/go-client/splitio/engine/evaluator"

// MockEvaluator mock evaluator
type MockEvaluator struct {
	EvaluateFeatureCall  func(key string, bucketingKey *string, feature string, attributes map[string]interface{}) *evaluator.Result
	EvaluateFeaturesCall func(key string, bucketingKey *string, features []string, attributes map[string]interface{}) evaluator.Results
}

// EvaluateFeature mock
func (m MockEvaluator) EvaluateFeature(key string, bucketingKey *string, feature string, attributes map[string]interface{}) *evaluator.Result {
	return m.EvaluateFeatureCall(key, bucketingKey, feature, attributes)
}

// EvaluateFeatures mock
func (m MockEvaluator) EvaluateFeatures(key string, bucketingKey *string, features []string, attributes map[string]interface{}) evaluator.Results {
	return m.EvaluateFeaturesCall(key, bucketingKey, features, attributes)
}
