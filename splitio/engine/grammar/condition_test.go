package grammar

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/logging"
	"testing"
)

func TestConditionWrapperObject(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	condition1 := dtos.ConditionDTO{
		ConditionType: "WHITELIST",
		Label:         "Label1",
		MatcherGroup: dtos.MatcherGroupDTO{
			Combiner: "AND",
			Matchers: []dtos.MatcherDTO{
				dtos.MatcherDTO{
					Negate:      false,
					MatcherType: "ALL_KEYS",
					KeySelector: &dtos.KeySelectorDTO{
						Attribute:   nil,
						TrafficType: "something",
					},
				},
				dtos.MatcherDTO{
					Negate:      true,
					MatcherType: "ALL_KEYS",
					KeySelector: &dtos.KeySelectorDTO{
						Attribute:   nil,
						TrafficType: "something",
					},
				},
			},
		},
		Partitions: []dtos.PartitionDTO{
			dtos.PartitionDTO{
				Size:      75,
				Treatment: "on",
			},
			dtos.PartitionDTO{
				Size:      25,
				Treatment: "off",
			},
		},
	}

	wrapped := NewCondition(&condition1, nil, logger)

	if wrapped.Label() != "Label1" {
		t.Error("Label not set properly")
	}

	if wrapped.combiner != "AND" {
		t.Error("Combiner not set properly")
	}

	if len(wrapped.matchers) != len(condition1.MatcherGroup.Matchers) {
		t.Error("Incorrect number of matchers")
	}

	if wrapped.ConditionType() != ConditionTypeWhitelist {
		t.Error("Incorrect condition type")
	}

	treatment1 := wrapped.CalculateTreatment(50)
	if treatment1 != nil && *treatment1 != "on" {
		t.Error("CalculateTreatment returned incorrect treatment")
	}

	treatment2 := wrapped.CalculateTreatment(80)
	if treatment2 != nil && *treatment2 != "off" {
		t.Error("CalculateTreatment returned incorrect treatment")
	}
}

func TestAnotherWrapper(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	condition1 := dtos.ConditionDTO{
		ConditionType: "ROLLOUT",
		Label:         "Label2",
		MatcherGroup: dtos.MatcherGroupDTO{
			Combiner: "AND",
			Matchers: []dtos.MatcherDTO{
				dtos.MatcherDTO{
					Negate:      false,
					MatcherType: "ALL_KEYS",
					KeySelector: &dtos.KeySelectorDTO{
						Attribute:   nil,
						TrafficType: "something",
					},
				},
				dtos.MatcherDTO{
					Negate:      true,
					MatcherType: "ALL_KEYS",
					KeySelector: &dtos.KeySelectorDTO{
						Attribute:   nil,
						TrafficType: "something",
					},
				},
			},
		},
		Partitions: []dtos.PartitionDTO{
			dtos.PartitionDTO{
				Size:      75,
				Treatment: "on",
			},
			dtos.PartitionDTO{
				Size:      25,
				Treatment: "off",
			},
		},
	}

	wrapped := NewCondition(&condition1, nil, logger)

	if wrapped.Label() != "Label2" {
		t.Error("Label not set properly")
	}

	if wrapped.combiner != "AND" {
		t.Error("Combiner not set properly")
	}

	if len(wrapped.matchers) != len(condition1.MatcherGroup.Matchers) {
		t.Error("Incorrect number of matchers")
	}

	if wrapped.ConditionType() != ConditionTypeRollout {
		t.Error("Incorrect condition type")
	}

	treatment1 := wrapped.CalculateTreatment(50)
	if treatment1 != nil && *treatment1 != "on" {
		t.Error("CalculateTreatment returned incorrect treatment")
	}

	treatment2 := wrapped.CalculateTreatment(80)
	if treatment2 != nil && *treatment2 != "off" {
		t.Error("CalculateTreatment returned incorrect treatment")
	}
}
