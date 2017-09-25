package storage

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"sync"
)

type MMSplitStorage struct {
	data  map[string]dtos.SplitDTO
	mutex *sync.Mutex
}

func NewMMSplitStorage() *MMSplitStorage {
	return &MMSplitStorage{
		data:  make(map[string]dtos.SplitDTO),
		mutex: &sync.Mutex{},
	}
}

func (m *MMSplitStorage) Get(splitName string) (*dtos.SplitDTO, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	item, exists = m.data[splitName]
	return &item, exists
}

func (m *MMSplitStorage) PutMany(splits *[]dtos.SplitDTO) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, split := range splits {
		data[split.Name] = split
	}
}

func (m *MMSplitStorage) Remove(splitName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	idelete(m.data, splitName)
}
