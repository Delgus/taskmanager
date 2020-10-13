package taskmanager

import "sync"

// Queue implement queue with priority
type Queue struct {
	store [5]*storage
}

func NewQueue() *Queue {
	return &Queue{
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
func (q *Queue) AddTask(task TaskInterface) {
	q.store[task.Priority()].push(task)
}

// GetTask get task
func (q *Queue) GetTask() (task TaskInterface) {
	for _, t := range q.store {
		task := t.pop()
		if task != nil {
			return task
		}
	}
	return nil
}

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
