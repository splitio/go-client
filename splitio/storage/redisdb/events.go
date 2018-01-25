package redisdb

import (
	"encoding/json"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
)

// RedisEventsStorage redis implementation of EventsStorage interface
type RedisEventsStorage struct {
	client          prefixedRedisClient
	logger          logging.LoggerInterface
	redisKey        string
	metadataMessage dtos.QueueStoredMachineMetadataDTO
}

// NewRedisEventsStorage returns an instance of RedisEventsStorage
func NewRedisEventsStorage(
	host string,
	port int,
	db int,
	password string,
	prefix string,
	instanceID string,
	sdkVersion string,
	logger logging.LoggerInterface,
) *RedisEventsStorage {
	return &RedisEventsStorage{
		client:          *newPrefixedRedisClient(host, port, db, password, prefix),
		logger:          logger,
		redisKey:        redisEvents,
		metadataMessage: dtos.QueueStoredMachineMetadataDTO{SDKVersion: sdkVersion, MachineIP: instanceID, MachineName: "unknown"},
	}
}

// Push events into Redis LIST data type with RPUSH command
func (r *RedisEventsStorage) Push(event dtos.EventDTO) error {

	var queueMessage = dtos.QueueStoredEventDTO{Metadata: r.metadataMessage, Event: event}

	eventJSON, err := json.Marshal(queueMessage)
	if err != nil {
		r.logger.Error("Somethig were wrong marshaling provided event to JSON", err.Error())
		return err
	}

	_, errPush := r.client.RPush(r.redisKey, eventJSON)
	if errPush != nil {
		r.logger.Error("Something were wrong pushing event to redis", errPush)
		return errPush
	}

	return nil
}

// PopN return N elements from 0 to N
func (r *RedisEventsStorage) PopN(n int64) ([]dtos.QueueStoredEventDTO, error) {
	return nil, nil
}
