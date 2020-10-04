package taskmanager

import "fmt"

// HeapQueue implement queue with priority
type SliceQueue struct {
	store [5]chan TaskInterface
	size  int
}

func NewSliceQueue(buffer int) *SliceQueue {
	return &SliceQueue{
		store: [5]chan TaskInterface{
			HighestPriority: make(chan TaskInterface, buffer),
			HighPriority:    make(chan TaskInterface, buffer),
			MiddlePriority:  make(chan TaskInterface, buffer),
			LowPriority:     make(chan TaskInterface, buffer),
			LowestPriority:  make(chan TaskInterface, buffer),
		},
		size: buffer,
	}
}

// AddTask add task
func (q *SliceQueue) AddTask(task TaskInterface) error {
	if len(q.store[task.Priority()]) == q.size {
		return fmt.Errorf("can not add task: no space left in buffer")
	}
	q.store[task.Priority()] <- task
	return nil
}

// GetTask get task
func (q *SliceQueue) GetTask() (task TaskInterface, err error) {
	for _, ch := range q.store {
		if len(ch) > 0 {
			task = <-ch
			return
		}
	}
	return
}
