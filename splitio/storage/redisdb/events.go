package redisdb

import (
	"encoding/json"
	"sync"

	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
	"github.com/splitio/go-toolkit/redis"
)

// RedisEventsStorage redis implementation of EventsStorage interface
type RedisEventsStorage struct {
	client          *redis.PrefixedRedisClient
	mutex           *sync.Mutex
	logger          logging.LoggerInterface
	redisKey        string
	metadataMessage dtos.QueueStoredMachineMetadataDTO
}

// NewRedisEventsStorage returns an instance of RedisEventsStorage
func NewRedisEventsStorage(redisClient *redis.PrefixedRedisClient, metadata *splitio.SdkMetadata, logger logging.LoggerInterface) *RedisEventsStorage {
	return &RedisEventsStorage{
		client:   redisClient,
		logger:   logger,
		redisKey: redisEvents,
		mutex:    &sync.Mutex{},
		metadataMessage: dtos.QueueStoredMachineMetadataDTO{
			SDKVersion:  metadata.SDKVersion,
			MachineIP:   metadata.MachineIP,
			MachineName: metadata.MachineName,
		},
	}
}

// Push events into Redis LIST data type with RPUSH command
func (r *RedisEventsStorage) Push(event dtos.EventDTO, _ int) error {

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
	lrange, err := r.client.LRange(r.redisKey, 0, n-1)
	if err != nil {
		r.logger.Error("Fetching events", err.Error())
		r.mutex.Unlock()
		return nil, err
	}
	totalFetchedEvents := int64(len(lrange))

	idxFrom := n
	if totalFetchedEvents < n {
		idxFrom = totalFetchedEvents
	}

	err = r.client.LTrim(r.redisKey, idxFrom, -1)
	if err != nil {
		r.logger.Error("Trim events", err.Error())
		r.mutex.Unlock()
		return nil, err
	}
	r.mutex.Unlock()

	//JSON unmarshal
	for _, se := range lrange {
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
	return r.client.LLen(r.redisKey)
}

// Empty returns true if redis list is zero length
func (r *RedisEventsStorage) Empty() bool {
	return r.client.LLen(r.redisKey) == 0
}
