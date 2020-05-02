package memory

import (
	"sync"

	"github.com/delgus/taskmanager"
)

// Queue implement queue with priority
type Queue struct {
	queue queue
	mu    sync.Mutex
}

// AddTask add task
func (q *Queue) AddTask(task taskmanager.Task) {
	q.mu.Lock()
	q.queue.push(task)
	q.mu.Unlock()
}

// GetTask get task
func (q *Queue) GetTask() taskmanager.Task {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.queue) > 0 {
		return q.queue.pop()
	}
	return nil
}

type queue []taskmanager.Task

func (q *queue) push(t taskmanager.Task) {
	*q = append(*q, t)
	q.up()
}

func (q *queue) pop() taskmanager.Task {
	q.swap(0, len(*q)-1)
	q.down()

	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*q = old[0 : n-1]
	return item
}

func (q queue) less(i, j int) bool {
	return q[i].Priority() > q[j].Priority()
}

func (q queue) swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

// up (see container/heap)
func (q queue) up() {
	j := len(q) - 1
	for {
		i := (j - 1) / 2
		if i == j || !q.less(j, i) {
			break
		}
		q.swap(i, j)
		j = i
	}
}

// down (see container/heap)
func (q queue) down() {
	n := len(q) - 1
	var i int
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && q.less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		if !q.less(j, i) {
			break
		}
		q.swap(i, j)
		i = j
	}
}

// Task implement taskmanager.Task
type Task struct {
	events   map[taskmanager.Event][]taskmanager.EventHandler
	priority taskmanager.Priority
	handler  TaskHandler
}

// TaskHandler type handle for task or job
type TaskHandler func() error

// NewTask created new task
func NewTask(priority taskmanager.Priority, handler TaskHandler) *Task {
	task := &Task{
		events:   make(map[taskmanager.Event][]taskmanager.EventHandler),
		priority: priority,
		handler:  handler,
	}
	return task
}

// Priority return priority of task
func (t *Task) Priority() taskmanager.Priority {
	return t.priority
}

// Exec - Perform the task
// When the task starts emitting BeforeExecEvent
// If the task is unsuccessful calling the FailedEvent
// If you need more flexible error handling - implement your Task
// At the end of the task, if it has passed successfully, call the AfterExecEvent event
func (t *Task) Exec() error {
	t.EmitEvent(taskmanager.BeforeExecEvent)
	if err := t.handler(); err != nil {
		t.EmitEvent(taskmanager.FailedEvent)
		return err
	}
	t.EmitEvent(taskmanager.AfterExecEvent)
	return nil
}

// OnEvent - added handler on event
func (t *Task) OnEvent(event taskmanager.Event, handler taskmanager.EventHandler) {
	t.events[event] = append(t.events[event], handler)
}

// EmitEvent emit event
func (t *Task) EmitEvent(event taskmanager.Event) {
	for _, h := range t.events[event] {
		h()
	}
}
