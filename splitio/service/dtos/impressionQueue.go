package dtos

// ImpressionQueueDTO struct to map an impression
type ImpressionQueueDTO struct {
	KeyName      string `json:"k"`
	BucketingKey string `json:"b"`
	FeatureName  string `json:"f"`
	Treatment    string `json:"t"`
	Label        string `json:"r"`
	ChangeNumber int64  `json:"c"`
	Time         int64  `json:"m"`
}

// ImpressionsQueueDTO struct mapping impressions to post
type ImpressionsQueueDTO struct {
	Metadata   QueueStoredMachineMetadataDTO `json:"m"`
	Impression ImpressionQueueDTO            `json:"i"`
}
