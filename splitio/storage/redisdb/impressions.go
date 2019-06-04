package redisdb

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-toolkit/logging"
)

// RedisImpressionStorage is a redis-based implementation of split storage
type RedisImpressionStorage struct {
	client          prefixedRedisClient
	mutex           *sync.Mutex
	logger          logging.LoggerInterface
	impTemplate     string
	redisKey        string
	impressionsTTL  time.Duration
	metadataMessage dtos.QueueStoredMachineMetadataDTO
}

// NewRedisImpressionStorage creates a new RedisSplitStorage and returns a reference to it
func NewRedisImpressionStorage(
	host string,
	port int,
	db int,
	password string,
	prefix string,
	instanceID string,
	instanceName string,
	sdkVersion string,
	logger logging.LoggerInterface,
) *RedisImpressionStorage {
	impTemplate := strings.Replace(redisImpressions, "{sdkVersion}", sdkVersion, 1)
	impTemplate = strings.Replace(impTemplate, "{instanceId}", instanceID, 1)
	return &RedisImpressionStorage{
		client:          *newPrefixedRedisClient(host, port, db, password, prefix),
		mutex:           &sync.Mutex{},
		logger:          logger,
		impTemplate:     impTemplate,
		redisKey:        redisImpressionsQueue,
		impressionsTTL:  redisImpressionsTTL,
		metadataMessage: dtos.QueueStoredMachineMetadataDTO{SDKVersion: sdkVersion, MachineIP: instanceID, MachineName: instanceName},
	}
}

// LogImpressions stores impressions in redis as Queue
func (r *RedisImpressionStorage) LogImpressions(impressions []storage.Impression) error {
	var impressionsToStore []storage.ImpressionQueueObject
	for _, i := range impressions {
		var impression = storage.ImpressionQueueObject{Metadata: r.metadataMessage, Impression: i}
		impressionsToStore = append(impressionsToStore, impression)
	}

	if len(impressionsToStore) > 0 {
		return r.Push(impressionsToStore)
	}
	return nil
}

// Push stores impressions in redis
func (r *RedisImpressionStorage) Push(impressions []storage.ImpressionQueueObject) error {
	var impressionsJSON []interface{}
	for _, impression := range impressions {
		iJSON, err := json.Marshal(impression)
		if err != nil {
			r.logger.Error("Error encoding impression in json")
			r.logger.Error(err)
		} else {
			impressionsJSON = append(impressionsJSON, string(iJSON))
		}
	}

	r.logger.Debug("Pushing impressions to: ", r.redisKey, len(impressionsJSON))

	inserted, errPush := r.client.RPush(r.redisKey, impressionsJSON...)
	if errPush != nil {
		r.logger.Error("Something were wrong pushing impressions to redis", errPush)
		return errPush
	}

	// Checks if expiration needs to be set
	if inserted == int64(len(impressionsJSON)) {
		r.logger.Debug("Proceeding to set expiration for: ", r.redisKey)
		result := r.client.Expire(r.redisKey, time.Duration(r.impressionsTTL)*time.Minute).Val()
		if result == false {
			r.logger.Error("Something were wrong setting expiration", errPush)
		}
	}
	return nil
}

// PopN return N elements from 0 to N
func (r *RedisImpressionStorage) PopN(n int64) ([]storage.Impression, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	toReturn := make([]storage.Impression, 0)

	lrange := r.client.LRange(r.redisKey, 0, n-1)
	if lrange.Err() != nil {
		r.logger.Error("Fetching impressions", lrange.Err().Error())
		r.mutex.Unlock()
		return nil, lrange.Err()
	}
	totalFetchedImpressions := int64(len(lrange.Val()))

	idxFrom := n
	if totalFetchedImpressions < n {
		idxFrom = totalFetchedImpressions
	}

	res := r.client.LTrim(r.redisKey, idxFrom, -1)
	if res.Err() != nil {
		r.logger.Error("Trim impressions", res.Err().Error())
		r.mutex.Unlock()
		return nil, res.Err()
	}

	//JSON unmarshal
	listOfImpressions := lrange.Val()
	for _, se := range listOfImpressions {
		storedImpression := storage.ImpressionQueueObject{}
		err := json.Unmarshal([]byte(se), &storedImpression)
		if err != nil {
			r.logger.Error("Error decoding impression JSON", err.Error())
			continue
		}
		if storedImpression.Metadata.MachineIP == r.metadataMessage.MachineIP &&
			storedImpression.Metadata.MachineName == r.metadataMessage.MachineName &&
			storedImpression.Metadata.SDKVersion == r.metadataMessage.SDKVersion {
			toReturn = append(toReturn, storedImpression.Impression)
		}
	}

	return toReturn, nil
}
