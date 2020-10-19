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

type havingPriority interface {
	Priority() Priority
}

type attempting interface {
	Attempts() uint32
}

// Task interface
type TaskInterface interface {
	executor
	havingPriority
	attempting
}

// Queue interface
type QueueInterface interface {
	AddTask(task TaskInterface)
	GetTask() TaskInterface
}
