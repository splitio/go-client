package impressionlistener

import (
	"github.com/splitio/go-client/splitio/service/dtos"
)

// ListenerInterface interface to be implemented by structs that will subscribe to impressions publishing
type ListenerInterface interface {
	// The method notify will be called in it's own goroutine.
	// It's the developer's responsibility that this function takes less time
	// than the impression recording period in cfg.Advanced, otherwhise executions
	// may overlap causing unexpected behaviour.
	Notify(impressionBulk []dtos.ImpressionsDTO)
}
