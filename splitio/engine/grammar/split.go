package grammar

import (
	"github.com/splitio/go-client/splitio/service/dtos"
)

// Split struct with added logic that wraps around a DTO
type Split struct {
	splitData  *dtos.SplitDTO
	conditions []Condition
}

// NewSplit instantiates a new Split object and all it's internal structures mapped to model classes
func NewSplit(splitDTO *dtos.SplitDTO) *Split {
	conditions := make([]Condition, 0)
	for _, cond := range splitDTO.Conditions {
		partitions := make([]Partition, 0)
		for _, part := range cond.Partitions {
			partitions = append(partitions, Partition{partitionData: &part})
		}
		matchers := make([]Matcher, 0)
		for _, matcher := range cond.MatcherGroup.Matchers {
			// TODO: CREATE A FUNCTION TO BUILD APPROPRIATE MATCHERS!
			matchers = append(matchers, Matcher{matcherData: &matcher})
		}
		conditions = append(conditions, Condition{
			combiner:      cond.MatcherGroup.Combiner,
			conditionData: &cond,
			matchers:      matchers,
			partitions:    partitions,
		})
	}
	return &Split{
		conditions: conditions,
		splitData:  splitDTO,
	}
}

// Name returns the name of the feature
func (s *Split) Name() string {
	return s.splitData.Name
}

// Seed returns the seed use for hashing
func (s *Split) Seed() int64 {
	return s.splitData.Seed
}

// Status returns whether the split is active or arhived
func (s *Split) Status() string {
	status := s.splitData.Status
	if status == "" || (status != SplitStatusActive && status != SplitStatusArchived) {
		return SplitStatusActive
	}
	return status
}

// Killed returns whether the split has been killed or not
func (s *Split) Killed() bool {
	return s.splitData.Killed
}

// DefaultTreatment returns the default treatment for the current split
func (s *Split) DefaultTreatment() string {
	return s.splitData.DefaultTreatment
}

// TrafficAllocation returns the traffic allocation configured for the current split
func (s *Split) TrafficAllocation() int {
	return s.splitData.TrafficAllocation
}

// TrafficAllocationSeed returns the seed for traffic allocation configured for this split
func (s *Split) TrafficAllocationSeed() int64 {
	return s.splitData.TrafficAllocationSeed
}

// Algo returns the hashing algorithm configured for this split
func (s *Split) Algo() int {
	algo := s.splitData.Algo
	if algo != SplitAlgoLegacy && algo != SplitAlgoMurmur {
		return SplitAlgoLegacy
	}
	return algo
}

// Conditions returns a slice of Condition objects
func (s *Split) Conditions() []Condition {
	return s.conditions
}
