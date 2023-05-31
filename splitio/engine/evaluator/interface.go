package evaluator

// Interface should be implemented by concrete treatment evaluator structs
type Interface interface {
	EvaluateFeature(key string, bucketingKey *string, featureFlag string, attributes map[string]interface{}) *Result
	EvaluateFeatures(key string, bucketingKey *string, featureFlags []string, attributes map[string]interface{}) Results
}
