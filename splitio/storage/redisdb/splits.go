package redisdb

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
)

// RedisSplitStorage is a redis-based implementation of split storage
type RedisSplitStorage struct {
	client *PrefixedRedisClient
	logger logging.LoggerInterface
}

// NewRedisSplitStorage creates a new RedisSplitStorage and returns a reference to it
func NewRedisSplitStorage(redisClient *PrefixedRedisClient, logger logging.LoggerInterface) *RedisSplitStorage {
	return &RedisSplitStorage{
		client: redisClient,
		logger: logger,
	}
}

// Get fetches a feature in redis and returns a pointer to a split dto
func (r *RedisSplitStorage) Get(feature string) *dtos.SplitDTO {
	keyToFetch := strings.Replace(redisSplit, "{split}", feature, 1)
	val, err := r.client.Get(keyToFetch)

	if err != nil {
		r.logger.Error(fmt.Sprintf("Could not fetch feature \"%s\" from redis: %s", feature, err.Error()))
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
			r.logger.Error(fmt.Sprintf("Could not store split \"%s\" in redis: %s", split.Name, err.Error()))
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
func (r *RedisSplitStorage) SegmentNames() *set.ThreadUnsafeSet {
	segmentNames := set.NewSet()
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
		if err != nil {
			r.logger.Error(fmt.Sprintf("Error parsing json for split %s", key))
			continue
		}
		for _, condition := range split.Conditions {
			for _, matcher := range condition.MatcherGroup.Matchers {
				if matcher.UserDefinedSegment != nil {
					segmentNames.Add(matcher.UserDefinedSegment.SegmentName)
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
		if err != nil {
			r.logger.Error(fmt.Sprintf("Error parsing json for split %s", key))
			continue
		}
		splits = append(splits, split)
	}
	return splits
}

// Clear removes all splits from storage
func (r *RedisSplitStorage) Clear() {
	r.client.WrapTransaction(func(t *prefixedTx) error {
		keys, err := t.Keys(strings.Replace(redisSplit, "{split}", "*", 1))
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			err = t.Del(keys...)
		}

		return err
	})
}

// TrafficTypeExists returns true or false depending on existence and counter
// of trafficType
func (r *RedisSplitStorage) TrafficTypeExists(trafficType string) bool {
	keyToFetch := strings.Replace(redisTrafficType, "{trafficType}", trafficType, 1)
	res, err := r.client.Get(keyToFetch)

	if err != nil {
		r.logger.Error(fmt.Sprintf("Could not fetch trafficType \"%s\" from redis: %s", trafficType, err.Error()))
		return false
	}

	val, err := strconv.ParseInt(res, 10, 64)
	if err != nil {
		r.logger.Error("TrafficType could not be converted")
		return false
	}
	return val > 0
}
