package redisdb

import (
	"fmt"

	"github.com/splitio/go-client/splitio/conf"
	"github.com/splitio/go-toolkit/redis"
	"github.com/splitio/go-toolkit/redis/helpers"
)

// NewPrefixedRedisClient returns a new Prefixed Redis Client
func NewPrefixedRedisClient(config *conf.RedisConfig) (*redis.PrefixedRedisClient, error) {
	rClient, err := redis.NewClient(&redis.UniversalOptions{
		Addrs:     []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Password:  config.Password,
		DB:        config.Database,
		TLSConfig: config.TLSConfig,
	})

	if err != nil {
		panic(err.Error())
	}
	helpers.EnsureConnected(rClient)

	return redis.NewPrefixedRedisClient(rClient, config.Prefix)
}
