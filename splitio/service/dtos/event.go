package dtos

//
// Events DTOs
//

// RedisStoredEventDTO alias of QueueStoredEventDTO
type RedisStoredEventDTO QueueStoredEventDTO

// EventDTO struct mapping events json
type EventDTO struct {
	Key             string                 `json:"key"`
	TrafficTypeName string                 `json:"trafficTypeName"`
	EventTypeID     string                 `json:"eventTypeId"`
	Value           interface{}            `json:"value"`
	Timestamp       int64                  `json:"timestamp"`
	Properties      map[string]interface{} `json:"properties"`
}

// QueueStoredEventDTO maps the stored JSON object in redis by SDKs
type QueueStoredEventDTO struct {
	Metadata QueueStoredMachineMetadataDTO `json:"m"`
	Event    EventDTO                      `json:"e"`
}
