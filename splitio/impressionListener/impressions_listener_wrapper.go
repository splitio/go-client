package impressionlistener

import (
	"fmt"

	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/service/dtos"
)

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
func (i *WrapperImpressionListener) SendDataToClient(impression dtos.ImpressionsDTO, attributes map[string]interface{}) {
	fmt.Println("KEY")
	fmt.Println(impression.TestName)

	impressionData := dtos.ImpressionDataDTO{
		KeyName:      impression.KeyImpressions[0].KeyName,
		Feature:      impression.TestName,
		BucketingKey: impression.KeyImpressions[0].BucketingKey,
		ChangeNumber: impression.KeyImpressions[0].ChangeNumber,
		Label:        impression.KeyImpressions[0].Label,
		Time:         impression.KeyImpressions[0].Time,
		Treatment:    impression.KeyImpressions[0].Treatment,
	}

	datToSend := dtos.ImpressionListenerDTO{
		Impression:         impressionData,
		Attributes:         attributes,
		InstanceID:         "",
		SDKLanguageVersion: splitio.Version,
	}

	i.ImpressionListener.LogImpression(datToSend)
}
