package taskmanager

import (
	"sync"
)

type storage struct {
	sync.RWMutex
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

func (s *storage) last() TaskInterface {
	s.Lock()
	defer s.Unlock()
	item := s.items[len(s.items)-1]
	s.items[len(s.items)-1] = nil
	s.items = s.items[:len(s.items)-1]
	return item
}

// HeapQueue implement queue with priority
type SliceQueue struct {
	store [5]*storage
}

func NewSliceQueue() *SliceQueue {
	return &SliceQueue{
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
func (q *SliceQueue) AddTask(task TaskInterface) error {
	q.store[task.Priority()].append(task)
	return nil
}

// GetTask get task
func (q *SliceQueue) GetTask() (task TaskInterface, err error) {
	for _, t := range q.store {
		if t.len() > 0 {
			return t.last(), nil
		}
	}
	return nil, nil
}
