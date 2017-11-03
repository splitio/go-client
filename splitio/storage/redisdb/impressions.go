package redisdb

import (
	"encoding/json"
	"fmt"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
	"strings"
)

// RedisImpressionStorage is a redis-based implementation of split storage
type RedisImpressionStorage struct {
	client      prefixedRedisClient
	logger      logging.LoggerInterface
	impTemplate string
}

// NewRedisImpressionStorage creates a new RedisSplitStorage and returns a reference to it
func NewRedisImpressionStorage(
	host string,
	port int,
	db int,
	password string,
	prefix string,
	instanceID string,
	sdkVersion string,
	logger logging.LoggerInterface,
) *RedisImpressionStorage {
	impTemplate := strings.Replace(redisImpressions, "{sdkVersion}", sdkVersion, 1)
	impTemplate = strings.Replace(impTemplate, "{instanceId}", instanceID, 1)
	return &RedisImpressionStorage{
		client:      *newPrefixedRedisClient(host, port, db, password, prefix),
		logger:      logger,
		impTemplate: impTemplate,
	}
}

// Put stores an impression in redis
func (r *RedisImpressionStorage) Put(feature string, impression *dtos.ImpressionDTO) {
	keyToStore := strings.Replace(r.impTemplate, "{feature}", feature, 1)
	encoded, err := json.Marshal(impression)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Error encoding impression in json for feature %s", feature))
		r.logger.Error(err)
		return
	}

	_, err = r.client.SAdd(keyToStore, string(encoded))
	if err != nil {
		r.logger.Error(fmt.Sprintf("Error storing impression in redis for feature %s", feature))
		r.logger.Error(err)
	}
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
