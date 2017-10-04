package grammar

import (
	"github.com/splitio/go-client/splitio/service/dtos"
)

// Condition struct with added logic that wraps around a DTO
type Condition struct {
	conditionData *dtos.ConditionDTO
	matchers      []Matcher
	combiner      string
	partitions    []Partition
}

// Partition struct with added logic that wraps around a DTO
type Partition struct {
	partitionData *dtos.PartitionDTO
}

// Matcher struct with added logic that wraps around a DTO
type Matcher struct {
	matcherData *dtos.MatcherDTO
}

// ConditionType returns validated condition type. Whitelist by default
func (c *Condition) ConditionType() string {
	cond := c.conditionData.ConditionType
	if cond != ConditionTypeWhitelist && cond != ConditionTypeRollout {
		return ConditionTypeWhitelist
	}
	return cond
}

// Label returns the condition's label
func (c *Condition) Label() string {
	return c.conditionData.Label
}

// Matches returns true if the condition matches for a specific key and/or set of attributes
func (c *Condition) Matches(key string, attributes map[string]interface{}) bool {
	// TODO
	return false
}

// CalculateTreatment calulates the treatment for a specific condition based on the bucket
func (c *Condition) CalculateTreatment(bucket int) *string {
	accum := 0
	for _, partition := range c.partitions {
		accum += partition.partitionData.Size
		if bucket <= accum {
			return &partition.partitionData.Treatment
		}
	}
	return nil
}
