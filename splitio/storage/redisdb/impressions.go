package redisdb

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/splitio/go-client/splitio/service/dtos"
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
func (r *RedisImpressionStorage) LogImpressions(impressions []dtos.ImpressionsDTO) error {
	if len(impressions) > 0 {
		var impressionsToStore []dtos.ImpressionsQueueDTO
		for _, i := range impressions {
			var impression = dtos.ImpressionsQueueDTO{Metadata: r.metadataMessage, Impression: dtos.ImpressionQueueDTO{
				KeyName:      i.KeyImpressions[0].KeyName,
				BucketingKey: i.KeyImpressions[0].BucketingKey,
				FeatureName:  i.TestName,
				Treatment:    i.KeyImpressions[0].Treatment,
				Label:        i.KeyImpressions[0].Label,
				ChangeNumber: i.KeyImpressions[0].ChangeNumber,
				Time:         i.KeyImpressions[0].Time,
			}}
			impressionsToStore = append(impressionsToStore, impression)
		}
		return r.Push(impressionsToStore)
	}
	return nil
}

// Push stores impressions in redis
func (r *RedisImpressionStorage) Push(impressions []dtos.ImpressionsQueueDTO) error {
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
func (r *RedisImpressionStorage) PopN(n int64) ([]dtos.ImpressionQueueDTO, error) {
	toReturn := make([]dtos.ImpressionQueueDTO, 0)

	r.mutex.Lock()
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
	r.mutex.Unlock()

	//JSON unmarshal
	listOfImpressions := lrange.Val()
	for _, se := range listOfImpressions {
		storedImpressionDTO := dtos.ImpressionsQueueDTO{}
		err := json.Unmarshal([]byte(se), &storedImpressionDTO)
		if err != nil {
			r.logger.Error("Error decoding impression JSON", err.Error())
			continue
		}
		if storedImpressionDTO.Metadata.MachineIP == r.metadataMessage.MachineIP &&
			storedImpressionDTO.Metadata.MachineName == r.metadataMessage.MachineName &&
			storedImpressionDTO.Metadata.SDKVersion == r.metadataMessage.SDKVersion {
			toReturn = append(toReturn, storedImpressionDTO.Impression)
		}
	}

	return toReturn, nil
}

// PopAll returns and clears all impressions in redis.
func (r *RedisImpressionStorage) PopAll() []dtos.ImpressionsDTO {
	toRemove := strings.Replace(r.impTemplate, "{feature}", "", 1) // String that will be removed from every key
	rawImpressions := make(map[string][]string)
	err := r.client.WrapTransaction(func(t *prefixedTx) error {
		keys, err := t.Keys(strings.Replace(r.impTemplate, "{feature}", "*", 1))
		if err != nil {
			r.logger.Error("Could not retrieve impression keys from redis")
			return err
		}

		for _, key := range keys {
			members, err := t.Smembers(key)
			if err != nil {
				r.logger.Error(fmt.Sprintf("Could not retrieve impressions for key %s", key))
				r.logger.Error(err)
				continue
			}
			rawImpressions[strings.Replace(key, toRemove, "", 1)] = members
			t.Del(key)
		}
		return nil
	})
	if err != nil {
		r.logger.Error(err)
		return nil
	}

	all := make([]dtos.ImpressionsDTO, len(rawImpressions))
	allIndex := 0
	for feature, impressions := range rawImpressions {
		featureImpressions := make([]dtos.ImpressionDTO, len(impressions))
		for impressionIndex, impression := range impressions {
			var impDto dtos.ImpressionDTO
			err := json.Unmarshal([]byte(impression), &impDto)
			if err != nil {
				r.logger.Error("Could not decode json-stored impression")
				r.logger.Error(err)
				continue
			}
			featureImpressions[impressionIndex] = impDto
		}
		all[allIndex] = dtos.ImpressionsDTO{TestName: feature, KeyImpressions: featureImpressions}
		allIndex++
	}
	return all
}
