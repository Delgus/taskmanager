package taskmanager

// Priority - type for priority level
type Priority uint8

const (
	HighestPriority Priority = 0
	HighPriority    Priority = 1
	MiddlePriority  Priority = 2
	LowPriority     Priority = 3
	LowestPriority  Priority = 4
)

type executor interface {
	Exec() error
}

type priorityHaving interface {
	Priority() Priority
}

type attempts interface {
	Attempts() uint32
}

// Event - type for event name
type Event string

// EventHandler - handle for events
type EventHandler func()

const (
	BeforeExecEvent Event = "before_exec"
	AfterExecEvent  Event = "after_exec"
	FailedEvent     Event = "failed"
)

type eventEmitter interface {
	OnEvent(event Event, handler EventHandler)
	EmitEvent(event Event)
}

// Task interface
type TaskInterface interface {
	executor
	priorityHaving
	eventEmitter
	attempts
}

// Queue interface
type QueueInterface interface {
	AddTask(task TaskInterface)
	GetTask() TaskInterface
}
