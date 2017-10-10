package grammar

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/injection"

	"fmt"
)

// Condition struct with added logic that wraps around a DTO
type Condition struct {
	conditionData *dtos.ConditionDTO
	matchers      []MatcherInterface
	combiner      string
	partitions    []Partition
}

// NewCondition instantiates a new Condition struct with appropriate wrappers around dtos and returns it.
func NewCondition(cond *dtos.ConditionDTO, ctx *injection.Context) *Condition {
	partitions := make([]Partition, 0)
	for _, part := range cond.Partitions {
		fmt.Println("Condition:", part)
		partitions = append(partitions, Partition{partitionData: part})
	}
	matchers := make([]MatcherInterface, 0)
	for _, matcher := range cond.MatcherGroup.Matchers {
		m, err := BuildMatcher(&matcher, ctx)
		if err == nil {
			matchers = append(matchers, m)
		}
	}

	return &Condition{
		combiner:      cond.MatcherGroup.Combiner,
		conditionData: cond,
		matchers:      matchers,
		partitions:    partitions,
	}
}

// Partition struct with added logic that wraps around a DTO
type Partition struct {
	partitionData dtos.PartitionDTO
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
	partial := make([]bool, len(c.matchers))
	for i, matcher := range c.matchers {
		partial[i] = matcher.Match(key, attributes, nil)
		if matcher.Negate() {
			partial[i] = !partial[i]
		}
	}
	return applyCombiner(partial, c.combiner)
}

// CalculateTreatment calulates the treatment for a specific condition based on the bucket
func (c *Condition) CalculateTreatment(bucket int) *string {
	accum := 0
	fmt.Println(c.partitions)
	for _, partition := range c.partitions {

		accum += partition.partitionData.Size
		if bucket <= accum {
			return &partition.partitionData.Treatment
		}
	}
	return nil
}

func applyCombiner(results []bool, combiner string) bool {
	temp := true
	switch combiner {
	case "AND":
		for _, result := range results {
			temp = temp && result
		}
	default:
		return false
	}
	return temp
}
