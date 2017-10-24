package workerpool

import (
	"fmt"
	"github.com/splitio/go-client/splitio/util/logging"
	"sync"
	"time"
)

// WorkerAdmin struct handles multiple worker execution, popping jobs from a single queue
type WorkerAdmin struct {
	enabled      map[string]bool
	queue        chan interface{}
	logger       logging.LoggerInterface
	enabledMutex sync.RWMutex
}

// Worker interface should be implemented by concrete workers that will perform the actual job
type Worker interface {
	// Name should return a unique identifier for a particular worker
	Name() string
	// DoWork should receive a message, and perform the actual work, only an error should be returned
	DoWork(message interface{}) error
	// OnError will be called if DoWork returns an error != nil
	OnError(e error)
	// Cleanup will be called when the worker is shutting down
	Cleanup() error
	// FailureTime should return the amount of time the worker should wait after resuming work if an error occurs
	FailureTime() int64
}

func (a *WorkerAdmin) workerWrapper(w Worker) {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Error(fmt.Sprintf(
				"Worker %s is panicking with the following error \"%s\" and will be shutted down.",
				w.Name(),
				r,
			))
			a.enabledMutex.Lock()
			a.enabled[w.Name()] = false
			a.enabledMutex.Unlock()
		}
	}()
	defer w.Cleanup()
	for a.shouldBeWorking(w.Name()) {
		select {
		case msg := <-a.queue:
			if !a.shouldBeWorking(w.Name()) {
				// If by the time the worker wakes up it's execution has been cancelled,
				// Put the message back in the queue and return
				a.queue <- msg
				return
			}
			if err := w.DoWork(msg); err != nil {
				w.OnError(err)
				time.Sleep(time.Duration(w.FailureTime()) * time.Millisecond)
			}
		case <-time.After(time.Millisecond * 500):
			if !a.shouldBeWorking(w.Name()) {
				return
			}
		}
	}
}

// AddWorker registers a new worker in the admin
func (a *WorkerAdmin) AddWorker(w Worker) {
	if w == nil {
		a.logger.Error("AddWorker called with nil")
		return
	}
	go a.workerWrapper(w)
	a.enabledMutex.Lock()
	a.enabled[w.Name()] = true
	a.enabledMutex.Unlock()
}

// QueueMessage adds a new message that will be popped by a worker and processed
func (a *WorkerAdmin) QueueMessage(m interface{}) bool {
	if m == nil {
		a.logger.Warning("Nil message not added to queue")
		return false
	}
	select {
	case a.queue <- m:
		return true
	default:
		return false
	}
}

func (a *WorkerAdmin) shouldBeWorking(name string) bool {
	a.enabledMutex.RLock()
	status, found := a.enabled[name]
	a.enabledMutex.RUnlock()
	return found && status
}

// StopWorker ends the worker's event loop, preventing it from picking further jobs
func (a *WorkerAdmin) StopWorker(name string) {
	a.enabledMutex.Lock()
	a.enabled[name] = false
	a.enabledMutex.Unlock()
}

// StopAll ends all worker's event loops
func (a *WorkerAdmin) StopAll() {
	a.enabledMutex.Lock()
	for workerName := range a.enabled {
		a.enabled[workerName] = false
	}
	a.enabledMutex.Unlock()
}

// QueueSize returns the current queue size
func (a *WorkerAdmin) QueueSize() int {
	return len(a.queue)
}

// NewWorkerAdmin instantiates a new WorkerAdmin and returns a pointer to it.
func NewWorkerAdmin(queueSize int, logger logging.LoggerInterface) *WorkerAdmin {
	return &WorkerAdmin{
		enabled: make(map[string]bool, 0),
		logger:  logger,
		queue:   make(chan interface{}, queueSize),
	}
}
