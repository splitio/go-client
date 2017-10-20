package workerpool

import (
	"fmt"
	"github.com/splitio/go-client/splitio/util/logging"
	"time"
)

// WorkerAdmin struct handles multiple worker execution, popping jobs from a single queue
type WorkerAdmin struct {
	enabled map[string]bool
	queue   chan interface{}
	logger  logging.LoggerInterface
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
			a.enabled[w.Name()] = false
		}
	}()
	defer w.Cleanup()
	for a.shouldBeWorking(w.Name()) {
		msg := <-a.queue
		if err := w.DoWork(msg); err != nil {
			w.OnError(err)
			time.Sleep(time.Duration(w.FailureTime()) * time.Millisecond)
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
	a.enabled[w.Name()] = true
}

// QueueMessage adds a new message that will be popped by a worker and processed
func (a *WorkerAdmin) QueueMessage(m interface{}) {
	if m == nil {
		a.logger.Warning("Nil message not added to queue")
		return
	}
	a.queue <- m
}

func (a *WorkerAdmin) shouldBeWorking(name string) bool {
	status, found := a.enabled[name]
	return found && status
}

// RemoveWorker ends the worker's event loop, preventing it from picking further jobs
func (a *WorkerAdmin) RemoveWorker(name string) {
	a.enabled[name] = false
}

// StopAll ends all worker's event loops
func (a *WorkerAdmin) StopAll() {
	for workerName := range a.enabled {
		a.enabled[workerName] = false
	}
}

// NewWorkerAdmin instantiates a new WorkerAdmin and returns a pointer to it.
func NewWorkerAdmin(queueSize int, logger logging.LoggerInterface) *WorkerAdmin {
	return &WorkerAdmin{
		enabled: make(map[string]bool, 0),
		logger:  logger,
		queue:   make(chan interface{}, queueSize),
	}
}
