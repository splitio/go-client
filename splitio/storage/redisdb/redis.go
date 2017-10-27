package redisdb

import (
	"fmt"
	"github.com/go-redis/redis"
	"strings"
	"time"
)

const (
	redisSplit       = "SPLITIO.split.{split}"                                              // split object
	redisSplitTill   = "SPLITIO.splits.till"                                                // last split fetch
	redisSegments    = "SPLITIO.segments.registered"                                        // segments that appear in fetched splits
	redisSegment     = "SPLITIO.segment.{segment}"                                          // segment object
	redisSegmentTill = "SPLITIO.segment.{segment}.till"                                     // last segment fetch
	redisImpressions = "SPLITIO/{sdkVersion}/{instanceId}/impressions.{feature}"            // impressions for a feature
	redisLatency     = "SPLITIO/{sdkVersion}/{instanceId}/latency.{metric}.bucket.{bucket}" // latency bucket
	redisCount       = "SPLITIO/{sdkVersion}/{instanceId}/count.{metric}"                   // counter
	redisGauge       = "SPLITIO/{sdkVersion}/{instanceId}/gauge.{metric}"                   // gauge
)

type prefixable struct {
	prefix string
}

// withPrefix adds a prefix to the key if the prefix supplied is different from an empty string
func (p *prefixable) withPrefix(key string) string {
	if len(p.prefix) > 0 {
		return fmt.Sprintf("%s.%s", p.prefix, key)
	}
	return key
}

// prefixedPipe struct is used to as a wrapper used for redis.Pipeliner in order to
// automatically populate the prefix if any
type prefixedPipe struct {
	prefixable
	pipe redis.Pipeliner
}

// wrap redis "set" operation with a prefix inside a transaction
func (p *prefixedPipe) Set(key string, value interface{}, expiration time.Duration) {
	p.pipe.Set(p.withPrefix(key), value, expiration)
}

// wrap redis "sadd" operation with a prefix inside a transaction
func (p *prefixedPipe) SAdd(key string, members ...interface{}) {
	p.pipe.SAdd(p.withPrefix(key), members...)
}

// wrap redis "del" operation with a prefix inside a transaction
func (p *prefixedPipe) Del(key string) {
	p.pipe.Del(p.withPrefix(key))
}

// newPrefixedPipe instantiates a new pipewrapper and returns a reference
func newPrefixedPipe(pipe redis.Pipeliner, prefix string) *prefixedPipe {
	return &prefixedPipe{
		prefixable: prefixable{prefix: prefix},
		pipe:       pipe,
	}
}

// ---------

// prefixedRedisClient is a redis client that adds/remove prefixes in every operation where needed
// it also uses prefixedPipe for redis trasactions (serialized atomic operations)
type prefixedRedisClient struct {
	prefixable
	client *redis.Client
}

// newPrefixedRedisClient returns a new Prefixed Redis Client
func newPrefixedRedisClient(
	host string,
	port int,
	db int,
	password string,
	prefix string,
) *prefixedRedisClient {
	return &prefixedRedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
			DB:       db,
		}),
		prefixable: prefixable{prefix: prefix},
	}
}

// Get wraps aound redis get method by adding prefix and returning string and error directly
func (r *prefixedRedisClient) Get(key string) (string, error) {
	return r.client.Get(r.withPrefix(key)).Result()
}

// Set wraps around redis get method by adding prefix and returning error directly
func (r *prefixedRedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(r.withPrefix(key), value, expiration).Err()
}

// Keys wraps around redis keys method by adding prefix and returning []string and error directly
func (r *prefixedRedisClient) Keys(pattern string) ([]string, error) {
	keys, err := r.client.Keys(r.withPrefix(pattern)).Result()
	if err != nil {
		return nil, err
	}
	toRemove := fmt.Sprintf("%s.", r.prefix)
	for index, key := range keys {
		keys[index] = strings.Replace(key, toRemove, "", 1)
	}
	return keys, nil

}

// Del wraps around redis del method by adding prefix and returning int64 and error directly
func (r *prefixedRedisClient) Del(keys ...string) (int64, error) {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = r.withPrefix(k)
	}
	return r.client.Del(prefixedKeys...).Result()
}

// SMembers returns a slice with all the members of a set
func (r *prefixedRedisClient) SMembers(key string) ([]string, error) {
	return r.client.SMembers(r.withPrefix(key)).Result()
}

// SAdd adds new members to a set
func (r *prefixedRedisClient) SAdd(key string, members ...string) (int64, error) {
	return r.client.SAdd(r.withPrefix(key), members).Result()
}

// SRem removes members from a set
func (r *prefixedRedisClient) SRem(key string, members ...string) (int64, error) {
	return r.client.SRem(r.withPrefix(key), members).Result()
}

// Exists returns true if a key exists in redis
func (r *prefixedRedisClient) Exists(key string) (bool, error) {
	val, err := r.client.Exists(r.withPrefix(key)).Result()
	return (val == 1), err
}

// WrapTransaction accepts a function that performs a set of operations that will
// be serialized and executed atomically. The function passed will recive a prefixedPipe
func (r *prefixedRedisClient) WrapTransaction(f func(p *prefixedPipe)) error {
	pipe := r.client.TxPipeline()
	f(&prefixedPipe{pipe: pipe, prefixable: prefixable{prefix: r.prefix}})
	_, err := pipe.Exec()
	return err
}
