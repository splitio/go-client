package telemetry

// RegularMetrics sent by sdks pereiodically
type RegularMetrics struct {
	MethodLatencies      map[string][]int64  `json:"mL"`
	MethodExceptions     map[string]int64    `json:"mE"`
	LastSynchronizations LastSynchronization `json:"lS"`
	HTTPErrors           HTTPErrors          `json:"hE"`
	TokenRefreshes       int64               `json:"tR"`
	AuthRejections       int64               `json:"aR"`
	ImpressionsQueued    int64               `json:"iQ"`
	ImpressionsDeduped   int64               `json:"iDe"`
	ImpressionsDropped   int64               `json:"iDr"`
	SplitCount           int64               `json:"spC"`
	SegmentCount         int64               `json:"seC"`
	SegmentKeyCount      int64               `json:"skC"`
	SessionLengthMs      int64               `json:"sL"`
	EventsQueued         int64               `json:"eQ"`
	EventsDropped        int64               `json:"eD"`
	StreamingEvents      []StreamingEvent    `json:"sE"`
	Tags                 []string            `json:"t"`
}

// LastSynchronization struct
type LastSynchronization struct {
	Splits      int64 `json:"sp"`
	Segments    int64 `json:"se"`
	Impressions int64 `json:"im"`
	Events      int64 `json:"ev"`
	Token       int64 `json:"to"`
	Telemetry   int64 `json:"te"`
}

// HTTPErrors struct
type HTTPErrors struct {
	Splits      map[int]int64 `json:"sp"`
	Segments    map[int]int64 `json:"se"`
	Impressions map[int]int64 `json:"im"`
	Events      map[int]int64 `json:"ev"`
	Token       map[int]int64 `json:"to"`
	Telemetry   map[int]int64 `json:"te"`
}

// StreamingEvent struct
type StreamingEvent struct {
	Type      string `json:"event"`
	Data      int64  `json:"data"`
	Timestamp int64  `json:"timestamp"`
}
