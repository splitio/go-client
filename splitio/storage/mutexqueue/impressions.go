package mutexqueue

import (
	"container/list"
	"sync"

	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-toolkit/logging"
)

// NewMQImpressionsStorage returns an instance of MQEventsStorage
func NewMQImpressionsStorage(queueSize int, isFull chan<- string, logger logging.LoggerInterface) *MQImpressionsStorage {
	return &MQImpressionsStorage{
		queue:      list.New(),
		size:       queueSize,
		mutexQueue: &sync.Mutex{},
		fullChan:   isFull,
		logger:     logger,
	}
}

// MQImpressionsStorage in memory events storage
type MQImpressionsStorage struct {
	queue      *list.List
	size       int
	mutexQueue *sync.Mutex
	fullChan   chan<- string //only write channel
	logger     logging.LoggerInterface
}

func (s *MQImpressionsStorage) sendSignalIsFull() {
	// Nom blocking select
	select {
	case s.fullChan <- "IMPRESSIONS_FULL":
		// Send "queue is full" signal
		break
	default:
		s.logger.Debug("Some error occurred on sending signal for impressions")
		break
	}
}

// LogImpressions inserts impressions into the queue
func (s *MQImpressionsStorage) LogImpressions(impressions []storage.Impression) error {
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	for _, impression := range impressions {
		if s.queue.Len()+1 > s.size {
			s.sendSignalIsFull()
			return ErrorMaxSizeReached
		}
		// Add element
		s.queue.PushBack(impression)

		if s.queue.Len() == s.size {
			s.sendSignalIsFull()
		}
	}
	return nil
}

// PopN pop N elements from queue
func (s *MQImpressionsStorage) PopN(n int64) ([]storage.Impression, error) {
	var toReturn []storage.Impression
	var totalItems int

	// Mutexing queue
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	if int64(s.queue.Len()) >= n {
		totalItems = int(n)
	} else {
		totalItems = s.queue.Len()
	}

	toReturn = make([]storage.Impression, totalItems)
	for i := 0; i < totalItems; i++ {
		toReturn[i] = s.queue.Remove(s.queue.Front()).(storage.Impression)
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
