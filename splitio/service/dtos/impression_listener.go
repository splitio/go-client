package dtos

// ImpressionListenerDTO struct to map entire data for listener
type ImpressionListenerDTO struct {
	Impression         ImpressionDataDTO
	Attributes         map[string]interface{}
	InstanceID         string
	SDKLanguageVersion string
}

// ImpressionDataDTO impression data for listener
type ImpressionDataDTO struct {
	Feature      string
	KeyName      string
	Treatment    string
	Time         int64
	ChangeNumber int64
	Label        string
	BucketingKey string
}
