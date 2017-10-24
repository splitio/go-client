package workerpool

import (
	"fmt"
	"github.com/splitio/go-client/splitio/util/logging"
	"sync"
	"testing"
	"time"
)

var resMutex sync.RWMutex

type okWorker struct {
	id      int
	results map[string]int
}

func (w *okWorker) Name() string {
	return fmt.Sprintf("worker_%d", w.id)
}

func (w *okWorker) DoWork(msg interface{}) error {
	asInt := msg.(int)
	resMutex.Lock()
	w.results[w.Name()] = asInt
	resMutex.Unlock()
	return nil
}

// Dummy implementations just to comply with Worker interface
func (w *okWorker) Cleanup() error     { return nil }
func (w *okWorker) OnError(err error)  {}
func (w *okWorker) FailureTime() int64 { return 0 }

func TestWorkerAdminConstructionAndNormalOperation(t *testing.T) {
	// To test this, we'll use a map of strings to int.
	// Each worker will store the number popped from the queue in the map with the key being the worker name.
	// This will allow us to track which worker has processed each message and determine if a stopped worker keeps processing messages.
	// TODO
	logger := logging.NewLogger(&logging.LoggerOptions{})
	wa := NewWorkerAdmin(100, logger)
	results := make(map[string]int)
	wa.AddWorker(&okWorker{id: 1, results: results})
	wa.AddWorker(&okWorker{id: 2, results: results})
	wa.AddWorker(&okWorker{id: 3, results: results})

	for i := 0; i < 10; i++ {
		wa.QueueMessage(i)
	}
	wa.StopWorker("worker_2")
	for i := 10; i < 20; i++ {
		wa.QueueMessage(i)
	}

	resMutex.RLock()
	if results["worker_2"] > 10 {
		t.Error("Worker should have stopped working!")
	}
	resMutex.RUnlock()
	time.Sleep(time.Second * 1)
	wa.StopAll()
}

type failingWorker struct {
	id int
}

func (w *failingWorker) Name() string {
	return fmt.Sprintf("worker_%d", w.id)
}

func (w *failingWorker) DoWork(msg interface{}) error {
	panic("explota todooo")
}

// Dummy implementations just to comply with Worker interface
func (w *failingWorker) Cleanup() error     { return nil }
func (w *failingWorker) OnError(err error)  {}
func (w *failingWorker) FailureTime() int64 { return 0 }

func TestWorkerAdminWithFailingWorkers(t *testing.T) {
	// This test asserts that if a worker panics when performing work,
	// the panic is caught and the goroutine ends gracefully without the panic
	// being propagated.
	logger := logging.NewLogger(&logging.LoggerOptions{})
	wa := NewWorkerAdmin(100, logger)
	wa.AddWorker(&failingWorker{id: 1})
	wa.AddWorker(&failingWorker{id: 2})
	wa.AddWorker(&failingWorker{id: 3})

	for i := 0; i < 10; i++ {
		wa.QueueMessage(i)
	}
}
