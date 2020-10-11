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

func (s *storage) len() int {
	s.Lock()
	defer s.Unlock()
	return len(s.items)
}

func (s *storage) append(item TaskInterface) {
	s.Lock()
	defer s.Unlock()

	s.items = append(s.items, item)
}

func (s *storage) first() TaskInterface {
	s.Lock()
	defer s.Unlock()
	item := s.items[0]
	s.items = s.items[1:]
	return item
}

// HeapQueue implement queue with priority
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
func (q *MemoryQueue) AddTask(task TaskInterface) error {
	q.store[task.Priority()].append(task)
	return nil
}

// GetTask get task
func (q *MemoryQueue) GetTask() (task TaskInterface, err error) {
	for _, t := range q.store {
		if t.len() > 0 {
			return t.first(), nil
		}
	}
	return nil, nil
}
