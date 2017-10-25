package client

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-client/splitio/storage"
)

// Manager provides information of the currently stored splits
type Manager struct {
	splitStorage storage.SplitStorage
}

// SplitView is a partial representation of a currently stored split
type SplitView struct {
	Name         string
	TrafficType  string
	Killed       bool
	Treatments   []string
	ChangeNumber int64
}

func newSplitView(splitDto *dtos.SplitDTO) *SplitView {
	treatments := make([]string, 0)
	for _, condition := range splitDto.Conditions {
		for _, partition := range condition.Partitions {
			treatments = append(treatments, partition.Treatment)
		}
	}
	return &SplitView{
		ChangeNumber: splitDto.ChangeNumber,
		Killed:       splitDto.Killed,
		Name:         splitDto.Name,
		TrafficType:  splitDto.TrafficTypeName,
		Treatments:   treatments,
	}
}

// SplitNames returns a list with the name of all the currently stored splits
func (m *Manager) SplitNames() []string {
	return m.splitStorage.SplitNames()
}

// Splits returns a list of a partial view of every currently stored split
func (m *Manager) Splits() []SplitView {
	splitViews := make([]SplitView, 0)
	splits := m.splitStorage.GetAll()
	for _, split := range splits {
		splitViews = append(splitViews, *newSplitView(&split))
	}
	return splitViews
}

// Split returns a partial view of a particular split
func (m *Manager) Split(feature string) *SplitView {
	split := m.splitStorage.Get(feature)
	if split != nil {
		return newSplitView(split)
	}
	return nil
}
