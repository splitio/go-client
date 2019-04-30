package dtos

// ImpressionDTO struct to map an impression
type ImpressionDTO struct {
	KeyName      string `json:"k"`
	BucketingKey string `json:"b"`
	FeatureName  string `json:"f"`
	Treatment    string `json:"t"`
	Label        string `json:"r"`
	ChangeNumber int64  `json:"c"`
	Time         int64  `json:"m"`
}

// ImpressionsDTO struct mapping impressions
type ImpressionsDTO struct {
	Metadata   QueueStoredMachineMetadataDTO `json:"m"`
	Impression ImpressionDTO                 `json:"i"`
}
