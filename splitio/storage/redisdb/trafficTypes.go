package redisdb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/splitio/go-toolkit/logging"
)

// RedisTrafficTypeStorage is a redis-based implementation of traffic type storage
type RedisTrafficTypeStorage struct {
	client prefixedRedisClient
	logger logging.LoggerInterface
}

// NewRedisTrafficTypeStorage creates a new RedisTrafficTypeStorage and returns a reference to it
func NewRedisTrafficTypeStorage(
	host string,
	port int,
	db int,
	password string,
	prefix string,
	logger logging.LoggerInterface,
) *RedisTrafficTypeStorage {
	return &RedisTrafficTypeStorage{
		client: *newPrefixedRedisClient(host, port, db, password, prefix),
		logger: logger,
	}
}

// Increase increases value for a traffic type
func (r *RedisTrafficTypeStorage) Increase(trafficType string) {}

// Decrease decreases value for a traffic type
func (r *RedisTrafficTypeStorage) Decrease(trafficType string) {}

// Get fetches the value for a particular traffic type
func (r *RedisTrafficTypeStorage) Get(trafficType string) int64 {
	keyToFetch := strings.Replace(redisTrafficType, "{trafficType}", trafficType, 1)
	res, err := r.client.Get(keyToFetch)

	if err != nil {
		r.logger.Error(fmt.Sprintf("Could not fetch trafficType \"%s\" from redis: %s", trafficType, err.Error()))
		return 0
	}

	val, err := strconv.ParseInt(res, 10, 64)
	if err != nil {
		fmt.Println(err)
		r.logger.Error("TrafficType could not be converted")
		return 0
	}
	return val
}
