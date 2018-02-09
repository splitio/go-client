package dtos

//
// Events DTOs
//

// RedisStoredMachineMetadataDTO alias of QueueStoredMachineMetadataDTO
type RedisStoredMachineMetadataDTO QueueStoredMachineMetadataDTO

// RedisStoredEventDTO alias of QueueStoredEventDTO
type RedisStoredEventDTO QueueStoredEventDTO

// EventDTO struct mapping events json
type EventDTO struct {
	Key             string      `json:"key"`
	TrafficTypeName string      `json:"trafficTypeName"`
	EventTypeID     string      `json:"eventTypeId"`
	Value           interface{} `json:"value"`
	Timestamp       int64       `json:"timestamp"`
}

// QueueStoredMachineMetadataDTO maps sdk version, machine IP and machine name
type QueueStoredMachineMetadataDTO struct {
	SDKVersion  string `json:"s"`
	MachineIP   string `json:"i"`
	MachineName string `json:"n"`
}

// QueueStoredEventDTO maps the stored JSON object in redis by SDKs
type QueueStoredEventDTO struct {
	Metadata QueueStoredMachineMetadataDTO `json:"m"`
	Event    EventDTO                      `json:"e"`
}
