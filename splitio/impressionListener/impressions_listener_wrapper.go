package impressionlistener

import (
	"github.com/splitio/go-client/splitio/service/dtos"
)

// ILObject struct to map entire data for listener
type ILObject struct {
	Impression         ImpressionData
	Attributes         map[string]interface{}
	InstanceID         string
	SDKLanguageVersion string
}

// ImpressionData impression data for listener
type ImpressionData struct {
	Feature      string
	KeyName      string
	Treatment    string
	Time         int64
	ChangeNumber int64
	Label        string
	BucketingKey string
}

// WrapperImpressionListener struct
type WrapperImpressionListener struct {
	ImpressionListener ImpressionListener
}

// NewImpressionListenerWrapper instantiates a new ImpressionListenerWrapper
func NewImpressionListenerWrapper(impressionListener ImpressionListener) *WrapperImpressionListener {
	return &WrapperImpressionListener{
		ImpressionListener: impressionListener,
	}
}

// SendDataToClient sends the data to client
func (i *WrapperImpressionListener) SendDataToClient(impression dtos.ImpressionDTO, attributes map[string]interface{}, metadata dtos.QueueStoredMachineMetadataDTO) {
	impressionData := ImpressionData{
		KeyName:      impression.KeyName,
		Feature:      impression.FeatureName,
		BucketingKey: impression.BucketingKey,
		ChangeNumber: impression.ChangeNumber,
		Label:        impression.Label,
		Time:         impression.Time,
		Treatment:    impression.Treatment,
	}

	datToSend := ILObject{
		Impression:         impressionData,
		Attributes:         attributes,
		InstanceID:         metadata.MachineName,
		SDKLanguageVersion: "go-" + metadata.SDKVersion,
	}

	i.ImpressionListener.LogImpression(datToSend)
}
