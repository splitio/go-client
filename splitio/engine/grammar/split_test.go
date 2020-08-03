package grammar

import (
	"testing"

	"github.com/splitio/go-split-commons/dtos"
	"github.com/splitio/go-toolkit/logging"
)

func TestSplitCreation(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	dto := dtos.SplitDTO{
		Algo:                  1,
		ChangeNumber:          123,
		Conditions:            []dtos.ConditionDTO{},
		DefaultTreatment:      "def",
		Killed:                false,
		Name:                  "split1",
		Seed:                  1234,
		Status:                "ACTIVE",
		TrafficAllocation:     100,
		TrafficAllocationSeed: 333,
		TrafficTypeName:       "tt1",
	}
	split := NewSplit(&dto, nil, logger)

	if split.Algo() != SplitAlgoLegacy {
		t.Error("Algo() should return legacy")
	}

	if split.ChangeNumber() != 123 {
		t.Error("Incorrect changenumber returned")
	}

	if split.DefaultTreatment() != "def" {
		t.Error("Incorrect default treatment")
	}

	if split.Killed() {
		t.Error("Split should not be marked as killed")
	}

	if split.Seed() != 1234 {
		t.Error("Incorrect seed")
	}

	if split.Name() != "split1" {
		t.Error("Incorrect split name")
	}

	if split.Status() != SplitStatusActive {
		t.Error("Status should be active")
	}

	if split.TrafficAllocation() != 100 {
		t.Error("Traffic allocation should be 100")
	}
}
