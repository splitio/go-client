package dtos

import (
	"encoding/json"
)

// ImpressionDTO struct to map an impression
type ImpressionDTO struct {
	KeyName      string `json:"keyName"`
	Treatment    string `json:"treatment"`
	Time         int64  `json:"time"`
	ChangeNumber int64  `json:"changeNumber"`
	Label        string `json:"label"`
	BucketingKey string `json:"bucketingKey,omitempty"`
}

// MarshalBinary exports ImpressionDTO to JSON string
func (s ImpressionDTO) MarshalBinary() (data []byte, err error) {
	return json.Marshal(s)
}

// ImpressionsDTO struct mapping impressions to post
type ImpressionsDTO struct {
	TestName       string          `json:"testName"`
	KeyImpressions []ImpressionDTO `json:"keyImpressions"`
}
