package evaluator

// Interface should be implemented by concrete treatment evaluator structs
type Interface interface {
	Evaluate(key string, bucketingKey *string, feature string, attributes map[string]interface{}) *Result
}
