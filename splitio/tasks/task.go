package tasks

import (
	"fmt"
	"github.com/splitio/go-client/splitio/util/logging"
	"sync/atomic"
	"time"
)

// AsyncTask is a struct that wraps tasks that should run periodically and can be remotely stopped & started,
// as well as making it's status (running/stopped) available.
type AsyncTask struct {
	task       func(l logging.LoggerInterface) error
	name       string
	running    bool
	stopSignal atomic.Value
	period     int64
	onStop     func(l logging.LoggerInterface)
	logger     logging.LoggerInterface
}

// Start initiates the task. It wraps the execution in a closure guarded by a call to recover() in order
// to prevent the main application from crashin if something goes wrong while the sdk interacts with the backend.
func (t *AsyncTask) Start() {
	t.stopSignal.Store(false)

	if t.running {
		if t.logger != nil {
			t.logger.Warning("Task %s is already running. Aborting new execution.", t.name)
		}
		return
	}
	t.running = true

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if t.logger != nil {
					t.logger.Error(fmt.Sprintf(
						"AsyncTask %s is panicking! Delaying execution for %d seconds (1 period)",
						t.name,
						t.period,
					))
				}
				time.Sleep(time.Duration(t.period) * time.Second)
			}
		}()
		// Load and type assert the contents of the atomic variable `stopSignal`.
		// Keep the task running as long as the stopSignal is not true or until
		// something other than a boolean is stored in the atomic variable.
		stop, ok := t.stopSignal.Load().(bool)
		for ; ok && !stop; stop, ok = t.stopSignal.Load().(bool) {
			err := t.task(t.logger)
			if err != nil && t.logger != nil {
				t.logger.Error(err.Error())
			}
			time.Sleep(time.Duration(t.period) * time.Millisecond)
		}
		t.running = false
		if t.onStop != nil {
			t.onStop(t.logger)
		}
	}()
}

// Stop prevents future executions of the task
func (t *AsyncTask) Stop() {
	t.stopSignal.Store(true)
}

// IsRunning returns true if the task is currently running
func (t *AsyncTask) IsRunning() bool {
	return t.running
}

// NewAsyncTask creates a new task and returns a pointer to it
func NewAsyncTask(
	name string,
	task func(l logging.LoggerInterface) error,
	period int64,
	onStop func(l logging.LoggerInterface),
	logger logging.LoggerInterface,
) *AsyncTask {
	t := AsyncTask{
		name:    name,
		task:    task,
		running: false,
		period:  period,
		onStop:  onStop,
		logger:  logger,
	}
	t.stopSignal.Store(false)

	return &t
}
