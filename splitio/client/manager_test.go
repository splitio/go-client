package client

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
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
	if splitNames[0] != "split1" && splitNames[1] != "split2" {
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
	if s1.Name != all[0].Name || s1.ChangeNumber != all[0].ChangeNumber || s1.Killed != all[0].Killed || s1.TrafficType != all[0].TrafficType {
		t.Error("Split1 should be the same for Split() and Splits()")
	}

	if s2.Name != all[1].Name || s2.ChangeNumber != all[1].ChangeNumber || s2.Killed != all[1].Killed || s2.TrafficType != all[1].TrafficType {
		t.Error("Split2 should be the same for Split() and Splits()")
	}

	sx := manager.Split("split3492042")
	if sx != nil {
		t.Error("nonexistant split should return nil")
	}
}
