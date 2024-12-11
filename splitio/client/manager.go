package client

import (
	"fmt"

	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/go-split-commons/v6/storage"
	"github.com/splitio/go-toolkit/v5/logging"
)

// SplitManager provides information of the currently stored splits
type SplitManager struct {
	splitStorage  storage.SplitStorageConsumer
	validator     inputValidation
	logger        logging.LoggerInterface
	factory       *SplitFactory
	initTelemetry storage.TelemetryConfigProducer
}

// SplitView is a partial representation of a currently stored split
type SplitView struct {
	Name             string            `json:"name"`
	TrafficType      string            `json:"trafficType"`
	Killed           bool              `json:"killed"`
	Treatments       []string          `json:"treatments"`
	ChangeNumber     int64             `json:"changeNumber"`
	Configs          map[string]string `json:"configs"`
	DefaultTreatment string            `json:"defaultTreatment"`
	Sets             []string          `json:"sets"`
	TrackImpressions bool              `json:"trackImpressions"`
}

func newSplitView(splitDto *dtos.SplitDTO) *SplitView {
	treatments := make([]string, 0)
	for _, condition := range splitDto.Conditions {
		for _, partition := range condition.Partitions {
			treatments = append(treatments, partition.Treatment)
		}
	}
	sets := []string{}
	if splitDto.Sets != nil {
		sets = splitDto.Sets
	}
	return &SplitView{
		ChangeNumber:     splitDto.ChangeNumber,
		Killed:           splitDto.Killed,
		Name:             splitDto.Name,
		TrafficType:      splitDto.TrafficTypeName,
		Treatments:       treatments,
		Configs:          splitDto.Configurations,
		DefaultTreatment: splitDto.DefaultTreatment,
		Sets:             sets,
		TrackImpressions: splitDto.TrackImpressions == nil || *splitDto.TrackImpressions,
	}
}

// SplitNames returns a list with the name of all the currently stored feature flags
func (m *SplitManager) SplitNames() []string {
	if m.isDestroyed() {
		m.logger.Error("Client has already been destroyed - no calls possible")
		return []string{}
	}

	if !m.isReady() {
		m.logger.Warning("SplitNames: the SDK is not ready, results may be incorrect. Make sure to wait for SDK readiness before using this method")
		m.initTelemetry.RecordNonReadyUsage()
	}

	return m.splitStorage.SplitNames()
}

// Splits returns a list of a partial view of every currently stored feature flag
func (m *SplitManager) Splits() []SplitView {
	if m.isDestroyed() {
		m.logger.Error("Client has already been destroyed - no calls possible")
		return []SplitView{}
	}

	if !m.isReady() {
		m.logger.Warning("Splits: the SDK is not ready, results may be incorrect. Make sure to wait for SDK readiness before using this method")
		m.initTelemetry.RecordNonReadyUsage()
	}

	splitViews := make([]SplitView, 0)
	splits := m.splitStorage.All()
	for _, split := range splits {
		splitViews = append(splitViews, *newSplitView(&split))
	}
	return splitViews
}

// Split returns a partial view of a particular feature flag
func (m *SplitManager) Split(featureFlagName string) *SplitView {
	if m.isDestroyed() {
		m.logger.Error("Client has already been destroyed - no calls possible")
		return nil
	}

	if !m.isReady() {
		m.logger.Warning("Split: the SDK is not ready, results may be incorrect. Make sure to wait for SDK readiness before using this method")
		m.initTelemetry.RecordNonReadyUsage()
	}

	err := m.validator.ValidateManagerInputs(featureFlagName)
	if err != nil {
		m.logger.Error(err.Error())
		return nil
	}

	split := m.splitStorage.Split(featureFlagName)
	if split != nil {
		return newSplitView(split)
	}
	m.logger.Error(fmt.Sprintf("Split: you passed %s that does not exist in this environment, please double check what feature flags exist in the Split user interface.", featureFlagName))
	return nil
}

// BlockUntilReady Calls BlockUntilReady on factory to block manager on readiness
func (m *SplitManager) BlockUntilReady(timer int) error {
	return m.factory.BlockUntilReady(timer)
}

func (m *SplitManager) isDestroyed() bool {
	return m.factory.IsDestroyed()
}

func (m *SplitManager) isReady() bool {
	return m.factory.IsReady()
}
