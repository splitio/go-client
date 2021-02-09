package telemetry

const (
	latencyBucketCount = 23
)

// InitData struct
type InitData struct {
	OperationMode              int          `json:"oM"`
	StreamingEnabled           bool         `json:"sE"`
	Storage                    string       `json:"st"`
	Rates                      Rates        `json:"rR"`
	URLOverrides               URLOverrides `json:"uO"`
	ImpressionsQueueSize       int64        `json:"iQ"`
	EventsQueueSize            int64        `json:"eQ"`
	ImpressionsMode            int          `json:"iM"`
	ImpressionsListenerEnabled bool         `json:"iL"`
	HTTPProxyDetected          bool         `json:"hP"`
	ActiveFactories            int64        `json:"aF"`
	RedundantFactories         int64        `json:"rF"`
	TimeUntilReady             int64        `json:"tR"`
	BurTimeouts                int64        `json:"bT"`
	NonReadyUsages             int64        `json:"nR"`
	Tags                       []string     `json:"t"`
}

// Rates struct
type Rates struct {
	Splits      int64 `json:"sp"`
	Segments    int64 `json:"se"`
	Impressions int64 `json:"im"`
	Events      int64 `json:"ev"`
	Telemetry   int64 `json:"te"`
}

// URLOverrides struct
type URLOverrides struct {
	Sdk       bool `json:"s"`
	Events    bool `json:"e"`
	Auth      bool `json:"a"`
	Stream    bool `json:"st"`
	Telemetry bool `json:"t"`
}

// StatsData sent by sdks pereiodically
type StatsData struct {
	LastSynchronizations LastSynchronization `json:"lS"`
	MethodLatencies      MethodLatencies     `json:"mL"`
	MethodExceptions     MethodExceptions    `json:"mE"`
	HTTPErrors           HTTPErrors          `json:"hE"`
	HTTPLatencies        HTTPLatencies       `json:"hL"`
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

// HTTPLatencies struct
type HTTPLatencies struct {
	Splits      []int64 `json:"sp"`
	Segments    []int64 `json:"se"`
	Impressions []int64 `json:"im"`
	Events      []int64 `json:"ev"`
	Token       []int64 `json:"to"`
	Telemetry   []int64 `json:"te"`
}

// MethodLatencies struct
type MethodLatencies struct {
	Treatment            []int64 `json:"t"`
	Treatments           []int64 `json:"ts"`
	TreatmentWithConfig  []int64 `json:"tc"`
	TreatmentWithConfigs []int64 `json:"tcs"`
	Track                []int64 `json:"tr"`
}

// MethodExceptions struct
type MethodExceptions struct {
	Treatment            int64 `json:"t"`
	Treatments           int64 `json:"ts"`
	TreatmentWithConfig  int64 `json:"tc"`
	TreatmentWithConfigs int64 `json:"tcs"`
	Track                int64 `json:"tr"`
}

// StreamingEvent struct
type StreamingEvent struct {
	Type      int   `json:"e"`
	Data      int64 `json:"d"`
	Timestamp int64 `json:"t"`
}
