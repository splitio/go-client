package redisdb

import (
	"encoding/json"
	"sync"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
)

// RedisEventsStorage redis implementation of EventsStorage interface
type RedisEventsStorage struct {
	client          prefixedRedisClient
	mutex           *sync.Mutex
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
	instanceName string,
	sdkVersion string,
	logger logging.LoggerInterface,
) *RedisEventsStorage {
	return &RedisEventsStorage{
		client:          *newPrefixedRedisClient(host, port, db, password, prefix),
		logger:          logger,
		redisKey:        redisEvents,
		mutex:           &sync.Mutex{},
		metadataMessage: dtos.QueueStoredMachineMetadataDTO{SDKVersion: sdkVersion, MachineIP: instanceID, MachineName: instanceName},
	}
}

// Push events into Redis LIST data type with RPUSH command
func (r *RedisEventsStorage) Push(event dtos.EventDTO) error {

	var queueMessage = dtos.QueueStoredEventDTO{Metadata: r.metadataMessage, Event: event}

	eventJSON, err := json.Marshal(queueMessage)
	if err != nil {
		r.logger.Error("Something were wrong marshaling provided event to JSON", err.Error())
		return err
	}

	r.logger.Debug("Pushing events to:", r.redisKey, string(eventJSON))

	_, errPush := r.client.RPush(r.redisKey, eventJSON)
	if errPush != nil {
		r.logger.Error("Something were wrong pushing event to redis", errPush)
		return errPush
	}

	return nil
}

// PopN return N elements from 0 to N
func (r *RedisEventsStorage) PopN(n int64) ([]dtos.EventDTO, error) {
	toReturn := make([]dtos.EventDTO, 0)

	r.mutex.Lock()
	lrange := r.client.LRange(r.redisKey, 0, n-1)
	if lrange.Err() != nil {
		r.logger.Error("Fetching events", lrange.Err().Error())
		r.mutex.Unlock()
		return nil, lrange.Err()
	}
	totalFetchedEvents := int64(len(lrange.Val()))

	idxFrom := n
	if totalFetchedEvents < n {
		idxFrom = totalFetchedEvents
	}

	res := r.client.LTrim(r.redisKey, idxFrom, -1)
	if res.Err() != nil {
		r.logger.Error("Trim events", res.Err().Error())
		r.mutex.Unlock()
		return nil, res.Err()
	}
	r.mutex.Unlock()

	//JSON unmarshal
	listOfEvents := lrange.Val()
	for _, se := range listOfEvents {
		storedEventDTO := dtos.QueueStoredEventDTO{}
		err := json.Unmarshal([]byte(se), &storedEventDTO)
		if err != nil {
			r.logger.Error("Error decoding event JSON", err.Error())
			continue
		}
		if storedEventDTO.Metadata.MachineIP == r.metadataMessage.MachineIP &&
			storedEventDTO.Metadata.MachineName == r.metadataMessage.MachineName &&
			storedEventDTO.Metadata.SDKVersion == r.metadataMessage.SDKVersion {
			toReturn = append(toReturn, storedEventDTO.Event)
		}
	}

	return toReturn, nil
}

// Count returns the number of items in the redis list
func (r *RedisEventsStorage) Count() int64 {
	return r.client.LLen(r.redisKey).Val()
}

// Empty returns true if redis list is zero lenght
func (r *RedisEventsStorage) Empty() bool {
	return r.client.LLen(r.redisKey).Val() == 0
}
