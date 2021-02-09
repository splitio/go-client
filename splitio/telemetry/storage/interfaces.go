package storage

/*
type CustomTelemetryStorage struct {
	wrapper CustomTelemetryWrapper
}

func NewCustomTelemetryStorage(wrapper CustomTelemetryWrapper) TelemetryStorage {
	return &CustomTelemetryStorage{
		wrapper: wrapper,
	}
}

func (c *CustomTelemetryStorage) RecordSuccessfulEventsSync() {
	c.wrapper.Record("key", value)
}

// AddLatency adds latencies for method
func (c *CustomTelemetryStorage) AddLatency(name string, bucket int) {
	c.wrapper.Record(name, int64(bucket))
}

// Increment increments counter for method
func (c *CustomTelemetryStorage) Increment(name string, value int64) {
	c.wrapper.Increment(name, value)
}

////////////////////////////// CONSUMER //////////////////////////////

// GetFactories returns factories
func (c *CustomTelemetryStorage) GetFactories() map[string]int64 {
	toReturn := make(map[string]int64)
	// factories::12345678, 34
	// factories::12211121, 3
	factoriesData := c.wrapper.PopByPrefix("factories::")
	for key, item := range factoriesData {
		apikey, ok := key.(string)
		if !ok {
			continue
		}
		value, ok := item.(int64)
		if !ok {
			continue
		}
		toReturn[apikey] = value
	}
	return toReturn
}

// GetRecord gets record
func (c *CustomTelemetryStorage) GetRecord(key string) int64 {
	return c.wrapper.GetRecord(key)
}

// PopCounter pop
func (c *CustomTelemetryStorage) PopCounter(name string) int64 {
	return c.wrapper.PopItem(name)
}

func (c *CustomTelemetryStorage) popHTTPData(prefix string) map[int]int64 {
	toReturn := make(map[int]int64)
	httpData := c.wrapper.PopByPrefix(prefix)
	for key, value := range httpData {
		asString, ok := key.(string)
		if !ok {
			continue
		}
		splitted := strings.Split(asString, "::")
		if len(splitted) != 2 {
			continue
		}

		statusCode, err := strconv.Atoi(splitted[1])
		if err != nil {
			continue
		}
		counter, ok := value.(int64)
		if !ok {
			continue
		}
		toReturn[statusCode] = counter
	}
	return toReturn
}

// PopHTTPErrors returns errors stored
func (c *CustomTelemetryStorage) PopHTTPErrors() HTTPErrors {
	return HTTPErrors{
		Splits:      c.popHTTPData("split::"),
		Segments:    c.popHTTPData("segments::"),
		Impressions: c.popHTTPData("impressions::"),
		Events:      c.popHTTPData("events::"),
		Telemetry:   c.popHTTPData("telemetry::"),
		Token:       c.popHTTPData("token::"),
	}
}

// PopItems returns items
func (c *CustomTelemetryStorage) PopItems(name string) interface{} {
	switch name {
	case streamingType:
		toReturn := i.streamingEvents
		i.streamingEvents = make([]StreamingEvent, 0, maxStreamingEvents)
		return toReturn
	case tagsType:
		toReturn := i.tags
		i.tags = make([]string, 0, maxTags)
		return toReturn
	}
	return nil
}
*/
