package client

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-toolkit/datastructures/set"
	"testing"
)

func TestSplitManager(t *testing.T) {
	splitStorage := storage.NewMMSplitStorage()
	splitStorage.PutMany([]dtos.SplitDTO{
		dtos.SplitDTO{
			ChangeNumber:    123,
			Name:            "split1",
			Killed:          false,
			TrafficTypeName: "tt1",
			Conditions: []dtos.ConditionDTO{
				dtos.ConditionDTO{
					Partitions: []dtos.PartitionDTO{
						dtos.PartitionDTO{Treatment: "s1p1"},
						dtos.PartitionDTO{Treatment: "s1p2"},
						dtos.PartitionDTO{Treatment: "s1p3"},
					},
				},
			},
		},
		dtos.SplitDTO{
			ChangeNumber:    123,
			Name:            "split2",
			Killed:          true,
			TrafficTypeName: "tt2",
			Conditions: []dtos.ConditionDTO{
				dtos.ConditionDTO{
					Partitions: []dtos.PartitionDTO{
						dtos.PartitionDTO{Treatment: "s2p1"},
						dtos.PartitionDTO{Treatment: "s2p2"},
						dtos.PartitionDTO{Treatment: "s2p3"},
					},
				},
			},
		},
	}, 123)

	manager := Manager{splitStorage: splitStorage}

	splitNames := manager.SplitNames()
	splitNameSet := set.NewSet(splitNames[0], splitNames[1])
	if !splitNameSet.IsEqual(set.NewSet("split1", "split2")) {
		t.Error("Incorrect split names returned")
	}

	s1 := manager.Split("split1")
	if s1.Name != "split1" || s1.Killed || s1.TrafficType != "tt1" || s1.ChangeNumber != 123 {
		t.Error("Split 1 stored incorrectly")
	}
	if s1.Treatments[0] != "s1p1" && s1.Treatments[1] != "s1p2" && s1.Treatments[2] != "s1p3" {
		t.Error("Incorrect treatments for split 1")
	}

	s2 := manager.Split("split2")
	if s2.Name != "split2" || !s2.Killed || s2.TrafficType != "tt2" || s2.ChangeNumber != 123 {
		t.Error("Split 2 stored incorrectly")
	}
	if s2.Treatments[0] != "s1p2" && s2.Treatments[1] != "s2p2" && s2.Treatments[2] != "s2p3" {
		t.Error("Incorrect treatments for split 2")
	}

	all := manager.Splits()
	if len(all) != 2 {
		t.Error("Incorrect number of splits returned")
	}

	sx := manager.Split("split3492042")
	if sx != nil {
		t.Error("Nonexistant split should return nil")
	}
}
