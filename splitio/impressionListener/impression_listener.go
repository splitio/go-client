package impressionlistener

import "github.com/splitio/go-client/splitio/service/dtos"

// ImpressionListener declaration of ImpressionListener interface
type ImpressionListener interface {
	LogImpression(data dtos.ImpressionListenerDTO)
}
