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
}

// NewImpressionListenerWrapper instantiates a new ImpressionListenerWrapper
func NewImpressionListenerWrapper(impressionListener ImpressionListener) *WrapperImpressionListener {
	return &WrapperImpressionListener{
		ImpressionListener: impressionListener,
	}
}

// SendDataToClient sends the data to client
func (i *WrapperImpressionListener) SendDataToClient(impression storage.Impression, attributes map[string]interface{}, metadata splitio.SdkMetadata) {
	datToSend := ILObject{
		Impression:         impression,
		Attributes:         attributes,
		InstanceID:         metadata.MachineName,
		SDKLanguageVersion: "go-" + metadata.SDKVersion,
	}

	i.ImpressionListener.LogImpression(datToSend)
}
