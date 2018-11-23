package dtos

//
// Events DTOs
//

// RedisStoredMachineMetadataDTO alias of QueueStoredMachineMetadataDTO
type RedisStoredMachineMetadataDTO QueueStoredMachineMetadataDTO

// QueueStoredMachineMetadataDTO maps sdk version, machine IP and machine name
type QueueStoredMachineMetadataDTO struct {
	SDKVersion  string `json:"s"`
	MachineIP   string `json:"i"`
	MachineName string `json:"n"`
}
