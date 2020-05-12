package redisdb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
	"github.com/splitio/go-toolkit/redis"
)

// RedisSegmentStorage is a redis implementation of a storage for segments
type RedisSegmentStorage struct {
	client redis.PrefixedRedisClient
	logger logging.LoggerInterface
}

// NewRedisSegmentStorage creates a new RedisSegmentStorage and returns a reference to it
func NewRedisSegmentStorage(redisClient *redis.PrefixedRedisClient, logger logging.LoggerInterface) *RedisSegmentStorage {
	return &RedisSegmentStorage{
		client: *redisClient,
		logger: logger,
	}
}

// Get returns a segment wrapped in a set
func (r *RedisSegmentStorage) Get(segmentName string) *set.ThreadUnsafeSet {
	keyToFetch := strings.Replace(redisSegment, "{segment}", segmentName, 1)
	segmentKeys, err := r.client.SMembers(keyToFetch)
	if len(segmentKeys) <= 0 {
		r.logger.Warning(fmt.Sprintf("Nonexsitant segment requested: \"%s\"", segmentName))
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

// Till returns the changeNumber for a particular segment
func (r *RedisSegmentStorage) Till(segmentName string) int64 {
	segmentKey := strings.Replace(redisSegmentTill, "{segment}", segmentName, 1)
	tillStr, err := r.client.Get(segmentKey)
	if err != nil {
		return -1
	}

	asInt, err := strconv.ParseInt(tillStr, 10, 64)
	if err != nil {
		r.logger.Error("Error retrieving till. Returning -1: ", err.Error())
		return -1
	}
	return asInt
}

// Clear removes all splits from storage
func (r *RedisSegmentStorage) Clear() {
}
