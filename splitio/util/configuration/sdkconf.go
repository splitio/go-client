// Package configuration ...
// Contains configuration structures used to setup the SDK
package configuration

import (
	"errors"
	"fmt"
	"github.com/splitio/go-client/splitio/util/impressionlistener"
	"github.com/splitio/go-toolkit/datastructures/set"
	"github.com/splitio/go-toolkit/logging"
	"github.com/splitio/go-toolkit/nethelpers"
	"math"
	"os/user"
	"path"
	"strings"
)

// SplitSdkConfig struct ...
// struct used to setup a Split.io SDK client.
//
// Parameters:
// - Apikey: (Required) API-KEY used to authenticate user requests
// - OperationMode (Required) Must be one of ["inmemory-standalone", "redis-consumer", "redis-standalone"]
// - InstanceName (Optional) Name to be used when submitting metrics & impressions to split servers
// - Logger: (Optional) Custom logger complying with logging.LoggerInterface
// - LoggerConfig: (Optional) Options to setup the sdk's own logger
// - TaskPeriods: (Optional) How often should each task run
// - Redis: (Required for "redis-consumer" & "redis-standalone" operation modes. Sets up Redis config
// - Advanced: (Optional) Sets up various advanced options for the sdk
type SplitSdkConfig struct {
	Apikey          string
	OperationMode   string
	InstanceName    string
	IPAddress       string
	BlockUntilReady int
	SplitFile       string
	Logger          logging.LoggerInterface
	LoggerConfig    *logging.LoggerOptions
	TaskPeriods     *TaskPeriods
	Advanced        *AdvancedConfig
	Redis           *RedisConfig
}

// TaskPeriods struct is used to configure the period for each synchronization task
type TaskPeriods struct {
	SplitSync      int
	SegmentSync    int
	ImpressionSync int
	GaugeSync      int
	CounterSync    int
	LatencySync    int
}

// RedisConfig struct is used to cofigure the redis parameters
type RedisConfig struct {
	Host     string
	Port     int
	Database int
	Password string
	Prefix   string
}

// AdvancedConfig exposes more configurable parameters that can be used to further tailor the sdk to the user's needs
type AdvancedConfig struct {
	ImpressionListener impressionlistener.ListenerInterface
	HTTPTimeout        int
	SdkURL             string
	EventsURL          string
	SegmentQueueSize   int
	SegmentWorkers     int
}

func (c *SplitSdkConfig) normalizeIPAndInstanceID() {
	if c.IPAddress == "" {
		var err error
		c.IPAddress, err = nethelpers.ExternalIP()
		if err != nil {
			c.IPAddress = "unknown"
		}
	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("ip-%s", strings.Replace(c.IPAddress, ".", "-", -1))
	}
}

func (c *SplitSdkConfig) normalizeRedis() {
	if c.Redis == nil {
		c.Redis = &RedisConfig{}
	}

	if c.Redis.Host == "" {
		c.Redis.Host = defaultRedisHost
	}

	if c.Redis.Port == 0 {
		c.Redis.Port = defaultRedisPort
	}
}

func (c *SplitSdkConfig) normalizePeriods() {
	if c.TaskPeriods == nil {
		c.TaskPeriods = &TaskPeriods{}
	}

	if c.TaskPeriods.CounterSync == 0 {
		c.TaskPeriods.CounterSync = defaultTaskPeriod
	}

	if c.TaskPeriods.GaugeSync == 0 {
		c.TaskPeriods.GaugeSync = defaultTaskPeriod
	}

	if c.TaskPeriods.ImpressionSync == 0 {
		c.TaskPeriods.ImpressionSync = defaultTaskPeriod
	}

	if c.TaskPeriods.LatencySync == 0 {
		c.TaskPeriods.LatencySync = defaultTaskPeriod
	}

	if c.TaskPeriods.SegmentSync == 0 {
		c.TaskPeriods.SegmentSync = defaultTaskPeriod
	}

	if c.TaskPeriods.SplitSync == 0 {
		c.TaskPeriods.SplitSync = defaultTaskPeriod
	}
}

func (c *SplitSdkConfig) normalizeAdvancedConfig() {
	if c.Advanced == nil {
		c.Advanced = &AdvancedConfig{}
	}

	// NOTE: HTTPTimeout, sdkUrl & eventsURL are set in http client in service/api.
	if c.Advanced.SegmentQueueSize == 0 {
		c.Advanced.SegmentQueueSize = 500
	}

	// If the user did not specify a number of segment workers, use one for every 25 segments
	// The math.Max is used in case the queue size is less than 25, otherwise it would result
	// in zero segment workers
	if c.Advanced.SegmentWorkers == 0 {
		c.Advanced.SegmentWorkers = int(math.Max(float64(c.Advanced.SegmentQueueSize/25), 1))
	}
}

// Normalize checks for unset parameters and sets them to the default value.
// If required parameters are missing returns an error.
func (c *SplitSdkConfig) Normalize() error {

	// To keep the interface consistent with other sdks we accept "localhost" as an apikey,
	// which sets the operation mode to localhost
	if c.Apikey == "localhost" {
		c.OperationMode = "localhost"
	}

	// Default to inmemory-standalone if no operation mode is provided
	if c.OperationMode == "" {
		c.OperationMode = "inmemory-standalone"
	}

	// Fail if an invalid operation-mode is provided
	operationModes := set.NewSet(
		"localhost",
		"inmemory-standalone",
		"redis-consumer",
		"redis-standalone",
	)

	if !operationModes.Has(c.OperationMode) {
		return fmt.Errorf("OperationMode parameter must be one of: %v", operationModes.List())
	}

	// Fail if no apikey is provided
	if c.Apikey == "" && c.OperationMode != "localhost" {
		return errors.New("Config parameter \"Apikey\" is mandatory for operation modes other than localhost")
	}

	// Set Block until ready to default value if not provided
	if c.BlockUntilReady == 0 {
		c.BlockUntilReady = defaultBlockUntilReady
	}

	// Normalize IP and instance ID
	c.normalizeIPAndInstanceID()

	// If localhost mode selected and no splitFile provided, try to determine user's home dir
	// fail if this cannot be done
	if c.OperationMode == "localhost" && c.SplitFile == "" {
		usr, err := user.Current()
		if err != nil {
			return errors.New(
				"Localhost mode selected. No split file specified and cannot determine user's home dir",
			)
		}
		c.SplitFile = path.Join(usr.HomeDir, ".splits")
	}

	// Normalize Redis config if needed
	if c.OperationMode == "redis-consumer" || c.OperationMode == "redis-standalone" {
		c.normalizeRedis()
	}

	// Normalize task periods
	c.normalizePeriods()

	// Normalize advanced config
	c.normalizeAdvancedConfig()

	return nil
}
