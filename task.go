package taskmanager

import (
	"fmt"
	"sync/atomic"
)

// Task implement TaskInterface
type Task struct {
	priority Priority
	handler  TaskHandler
	attempts uint32
}

// TaskHandler type handle for task or job
type TaskHandler func() error

// NewTask created new task
func NewTask(priority Priority, handler TaskHandler) *Task {
	return &Task{
		priority: priority,
		handler:  handler,
		attempts: 1,
	}
}

// Priority return priority of task
func (t *Task) Priority() Priority {
	return t.priority
}

// SetAttempts set attempts
func (t *Task) SetAttempts(attempts uint32) {
	atomic.StoreUint32(&t.attempts, attempts)
}

// Attempts return left attempts
func (t *Task) Attempts() uint32 {
	return atomic.LoadUint32(&t.attempts)
}

func (t *Task) Exec() (err error) {
	atomic.StoreUint32(&t.attempts, atomic.LoadUint32(&t.attempts)-1)
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %s", e)
		}
	}()
	return t.handler()
}
