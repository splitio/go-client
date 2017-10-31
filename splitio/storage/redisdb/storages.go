package redisdb

import (
	"encoding/json"
	"fmt"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
	"strconv"
	"strings"
)

// RedisSplitStorage is a redis-based implementation of split storage
type RedisSplitStorage struct {
	client prefixedRedisClient
	logger logging.LoggerInterface
}

// NewRedisSplitStorage creates a new RedisSplitStorage and returns a reference to it
func NewRedisSplitStorage(
	host string,
	port int,
	db int,
	password string,
	prefix string,
	logger logging.LoggerInterface,
) *RedisSplitStorage {
	return &RedisSplitStorage{
		client: *newPrefixedRedisClient(host, port, db, password, prefix),
		logger: logger,
	}
}

// Get fetches a feature in redis and returns a pointer to a split dto
func (r *RedisSplitStorage) Get(feature string) *dtos.SplitDTO {
	keyToFetch := strings.Replace(redisSplit, "{split}", feature, 1)
	val, err := r.client.Get(keyToFetch)

	if err != nil {
		r.logger.Error(fmt.Sprintf("Could not fetch feature \"%s\" from redis", feature))
		return nil
	}

	var split dtos.SplitDTO
	err = json.Unmarshal([]byte(val), &split)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Could not parse feature \"%s\" fetched from redis", feature))
		return nil
	}

	return &split
}

// PutMany bulk stores splits in redis
func (r *RedisSplitStorage) PutMany(splits []dtos.SplitDTO, changeNumber int64) {
	for _, split := range splits {
		keyToStore := strings.Replace(redisSplit, "{split}", split.Name, 1)
		raw, err := json.Marshal(split)
		if err != nil {
			r.logger.Error(fmt.Sprintf("Could not dump feature \"%s\" to json", split.Name))
			continue
		}

		err = r.client.Set(keyToStore, raw, 0)
		if err != nil {
			r.logger.Error(fmt.Sprintf("Could not store split \"%s\" in redis", split.Name))
			r.logger.Error(err.Error())
		}
	}
	err := r.client.Set(redisSplitTill, changeNumber, 0)
	if err != nil {
		r.logger.Error("Could not update split changenumber")
	}
}

// Remove revemoves a split from redis
func (r *RedisSplitStorage) Remove(splitName string) {
	keyToDelete := strings.Replace(redisSplit, "{split}", splitName, 1)
	_, err := r.client.Del(keyToDelete)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Error deleting split \"%s\".", splitName))
	}
}

// Till returns the latest split changeNumber
func (r *RedisSplitStorage) Till() int64 {
	val, err := r.client.Get(redisSplitTill)
	if err != nil {
		return -1
	}
	asInt, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		r.logger.Error("Could not parse Till value from redis")
		return -1
	}
	return asInt
}

// SplitNames returns a slice of strings with all the split names
func (r *RedisSplitStorage) SplitNames() []string {
	splitNames := make([]string, 0)
	keyPattern := strings.Replace(redisSplit, "{split}", "*", 1)
	keys, err := r.client.Keys(keyPattern)
	if err == nil {
		toRemove := strings.Replace(redisSplit, "{split}", "", 1) // Create a string with all the prefix to remove
		for _, key := range keys {
			splitNames = append(splitNames, strings.Replace(key, toRemove, "", 1)) // Extract split name from key
		}
	}
	return splitNames
}

// SegmentNames returns a slice of strings with all the segment names
func (r *RedisSplitStorage) SegmentNames() []string {
	segmentNames := make([]string, 0)
	keyPattern := strings.Replace(redisSplit, "{split}", "*", 1)
	keys, err := r.client.Keys(keyPattern)
	if err != nil {
		r.logger.Error("Error fetching split keys. Returning empty segment list")
		return segmentNames
	}

	for _, key := range keys {
		raw, err := r.client.Get(key)
		if err != nil {
			r.logger.Error(fmt.Sprintf("Fetching key \"%s\", skipping.", key))
			continue
		}

		var split dtos.SplitDTO
		err = json.Unmarshal([]byte(raw), &split)
		for _, condition := range split.Conditions {
			for _, matcher := range condition.MatcherGroup.Matchers {
				if matcher.UserDefinedSegment != nil {
					segmentNames = append(segmentNames, matcher.UserDefinedSegment.SegmentName)
				}
			}
		}
	}
	return segmentNames
}

// GetAll returns a slice of splits dtos.
func (r *RedisSplitStorage) GetAll() []dtos.SplitDTO {
	splits := make([]dtos.SplitDTO, 0)
	keyPattern := strings.Replace(redisSplit, "{split}", "*", 1)
	keys, err := r.client.Keys(keyPattern)
	if err != nil {
		r.logger.Error("Error fetching split keys. Returning empty split list")
		return splits
	}

	for _, key := range keys {
		raw, err := r.client.Get(key)
		if err != nil {
			r.logger.Error(fmt.Sprintf("Fetching key \"%s\", skipping.", key))
			continue
		}

		var split dtos.SplitDTO
		err = json.Unmarshal([]byte(raw), &split)
		splits = append(splits, split)
	}
	return splits
}

// RedisSegmentStorage is a redis implementation of a storage for segments
type RedisSegmentStorage struct {
	client prefixedRedisClient
	logger logging.LoggerInterface
}

// NewRedisSegmentStorage creates a new RedisSegmentStorage and returns a reference to it
func NewRedisSegmentStorage(
	host string,
	port int,
	db int,
	password string,
	prefix string,
	logger logging.LoggerInterface,
) *RedisSegmentStorage {
	return &RedisSegmentStorage{
		client: *newPrefixedRedisClient(host, port, db, password, prefix),
		logger: logger,
	}
}

// Get returns a segment wrapped in a set
func (r *RedisSegmentStorage) Get(segmentName string) *set.ThreadUnsafeSet {
	keyToFetch := strings.Replace(redisSegment, "{segment}", segmentName, 1)
	segmentKeys, err := r.client.SMembers(keyToFetch)
	if len(segmentKeys) <= 0 {
		r.logger.Error(fmt.Sprintf("Nonexsitant segment requested: \"%s\"", segmentName))
		return nil
	}
	if err != nil {
		r.logger.Error(fmt.Sprintf("Error retrieving memebers from set %s", segmentName))
		return nil
	}
	segment := set.NewSet()
	for _, member := range segmentKeys {
		segment.Add(member)
	}
	return segment
}

// Put (over)writes a segment in redis with the one passed to this function
func (r *RedisSegmentStorage) Put(name string, segment *set.ThreadUnsafeSet, changeNumber int64) {
	segmentKey := strings.Replace(redisSegment, "{segment}", name, 1)
	segmentTillKey := strings.Replace(redisSegmentTill, "{segment}", name, 1)
	err := r.client.WrapTransaction(func(p *prefixedTx) error {
		p.Del(segmentKey)
		p.SAdd(segmentKey, segment.List()...)
		p.Set(segmentTillKey, changeNumber, 0)
		// TODO CAPTURE ERRORS!
		return nil
	})

	if err != nil {
		r.logger.Error(fmt.Sprintf("Updating segment %s failed.", name))
		r.logger.Error(err.Error())
	}

}

// Remove removes a segment from storage
func (r *RedisSegmentStorage) Remove(segmentName string) {
	segmentKey := strings.Replace(redisSegment, "{segment}", segmentName, 1)
	segmentTillKey := strings.Replace(redisSegmentTill, "{segment}", segmentName, 1)
	count, err := r.client.Del(segmentKey, segmentTillKey)
	if count != 2 || err != nil {
		r.logger.Error(fmt.Sprintf("Error removing segment %s from cache.", segmentName))
	}
}

// Till returns the changeNumber for a particular segment
func (r *RedisSegmentStorage) Till(segmentName string) int64 {
	segmentKey := strings.Replace(redisSegmentTill, "{segment}", segmentName, 1)
	tillStr, err := r.client.Get(segmentKey)
	if err != nil {
		r.logger.Error("Error retrieving till. Returning -1")
		r.logger.Error(err.Error())
		return -1
	}

	asInt, err := strconv.ParseInt(tillStr, 10, 64)
	if err != nil {
		r.logger.Error("Error retrieving till. Returning -1")
		r.logger.Error(err.Error())
		return -1
	}
	return asInt
}

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
	r.client.WrapTransaction(func(p *prefixedTx) error {
		keys, err := p.Keys(strings.Replace(r.impTemplate, "{feature}", "*", 1))
		if err != nil {
			r.logger.Error("Could not retrieve impression keys from redis")
			return err
		}

		for _, key := range keys {
			members, err := p.Smembers(key)
			if err != nil {
				r.logger.Error(fmt.Sprintf("Could not retrieve impressions for key %s", key))
				r.logger.Error(err)
				continue
			}
			rawImpressions[strings.Replace(key, toRemove, "", 1)] = members
		}

		return nil
	})

	allImpressions := make([]dtos.ImpressionsDTO, len(rawImpressions))
	index := 0
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

		allImpressions[index] = dtos.ImpressionsDTO{
			TestName:       feature,
			KeyImpressions: featureImpressions,
		}

		index++
	}
	return allImpressions
}
