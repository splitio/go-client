package client

import (
	"testing"

	"github.com/splitio/go-split-commons/v7/dtos"
	"github.com/splitio/go-split-commons/v7/flagsets"
	"github.com/splitio/go-split-commons/v7/storage/inmemory/mutexmap"
	"github.com/splitio/go-toolkit/v5/datastructures/set"
	"github.com/splitio/go-toolkit/v5/logging"
)

func TestSplitManager(t *testing.T) {
	flagSetFilter := flagsets.NewFlagSetFilter([]string{})
	splitStorage := mutexmap.NewMMSplitStorage(flagSetFilter)
	splitStorage.Update([]dtos.SplitDTO{
		{
			ChangeNumber:     123,
			Name:             "split1",
			Killed:           false,
			TrafficTypeName:  "tt1",
			Sets:             []string{"set1", "set2"},
			DefaultTreatment: "s1p1",
			Conditions: []dtos.ConditionDTO{
				{
					Partitions: []dtos.PartitionDTO{
						{Treatment: "s1p1"},
						{Treatment: "s1p2"},
						{Treatment: "s1p3"},
					},
				},
			},
		},
		{
			ChangeNumber:    123,
			Name:            "split2",
			Killed:          true,
			TrafficTypeName: "tt2",
			Prerequisites: []dtos.Prerequisite{
				{
					FeatureFlagName: "ff1",
					Treatments: []string{
						"off",
						"v1",
					},
				},
			},
			Conditions: []dtos.ConditionDTO{
				{
					Partitions: []dtos.PartitionDTO{
						{Treatment: "s2p1"},
						{Treatment: "s2p2"},
						{Treatment: "s2p3"},
					},
				},
			},
		},
	}, nil, 123)

	logger := logging.NewLogger(nil)
	factory := SplitFactory{}
	manager := SplitManager{
		splitStorage: splitStorage,
		validator:    inputValidation{logger: logger},
		logger:       logger,
		factory:      &factory,
	}

	factory.status.Store(sdkStatusReady)

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

	if len(s1.Sets) != 2 {
		t.Error("split1 should have 2 sets")
	}

	if s1.DefaultTreatment != "s1p1" {
		t.Error("the default treatment for split1 should be s1p1")
	}

	if s1.ImpressionsDisabled {
		t.Error("track impressions for split1 should be false")
	}

	if s1.Prerequisites != nil {
		t.Error("prerequisistes should be nil for s1")
	}

	s2 := manager.Split("split2")
	if s2.Name != "split2" || !s2.Killed || s2.TrafficType != "tt2" || s2.ChangeNumber != 123 {
		t.Error("Split 2 stored incorrectly")
	}
	if s2.Treatments[0] != "s1p2" && s2.Treatments[1] != "s2p2" && s2.Treatments[2] != "s2p3" {
		t.Error("Incorrect treatments for split 2")
	}

	if s2.Sets == nil && len(s2.Sets) != 0 {
		t.Error("split2 sets should be empty array")
	}

	if s2.ImpressionsDisabled {
		t.Error("track impressions for split2 should be false")
	}

	if len(s2.Prerequisites) != 1 {
		t.Error("prerequisites size should be 1")
	}

	all := manager.Splits()
	if len(all) != 2 {
		t.Error("Incorrect number of splits returned")
	}

	sx := manager.Split("split3492042")
	if sx != nil {
		t.Error("Nonexistent split should return nil")
	}
}

func TestSplitManagerWithConfigs(t *testing.T) {
	flagSetFilter := flagsets.NewFlagSetFilter([]string{})
	splitStorage := mutexmap.NewMMSplitStorage(flagSetFilter)
	splitStorage.Update([]dtos.SplitDTO{*valid, *killed, *noConfig}, nil, 123)

	logger := logging.NewLogger(nil)
	factory := SplitFactory{}
	manager := SplitManager{
		splitStorage: splitStorage,
		logger:       logger,
		validator:    inputValidation{logger: logger},
		factory:      &factory,
	}

	factory.status.Store(sdkStatusReady)
	manager.factory = &factory

	splitNames := manager.SplitNames()
	splitNameSet := set.NewSet(splitNames[0], splitNames[1], splitNames[2])
	if !splitNameSet.IsEqual(set.NewSet("valid", "killed", "noConfig")) {
		t.Error("Incorrect split names returned")
	}

	s1 := manager.Split("valid")
	if s1.Name != "valid" || s1.Killed || s1.TrafficType != "user" || s1.ChangeNumber != 1494593336752 {
		t.Error("Split 1 stored incorrectly")
	}
	if s1.Treatments[0] != "on" {
		t.Error("Incorrect treatments for split 1")
	}
	if s1.Configs == nil {
		t.Error("It should have configs")
	}
	if s1.Configs["on"] != "{\"color\": \"blue\",\"size\": 13}" {
		t.Error("It should have configs")
	}
	if s1.DefaultTreatment != "off" {
		t.Error("the default treatment for valid should be off")
	}
	if s1.ImpressionsDisabled {
		t.Error("ImpressionsDisabled for valid should be false")
	}

	s2 := manager.Split("killed")
	if s2.Name != "killed" || !s2.Killed || s2.TrafficType != "user" || s2.ChangeNumber != 1494593336752 {
		t.Error("Split 2 stored incorrectly")
	}
	if s2.Treatments[0] != "off" {
		t.Error("Incorrect treatments for split 2")
	}
	if s2.Configs == nil {
		t.Error("It should have configs")
	}
	if s2.Configs["defTreatment"] != "{\"color\": \"orange\",\"size\": 15}" {
		t.Error("It should have configs")
	}
	if s2.DefaultTreatment != "defTreatment" {
		t.Error("the default treatment for killed should be defTreatment")
	}
	if s2.ImpressionsDisabled {
		t.Error("track impressions for killed should be false")
	}

	s3 := manager.Split("noConfig")
	if s3.Name != "noConfig" || s3.Killed || s3.TrafficType != "user" || s3.ChangeNumber != 1494593336752 {
		t.Error("Split 3 stored incorrectly")
	}
	if s3.Treatments[0] != "off" {
		t.Error("Incorrect treatments for split 3")
	}
	if s3.Configs != nil {
		t.Error("It should not have configs")
	}
	if s3.DefaultTreatment != "defTreatment" {
		t.Error("the default treatment for killed should be defTreatment")
	}
	if s3.ImpressionsDisabled {
		t.Error("track impressions for noConfig should be false")
	}

	all := manager.Splits()
	if len(all) != 3 {
		t.Error("Incorrect number of splits returned")
	}

	sx := manager.Split("split3492042")
	if sx != nil {
		t.Error("Nonexistent split should return nil")
	}
}

func TestSplitManagerTrackImpressions(t *testing.T) {
	flagSetFilter := flagsets.NewFlagSetFilter([]string{})
	splitStorage := mutexmap.NewMMSplitStorage(flagSetFilter)
	valid.ImpressionsDisabled = true
	noConfig.ImpressionsDisabled = true
	splitStorage.Update([]dtos.SplitDTO{*valid, *killed, *noConfig}, nil, 123)

	logger := logging.NewLogger(nil)
	factory := SplitFactory{}
	manager := SplitManager{
		splitStorage: splitStorage,
		logger:       logger,
		validator:    inputValidation{logger: logger},
		factory:      &factory,
	}

	factory.status.Store(sdkStatusReady)
	manager.factory = &factory

	s1 := manager.Split("valid")
	if !s1.ImpressionsDisabled {
		t.Error("track impressions for valid should be true")
	}

	s2 := manager.Split("killed")
	if s2.ImpressionsDisabled {
		t.Error("track impressions for killed should be false")
	}

	s3 := manager.Split("noConfig")
	if !s3.ImpressionsDisabled {
		t.Error("track impressions for noConfig should be true")
	}
}
