package client

import (
	"testing"

	"github.com/splitio/go-split-commons/v9/dtos"
	"github.com/splitio/go-split-commons/v9/flagsets"
	"github.com/splitio/go-split-commons/v9/storage/inmemory/mutexmap"
	"github.com/splitio/go-toolkit/v5/logging"

	"github.com/stretchr/testify/assert"
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
	assert.ElementsMatch(t, []string{"split1", "split2"}, splitNames)

	s1 := manager.Split("split1")
	assert.Equal(t, "split1", s1.Name)
	assert.False(t, s1.Killed)
	assert.Equal(t, "tt1", s1.TrafficType)
	assert.Equal(t, int64(123), s1.ChangeNumber)
	assert.ElementsMatch(t, []string{"s1p1", "s1p2", "s1p3"}, s1.Treatments)
	assert.ElementsMatch(t, []string{"set1", "set2"}, s1.Sets)
	assert.Equal(t, "s1p1", s1.DefaultTreatment)
	assert.False(t, s1.ImpressionsDisabled)
	assert.Nil(t, s1.Prerequisites)

	s2 := manager.Split("split2")
	assert.Equal(t, "split2", s2.Name)
	assert.True(t, s2.Killed)
	assert.Equal(t, "tt2", s2.TrafficType)
	assert.Equal(t, int64(123), s2.ChangeNumber)
	assert.ElementsMatch(t, []string{"s2p1", "s2p2", "s2p3"}, s2.Treatments)
	assert.ElementsMatch(t, []string{}, s2.Sets)
	assert.False(t, s2.ImpressionsDisabled)
	assert.Len(t, s2.Prerequisites, 1)

	all := manager.Splits()
	assert.Len(t, all, 2)

	assert.Nil(t, manager.Split("split3492042"))
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
	assert.ElementsMatch(t, []string{"valid", "killed", "noConfig"}, splitNames)

	s1 := manager.Split("valid")
	assert.Equal(t, "valid", s1.Name)
	assert.False(t, s1.Killed)
	assert.Equal(t, "user", s1.TrafficType)
	assert.Equal(t, int64(1494593336752), s1.ChangeNumber)
	assert.ElementsMatch(t, []string{"on"}, s1.Treatments)
	assert.NotNil(t, s1.Configs)
	assert.Equal(t, "{\"color\": \"blue\",\"size\": 13}", s1.Configs["on"])
	assert.Equal(t, "off", s1.DefaultTreatment)
	assert.False(t, s1.ImpressionsDisabled)

	s2 := manager.Split("killed")
	assert.Equal(t, "killed", s2.Name)
	assert.True(t, s2.Killed)
	assert.Equal(t, "user", s2.TrafficType)
	assert.Equal(t, int64(1494593336752), s2.ChangeNumber)
	assert.ElementsMatch(t, []string{"off"}, s2.Treatments)
	assert.NotNil(t, s2.Configs)
	assert.Equal(t, "{\"color\": \"orange\",\"size\": 15}", s2.Configs["defTreatment"])
	assert.ElementsMatch(t, []string{"off"}, s2.Treatments)
	assert.NotNil(t, s2.Configs)
	assert.Equal(t, "{\"color\": \"orange\",\"size\": 15}", s2.Configs["defTreatment"])
	assert.ElementsMatch(t, []string{"off"}, s2.Treatments)
	assert.NotNil(t, s2.Configs)
	assert.False(t, s2.ImpressionsDisabled)

	s3 := manager.Split("noConfig")
	assert.Equal(t, "noConfig", s3.Name)
	assert.False(t, s3.Killed)
	assert.Equal(t, "user", s3.TrafficType)
	assert.Equal(t, int64(1494593336752), s3.ChangeNumber)
	assert.ElementsMatch(t, []string{"off"}, s3.Treatments)
	assert.Nil(t, s3.Configs)
	assert.Equal(t, "defTreatment", s3.DefaultTreatment)
	assert.False(t, s3.ImpressionsDisabled)

	all := manager.Splits()
	assert.Len(t, all, 3)

	assert.Nil(t, manager.Split("split3492042"))
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
	assert.True(t, s1.ImpressionsDisabled)

	s2 := manager.Split("killed")
	assert.False(t, s2.ImpressionsDisabled)

	s3 := manager.Split("noConfig")
	assert.True(t, s3.ImpressionsDisabled)
}
