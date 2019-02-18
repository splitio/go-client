package impressionlistener

import (
	"github.com/splitio/go-client/splitio"
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
func (i *WrapperImpressionListener) SendDataToClient(impression dtos.ImpressionsDTO, attributes map[string]interface{}, instanceName string) {
	if len(impression.KeyImpressions) > 0 {
		impressionData := ImpressionData{
			KeyName:      impression.KeyImpressions[0].KeyName,
			Feature:      impression.TestName,
			BucketingKey: impression.KeyImpressions[0].BucketingKey,
			ChangeNumber: impression.KeyImpressions[0].ChangeNumber,
			Label:        impression.KeyImpressions[0].Label,
			Time:         impression.KeyImpressions[0].Time,
			Treatment:    impression.KeyImpressions[0].Treatment,
		}

		datToSend := ILObject{
			Impression:         impressionData,
			Attributes:         attributes,
			InstanceID:         instanceName,
			SDKLanguageVersion: "go-" + splitio.Version,
		}

		i.ImpressionListener.LogImpression(datToSend)
	}
}
