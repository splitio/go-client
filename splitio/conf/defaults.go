package conf

const (
	defaultHTTPTimeout      = 30
	defaultTaskPeriod       = 60
	defaultRedisHost        = "localhost"
	defaultRedisPort        = 6379
	defaultRedisDb          = 0
	defaultSegmentQueueSize = 500
	defaultSegmentWorkers   = 10
)

const (
	minSplitSync      = 5
	minSegmentSync    = 30
	minImpressionSync = 1
	minEventSync      = 1
	minTelemetrySync  = 30
)
