package impressionlistener

import (
	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/storage"
)

// ILObject struct to map entire data for listener
type ILObject struct {
	Impression         storage.Impression
	Attributes         map[string]interface{}
	InstanceID         string
	SDKLanguageVersion string
}

// WrapperImpressionListener struct
type WrapperImpressionListener struct {
	ImpressionListener ImpressionListener
	metadata           *splitio.SdkMetadata
}

// NewImpressionListenerWrapper instantiates a new ImpressionListenerWrapper
func NewImpressionListenerWrapper(impressionListener ImpressionListener, metadata *splitio.SdkMetadata) *WrapperImpressionListener {
	return &WrapperImpressionListener{
		ImpressionListener: impressionListener,
		metadata:           metadata,
	}
}

// SendDataToClient sends the data to client
func (i *WrapperImpressionListener) SendDataToClient(impressions []storage.Impression, attributes map[string]interface{}) {
	for _, impression := range impressions {
		datToSend := ILObject{
			Impression:         impression,
			Attributes:         attributes,
			InstanceID:         i.metadata.MachineName,
			SDKLanguageVersion: i.metadata.SDKVersion,
		}

		i.ImpressionListener.LogImpression(datToSend)
	}
}
