package redisdb

import (
	"fmt"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
	"regexp"
	"strconv"
	"strings"
)

// RedisMetricsStorage is a redis-based implementation of split storage
type RedisMetricsStorage struct {
	client            prefixedRedisClient
	logger            logging.LoggerInterface
	gaugeTemplate     string
	countersTemplate  string
	latenciesTemplate string
	latenciesRegexp   *regexp.Regexp
}

// NewRedisMetricsStorage creates a new RedisSplitStorage and returns a reference to it
func NewRedisMetricsStorage(
	host string,
	port int,
	db int,
	password string,
	prefix string,
	instanceID string,
	sdkVersion string,
	logger logging.LoggerInterface,
) *RedisMetricsStorage {
	gaugeTemplate := strings.Replace(redisGauge, "{sdkVersion}", sdkVersion, 1)
	gaugeTemplate = strings.Replace(gaugeTemplate, "{instanceId}", instanceID, 1)
	countersTemplate := strings.Replace(redisCount, "{sdkVersion}", sdkVersion, 1)
	countersTemplate = strings.Replace(countersTemplate, "{instanceId}", instanceID, 1)
	latenciesTemplate := strings.Replace(redisLatency, "{sdkVersion}", sdkVersion, 1)
	latenciesTemplate = strings.Replace(latenciesTemplate, "{instanceId}", instanceID, 1)
	latencyRegex := regexp.MustCompile(redisLatencyRegex)
	return &RedisMetricsStorage{
		client:            *newPrefixedRedisClient(host, port, db, password, prefix),
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

// PopGauges returns and clears all gauges in redis.
func (r *RedisMetricsStorage) PopGauges() []dtos.GaugeDTO {
	toRemove := strings.Replace(r.gaugeTemplate, "{metric}", "", 1) // String that will be removed from every key
	rawGauges := make(map[string]float64)
	err := r.client.WrapTransaction(func(t *prefixedTx) error {
		keys, err := t.Keys(strings.Replace(r.gaugeTemplate, "{metric}", "*", 1))
		if err != nil {
			r.logger.Error("Could not retrieve gauge keys from redis")
			return err
		}

		for _, key := range keys {
			gauge, err := t.Get(key)
			if err != nil {
				r.logger.Error(fmt.Sprintf("Could not retrieve gauges for key %s: %s", key, err))
				continue
			}

			asFloat, err := strconv.ParseFloat(gauge, 64)
			if err != nil {
				r.logger.Error("Error parsing gauge as float")
				continue
			}

			rawGauges[strings.Replace(key, toRemove, "", 1)] = asFloat
			t.Del(key)
		}
		return nil
	})
	if err != nil {
		r.logger.Error(err)
		return nil
	}

	all := make([]dtos.GaugeDTO, len(rawGauges))
	allIndex := 0
	for metric, gauge := range rawGauges {
		all[allIndex] = dtos.GaugeDTO{
			Gauge:      gauge,
			MetricName: metric,
		}
		allIndex++
	}
	return all
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

// PopLatencies returns and clears all gauges in redis.
func (r *RedisMetricsStorage) PopLatencies() []dtos.LatenciesDTO {
	latencies := make(map[string][]int64)
	err := r.client.WrapTransaction(func(t *prefixedTx) error {
		pattern := strings.Replace(r.latenciesTemplate, "{metric}", "*", 1)
		pattern = strings.Replace(pattern, "{bucket}", "*", 1)
		keys, err := t.Keys(pattern)
		if err != nil {
			r.logger.Error("Could not retrieve latency keys from redis")
			return err
		}

		for _, key := range keys {
			latency, err := t.Get(key)
			if err != nil {
				r.logger.Error(fmt.Sprintf("Could not retrieve latency for key %s: %s", key, err.Error()))
				continue
			}

			asInt64, err := strconv.ParseInt(latency, 10, 64)
			if err != nil {
				r.logger.Error("Error parsing latency to int")
				continue
			}

			// We use a regular expression to parse the key and retrieve the
			// metric name and bucket
			matches := r.latenciesRegexp.FindStringSubmatch(key)
			if len(matches) != 3 {
				r.logger.Error(fmt.Sprintf("Error parsing latency key %s", key))
				continue
			}
			metricName := matches[1]
			bucket, converr := strconv.ParseInt(matches[2], 10, 64)
			if converr != nil || bucket > 22 { // TODO: Change 22 to a constant!
				r.logger.Error(fmt.Sprintf("Invalid bucket %s in key %s", matches[2], key))
			}

			if _, has := latencies[metricName]; !has {
				latencies[metricName] = make([]int64, 23) // TODO move 23 to a constant!
			}
			latencies[string(metricName)][bucket] = asInt64
			t.Del(key)
		}
		return nil
	})
	if err != nil {
		r.logger.Error(err)
		return nil
	}

	all := make([]dtos.LatenciesDTO, len(latencies))
	allIndex := 0
	for metric, latencies := range latencies {
		all[allIndex] = dtos.LatenciesDTO{
			Latencies:  latencies,
			MetricName: metric,
		}
		allIndex++
	}
	return all
}

// IncCounter incraeses the count for a specific metric
func (r *RedisMetricsStorage) IncCounter(metric string) {
	keyToIncr := strings.Replace(r.countersTemplate, "{metric}", metric, 1)
	err := r.client.Incr(keyToIncr)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Error incrementing counterfor metric \"%s\" in redis: %s", metric, err.Error()))
	}
}

// PopCounters returns and clears all counters in redis.
func (r *RedisMetricsStorage) PopCounters() []dtos.CounterDTO {
	toRemove := strings.Replace(r.countersTemplate, "{metric}", "", 1) // String that will be removed from every key
	rawCounters := make(map[string]int64)
	err := r.client.WrapTransaction(func(t *prefixedTx) error {
		keys, err := t.Keys(strings.Replace(r.countersTemplate, "{metric}", "*", 1))
		if err != nil {
			r.logger.Error("Could not retrieve counter keys from redis")
			return err
		}

		for _, key := range keys {
			counter, err := t.Get(key)
			if err != nil {
				r.logger.Error(fmt.Sprintf("Could not retrieve counters for key %s: %s", key, err.Error()))
				continue
			}

			asInt, err := strconv.ParseInt(counter, 10, 64)
			if err != nil {
				r.logger.Error("Error parsing counter as int")
				continue
			}

			rawCounters[strings.Replace(key, toRemove, "", 1)] = asInt
			t.Del(key)
		}
		return nil
	})
	if err != nil {
		r.logger.Error(err)
		return nil
	}

	all := make([]dtos.CounterDTO, len(rawCounters))
	allIndex := 0
	for metric, counter := range rawCounters {
		all[allIndex] = dtos.CounterDTO{
			Count:      counter,
			MetricName: metric,
		}
		allIndex++
	}
	return all
}
