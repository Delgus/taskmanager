package taskmanager

// Priority - type for priority level
type Priority int

const (
	// LowestPriority - lowest priority level
	LowestPriority Priority = 1
	// LowPriority - low priority level
	LowPriority Priority = 2
	// MiddlePriority - middle priority level
	MiddlePriority Priority = 3
	// HighPriority - high priority level
	HighPriority Priority = 4
	// HighestPriority - highest priority level
	HighestPriority Priority = 5
)

type executor interface {
	Exec() error
}
type prioritier interface {
	Priority() Priority
}

// Event - type for event name
type Event string

// EventHandler - handle for events
type EventHandler func()

const (
	// BeforeExecEvent ...
	BeforeExecEvent Event = "before_exec"
	// AfterExecEvent ...
	AfterExecEvent Event = "after_exec"
	// FailedEvent ...
	FailedEvent Event = "failed"
)

type eventer interface {
	OnEvent(event Event, handler EventHandler)
	EmitEvent(event Event)
}

// Task interface
type Task interface {
	executor
	prioritier
	eventer
}

// Queue interface
type Queue interface {
	// AddTask ...
	AddTask(task Task)
	// GetTask ...
	GetTask() Task
}
