package redisdb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/splitio/go-client/splitio"
	"github.com/splitio/go-toolkit/logging"
	"github.com/splitio/go-toolkit/redis"
)

// RedisMetricsStorage is a redis-based implementation of split storage
type RedisMetricsStorage struct {
	client            redis.PrefixedRedisClient
	logger            logging.LoggerInterface
	gaugeTemplate     string
	countersTemplate  string
	latenciesTemplate string
	latenciesRegexp   *regexp.Regexp
}

// NewRedisMetricsStorage creates a new RedisSplitStorage and returns a reference to it
func NewRedisMetricsStorage(redisClient *redis.PrefixedRedisClient, metadata *splitio.SdkMetadata, logger logging.LoggerInterface) *RedisMetricsStorage {
	gaugeTemplate := strings.Replace(redisGauge, "{sdkVersion}", metadata.SDKVersion, 1)
	gaugeTemplate = strings.Replace(gaugeTemplate, "{instanceId}", metadata.MachineName, 1)
	countersTemplate := strings.Replace(redisCount, "{sdkVersion}", metadata.SDKVersion, 1)
	countersTemplate = strings.Replace(countersTemplate, "{instanceId}", metadata.MachineName, 1)
	latenciesTemplate := strings.Replace(redisLatency, "{sdkVersion}", metadata.SDKVersion, 1)
	latenciesTemplate = strings.Replace(latenciesTemplate, "{instanceId}", metadata.MachineName, 1)
	latencyRegex := regexp.MustCompile(redisLatencyRegex)
	return &RedisMetricsStorage{
		client:            *redisClient,
		logger:            logger,
		gaugeTemplate:     gaugeTemplate,
		countersTemplate:  countersTemplate,
		latenciesTemplate: latenciesTemplate,
		latenciesRegexp:   latencyRegex,
	}
}

// PutGauge stores a gauge in redis
func (r *RedisMetricsStorage) PutGauge(key string, gauge float64) {
	keyToStore := strings.Replace(r.gaugeTemplate, "{metric}", key, 1)
	err := r.client.Set(keyToStore, gauge, 0)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Error storing gauge \"%s\" in redis: %s\n", key, err))
	}
}

// IncLatency incraeses the latency of a bucket for a specific metric
func (r *RedisMetricsStorage) IncLatency(metric string, index int) {
	keyToIncr := strings.Replace(r.latenciesTemplate, "{metric}", metric, 1)
	keyToIncr = strings.Replace(keyToIncr, "{bucket}", strconv.FormatInt(int64(index), 10), 1)
	err := r.client.Incr(keyToIncr)
	if err != nil {
		r.logger.Error(fmt.Sprintf(
			"Error incrementing latency bucket %d for metric \"%s\" in redis: %s", index, metric, err.Error(),
		))
	}
}

// IncCounter incraeses the count for a specific metric
func (r *RedisMetricsStorage) IncCounter(metric string) {
	keyToIncr := strings.Replace(r.countersTemplate, "{metric}", metric, 1)
	err := r.client.Incr(keyToIncr)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Error incrementing counterfor metric \"%s\" in redis: %s", metric, err.Error()))
	}
}
