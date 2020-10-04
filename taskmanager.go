package taskmanager

// Priority - type for priority level
type Priority int

const (
	// HighestPriority - highest priority level
	HighestPriority Priority = 0
	// HighPriority - high priority level
	HighPriority Priority = 1
	// MiddlePriority - middle priority level
	MiddlePriority Priority = 2
	// LowPriority - low priority level
	LowPriority Priority = 3
	// LowestPriority - lowest priority level
	LowestPriority Priority = 4
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
type TaskInterface interface {
	executor
	prioritier
	eventer
}

// Queue interface
type QueueInterface interface {
	// AddTask ...
	AddTask(task TaskInterface) error
	// GetTask ...
	GetTask() (TaskInterface, error)
}
