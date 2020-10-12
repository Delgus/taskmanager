package taskmanager

import "sync"

type storage struct {
	sync.Mutex
	items []TaskInterface
}

func newStorage() *storage {
	return &storage{
		items: make([]TaskInterface, 0),
	}
}

func (s *storage) push(item TaskInterface) {
	s.Lock()
	s.items = append(s.items, item)
	s.Unlock()
}

func (s *storage) pop() TaskInterface {
	s.Lock()
	defer s.Unlock()
	if len(s.items) > 0 {
		item := s.items[0]
		s.items[0] = nil
		s.items = s.items[1:]
		return item
	}
	return nil
}

// MemoryQueue implement queue with priority
type MemoryQueue struct {
	store [5]*storage
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		store: [5]*storage{
			HighestPriority: newStorage(),
			HighPriority:    newStorage(),
			MiddlePriority:  newStorage(),
			LowPriority:     newStorage(),
			LowestPriority:  newStorage(),
		},
	}
}

// AddTask add task
func (mq *MemoryQueue) AddTask(task TaskInterface) {
	mq.store[task.Priority()].push(task)
}

// GetTask get task
func (mq *MemoryQueue) GetTask() (task TaskInterface) {
	for _, t := range mq.store {
		task := t.pop()
		if task != nil {
			return task
		}
	}
	return nil
}
