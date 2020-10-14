package taskmanager

import (
	"sync/atomic"
)

// Task implement TaskInterface
type Task struct {
	events   map[Event][]EventHandler
	priority Priority
	handler  TaskHandler
	attempts uint32
}

// TaskHandler type handle for task or job
type TaskHandler func() error

// NewTask created new task
func NewTask(priority Priority, handler TaskHandler) *Task {
	task := &Task{
		events:   make(map[Event][]EventHandler),
		priority: priority,
		handler:  handler,
		attempts: 1,
	}
	return task
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

// Exec - Perform the task
// When the task starts emitting BeforeExecEvent
// If the task is unsuccessful calling the FailedEvent
// If you need more flexible error handling - implement your Task
// At the end of the task, if it has passed successfully, call the AfterExecEvent event
func (t *Task) Exec() error {
	t.EmitEvent(BeforeExecEvent)
	atomic.StoreUint32(&t.attempts, atomic.LoadUint32(&t.attempts)-1)
	if err := t.handler(); err != nil {
		t.EmitEvent(FailedEvent)
		return err
	}
	t.EmitEvent(AfterExecEvent)
	return nil
}

// OnEvent - added handler on event
func (t *Task) OnEvent(event Event, handler EventHandler) {
	t.events[event] = append(t.events[event], handler)
}

// EmitEvent emit event
func (t *Task) EmitEvent(event Event) {
	for _, h := range t.events[event] {
		h()
	}
}
