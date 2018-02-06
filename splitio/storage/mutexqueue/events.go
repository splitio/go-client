package mutexqueue

import (
	"container/list"
	"errors"
	"sync"

	"github.com/splitio/go-client/splitio/service/dtos"
)

// ErrorMaxSizeReached queue max size error
var ErrorMaxSizeReached = errors.New("Queue max size has been reached")

// NewMQEventsStorage returns an instance of MQEventsStorage
func NewMQEventsStorage(queueSize int, isFull chan<- bool) *MQEventsStorage {
	return &MQEventsStorage{
		queue:      list.New(),
		size:       queueSize,
		mutexQueue: &sync.Mutex{},
		fullChan:   isFull,
	}
}

// MQEventsStorage in memory events storage
type MQEventsStorage struct {
	queue      *list.List
	size       int
	mutexQueue *sync.Mutex
	fullChan   chan<- bool //only write channel
}

func (s *MQEventsStorage) sendSignalIsFull() {
	// Nom blocking select
	select {
	case s.fullChan <- true:
		//Send "queue is full" signal
		break
	default:
		break
	}
}

// Push an event into slice
func (s *MQEventsStorage) Push(event dtos.EventDTO) error {
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	if s.queue.Len()+1 > s.size {
		s.sendSignalIsFull()
		return ErrorMaxSizeReached
	}

	// Add element
	s.queue.PushBack(event)

	if s.queue.Len() == s.size {
		s.sendSignalIsFull()
	}

	return nil
}

// PopN pop N elements from queue
func (s *MQEventsStorage) PopN(n int64) ([]dtos.EventDTO, error) {
	var toReturn []dtos.EventDTO
	var totalItems int

	// Mutexing queue
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	if int64(s.queue.Len()) >= n {
		totalItems = int(n)
	} else {
		totalItems = s.queue.Len()
	}

	toReturn = make([]dtos.EventDTO, 0)
	for i := 0; i < totalItems; i++ {
		toReturn = append(toReturn, s.queue.Remove(s.queue.Front()).(dtos.EventDTO))
	}

	return toReturn, nil
}

// Empty returns if slice len if zero
func (s *MQEventsStorage) Empty() bool {
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	return s.queue.Len() == 0
}

// Count returns the number of events into slice
func (s *MQEventsStorage) Count() int64 {
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	return int64(s.queue.Len())
}
