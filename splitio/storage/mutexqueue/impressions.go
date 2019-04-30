package mutexqueue

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/splitio/go-client/splitio/service/dtos"
)

// NewMQImpressionsStorage returns an instance of MQEventsStorage
func NewMQImpressionsStorage(queueSize int, isFull chan<- bool) *MQImpressionsStorage {
	return &MQImpressionsStorage{
		queue:      list.New(),
		size:       queueSize,
		mutexQueue: &sync.Mutex{},
		fullChan:   isFull,
	}
}

// MQImpressionsStorage in memory events storage
type MQImpressionsStorage struct {
	queue      *list.List
	size       int
	mutexQueue *sync.Mutex
	fullChan   chan<- bool //only write channel
}

func (s *MQImpressionsStorage) sendSignalIsFull() {
	// Nom blocking select
	select {
	case s.fullChan <- true:
		//Send "queue is full" signal
		break
	default:
		break
	}
}

// LogImpressions inserts impressions into the queue
func (s *MQImpressionsStorage) LogImpressions(impressions []dtos.ImpressionDTO) error {
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	for _, impression := range impressions {
		s.Push(impression)
	}

	return nil
}

// Push an event into slice
func (s *MQImpressionsStorage) Push(impression dtos.ImpressionDTO) error {
	fmt.Println("IS Push")
	if s.queue.Len()+1 > s.size {
		s.sendSignalIsFull()
		return ErrorMaxSizeReached
	}

	// Add element
	s.queue.PushBack(impression)

	if s.queue.Len() == s.size {
		s.sendSignalIsFull()
	}

	return nil
}

// PopN pop N elements from queue
func (s *MQImpressionsStorage) PopN(n int64) ([]dtos.ImpressionDTO, error) {
	var toReturn []dtos.ImpressionDTO
	var totalItems int

	// Mutexing queue
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	if int64(s.queue.Len()) >= n {
		totalItems = int(n)
	} else {
		totalItems = s.queue.Len()
	}

	toReturn = make([]dtos.ImpressionDTO, 0)
	for i := 0; i < totalItems; i++ {
		toReturn = append(toReturn, s.queue.Remove(s.queue.Front()).(dtos.ImpressionDTO))
	}

	return toReturn, nil
}

// Empty returns if slice len if zero
func (s *MQImpressionsStorage) Empty() bool {
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	return s.queue.Len() == 0
}

// Count returns the number of events into slice
func (s *MQImpressionsStorage) Count() int64 {
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	return int64(s.queue.Len())
}
