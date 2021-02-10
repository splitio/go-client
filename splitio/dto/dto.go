package dto

// InitData struct
type InitData struct {
	OperationMode              int          `json:"oM,omitempty"`
	StreamingEnabled           bool         `json:"sE,omitempty"`
	Storage                    string       `json:"st,omitempty"`
	Rates                      Rates        `json:"rR"`
	URLOverrides               URLOverrides `json:"uO"`
	ImpressionsQueueSize       int64        `json:"iQ,omitempty"`
	EventsQueueSize            int64        `json:"eQ,omitempty"`
	ImpressionsMode            int          `json:"iM,omitempty"`
	ImpressionsListenerEnabled bool         `json:"iL,omitempty"`
	HTTPProxyDetected          bool         `json:"hP,omitempty"`
	ActiveFactories            int64        `json:"aF,omitempty"`
	RedundantFactories         int64        `json:"rF,omitempty"`
	TimeUntilReady             int64        `json:"tR,omitempty"`
	BurTimeouts                int64        `json:"bT,omitempty"`
	NonReadyUsages             int64        `json:"nR,omitempty"`
	Integrations               []string     `json:"i"`
	Tags                       []string     `json:"t"`
}

// Rates struct
type Rates struct {
	Splits      int64 `json:"sp,omitempty"`
	Segments    int64 `json:"se,omitempty"`
	Impressions int64 `json:"im,omitempty"`
	Events      int64 `json:"ev,omitempty"`
	Telemetry   int64 `json:"te,omitempty"`
}

// URLOverrides struct
type URLOverrides struct {
	Sdk       bool `json:"s,omitempty"`
	Events    bool `json:"e,omitempty"`
	Auth      bool `json:"a,omitempty"`
	Stream    bool `json:"st,omitempty"`
	Telemetry bool `json:"t,omitempty"`
}

// StatsData sent by sdks pereiodically
type StatsData struct {
	LastSynchronizations LastSynchronization `json:"lS"`
	MethodLatencies      MethodLatencies     `json:"mL"`
	MethodExceptions     MethodExceptions    `json:"mE"`
	HTTPErrors           HTTPErrors          `json:"hE"`
	HTTPLatencies        HTTPLatencies       `json:"hL"`
	TokenRefreshes       int64               `json:"tR,omitempty"`
	AuthRejections       int64               `json:"aR,omitempty"`
	ImpressionsQueued    int64               `json:"iQ,omitempty"`
	ImpressionsDeduped   int64               `json:"iDe,omitempty"`
	ImpressionsDropped   int64               `json:"iDr,omitempty"`
	SplitCount           int64               `json:"spC,omitempty"`
	SegmentCount         int64               `json:"seC,omitempty"`
	SegmentKeyCount      int64               `json:"skC,omitempty"`
	SessionLengthMs      int64               `json:"sL,omitempty"`
	EventsQueued         int64               `json:"eQ,omitempty"`
	EventsDropped        int64               `json:"eD,omitempty"`
	StreamingEvents      []StreamingEvent    `json:"sE"`
	Tags                 []string            `json:"t"`
}

// LastSynchronization struct
type LastSynchronization struct {
	Splits      int64 `json:"sp,omitempty"`
	Segments    int64 `json:"se,omitempty"`
	Impressions int64 `json:"im,omitempty"`
	Events      int64 `json:"ev,omitempty"`
	Token       int64 `json:"to,omitempty"`
	Telemetry   int64 `json:"te,omitempty"`
}

// HTTPErrors struct
type HTTPErrors struct {
	Splits      map[int]int64 `json:"sp,omitempty"`
	Segments    map[int]int64 `json:"se,omitempty"`
	Impressions map[int]int64 `json:"im,omitempty"`
	Events      map[int]int64 `json:"ev,omitempty"`
	Token       map[int]int64 `json:"to,omitempty"`
	Telemetry   map[int]int64 `json:"te,omitempty"`
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
	Treatment            int64 `json:"t,omitempty"`
	Treatments           int64 `json:"ts,omitempty"`
	TreatmentWithConfig  int64 `json:"tc,omitempty"`
	TreatmentWithConfigs int64 `json:"tcs,omitempty"`
	Track                int64 `json:"tr,omitempty"`
}

// StreamingEvent struct
type StreamingEvent struct {
	Type      int   `json:"e,omitempty"`
	Data      int64 `json:"d,omitempty"`
	Timestamp int64 `json:"t,omitempty"`
}
