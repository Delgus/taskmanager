package taskmanager

// Priority - type for priority level
type Priority uint8

const (
	HighestPriority Priority = iota
	HighPriority
	MiddlePriority
	LowPriority
	LowestPriority
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
