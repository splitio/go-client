package redisdb

import (
	"fmt"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
	"strconv"
	"strings"
)

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
