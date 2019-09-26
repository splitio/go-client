package redisdb

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/splitio/go-client/splitio/conf"
)

// prefixable is a struct intended to be embedded in anything that can have a prefix added.
// this currently includes a redis client and a redis transaction.
type prefixable struct {
	prefix string
}

// withPrefix adds a prefix to the key if the prefix supplied has a length greater than 0
func (p *prefixable) withPrefix(key string) string {
	if len(p.prefix) > 0 {
		return fmt.Sprintf("%s.%s", p.prefix, key)
	}
	return key
}

// withoutPrefix removes the prefix from a key if the prefix has a length greater than 0
func (p *prefixable) withoutPrefix(key string) string {
	if len(p.prefix) > 0 {
		return strings.Replace(key, fmt.Sprintf("%s.", p.prefix), "", 1)
	}
	return key
}

type prefixedTx struct {
	prefixable
	tx *redis.Tx
}

// wrap redis "set" operation with a prefix inside a transaction
func (t *prefixedTx) Set(key string, value interface{}, expiration time.Duration) error {
	return t.tx.Set(t.withPrefix(key), value, expiration).Err()
}

// wrap redis "sadd" operation with a prefix inside a transaction. returns a future-like result
func (t *prefixedTx) SAdd(key string, members ...interface{}) error {
	return t.tx.SAdd(t.withPrefix(key), members...).Err()
}

// wrap redis "del" operation with a prefix inside a transaction
func (t *prefixedTx) Del(keys ...string) error {
	prefixed := make([]string, 0)
	for _, key := range keys {
		prefixed = append(prefixed, t.withPrefix(key))
	}

	return t.tx.Del(prefixed...).Err()
}

// Wraps redis "smembers" operation with a prefix inside a transaction
func (t *prefixedTx) Smembers(key string) ([]string, error) {
	res := t.tx.SMembers(t.withPrefix(key))
	return res.Val(), res.Err()
}

// Keys wraps redis "keys" operation with a prefix inside a transaction
func (t *prefixedTx) Keys(pattern string) ([]string, error) {
	res := t.tx.Keys(t.withPrefix(pattern))
	woPrefix := make([]string, len(res.Val()))
	for index, key := range res.Val() {
		woPrefix[index] = t.withoutPrefix(key)
	}
	return woPrefix, res.Err()
}

func (t *prefixedTx) Get(key string) (string, error) {
	res := t.tx.Get(t.withPrefix(key))
	return res.Val(), res.Err()
}

// newPrefixedPipe instantiates a new pipewrapper and returns a reference
func newPrefixedTx(tx *redis.Tx, prefix string) *prefixedTx {
	return &prefixedTx{
		prefixable: prefixable{prefix: prefix},
		tx:         tx,
	}
}

// ---------

// PrefixedRedisClient is a redis client that adds/remove prefixes in every operation where needed
// it also uses prefixedPipe for redis trasactions (serialized atomic operations)
type PrefixedRedisClient struct {
	prefixable
	client *redis.Client
}

// NewPrefixedRedisClient returns a new Prefixed Redis Client
func NewPrefixedRedisClient(config *conf.RedisConfig) (*PrefixedRedisClient, error) {
	rClient := redis.NewClient(&redis.Options{
		Addr:      fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:  config.Password,
		DB:        config.Database,
		TLSConfig: config.TLSConfig,
	})

	err := rClient.Ping().Err()
	if err != nil {
		return nil, err
	}

	return &PrefixedRedisClient{
		client:     rClient,
		prefixable: prefixable{prefix: config.Prefix},
	}, nil
}

// Get wraps aound redis get method by adding prefix and returning string and error directly
func (r *PrefixedRedisClient) Get(key string) (string, error) {
	return r.client.Get(r.withPrefix(key)).Result()
}

// Set wraps around redis get method by adding prefix and returning error directly
func (r *PrefixedRedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(r.withPrefix(key), value, expiration).Err()
}

// Keys wraps around redis keys method by adding prefix and returning []string and error directly
func (r *PrefixedRedisClient) Keys(pattern string) ([]string, error) {
	keys, err := r.client.Keys(r.withPrefix(pattern)).Result()
	if err != nil {
		return nil, err
	}

	woPrefix := make([]string, len(keys))
	for index, key := range keys {
		woPrefix[index] = r.withoutPrefix(key)
	}
	return woPrefix, nil

}

// Del wraps around redis del method by adding prefix and returning int64 and error directly
func (r *PrefixedRedisClient) Del(keys ...string) (int64, error) {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = r.withPrefix(k)
	}
	return r.client.Del(prefixedKeys...).Result()
}

// SMembers returns a slice with all the members of a set
func (r *PrefixedRedisClient) SMembers(key string) ([]string, error) {
	return r.client.SMembers(r.withPrefix(key)).Result()
}

// SAdd adds new members to a set
func (r *PrefixedRedisClient) SAdd(key string, members ...interface{}) (int64, error) {
	return r.client.SAdd(r.withPrefix(key), members...).Result()
}

// SRem removes members from a set
func (r *PrefixedRedisClient) SRem(key string, members ...string) (int64, error) {
	return r.client.SRem(r.withPrefix(key), members).Result()
}

// Exists returns true if a key exists in redis
func (r *PrefixedRedisClient) Exists(key string) (bool, error) {
	val, err := r.client.Exists(r.withPrefix(key)).Result()
	return (val == 1), err
}

// Incr increments a key. Sets it in one if it doesn't exist
func (r *PrefixedRedisClient) Incr(key string) error {
	return r.client.Incr(r.withPrefix(key)).Err()
}

// WrapTransaction accepts a function that performs a set of operations that will
// be serialized and executed atomically. The function passed will recive a prefixedPipe
func (r *PrefixedRedisClient) WrapTransaction(f func(t *prefixedTx) error) error {
	return r.client.Watch(func(tx *redis.Tx) error {
		return f(newPrefixedTx(tx, r.prefix))
	})
}

// RPush insert all the specified values at the tail of the list stored at key
func (r *PrefixedRedisClient) RPush(key string, values ...interface{}) (int64, error) {
	return r.client.RPush(r.withPrefix(key), values...).Result()
}

// LRange Returns the specified elements of the list stored at key
func (r *PrefixedRedisClient) LRange(key string, start, stop int64) *redis.StringSliceCmd {
	return r.client.LRange(r.withPrefix(key), start, stop)
}

// LTrim Trim an existing list so that it will contain only the specified range of elements specified
func (r *PrefixedRedisClient) LTrim(key string, start, stop int64) *redis.StatusCmd {
	return r.client.LTrim(r.withPrefix(key), start, stop)
}

// LLen Returns the length of the list stored at key
func (r *PrefixedRedisClient) LLen(key string) *redis.IntCmd {
	return r.client.LLen(r.withPrefix(key))
}

// Expire set expiration time for particular key
func (r *PrefixedRedisClient) Expire(key string, value time.Duration) *redis.BoolCmd {
	return r.client.Expire(r.withPrefix(key), value)
}

// TTL for particular key
func (r *PrefixedRedisClient) TTL(key string) *redis.DurationCmd {
	return r.client.TTL(r.withPrefix(key))
}

// Mget fetchs multiple results
func (r *PrefixedRedisClient) Mget(keys []string) ([]interface{}, error) {
	keysWithPrefix := make([]string, 0)
	for _, key := range keys {
		keysWithPrefix = append(keysWithPrefix, r.withPrefix(key))
	}
	return r.client.MGet(keysWithPrefix...).Result()
}
