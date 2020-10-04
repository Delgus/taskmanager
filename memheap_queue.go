package taskmanager

import (
	"sync"
)

// HeapQueue implement queue with priority
type HeapQueue struct {
	queue queue
	mu    sync.Mutex
}

// AddTask add task
func (q *HeapQueue) AddTask(task TaskInterface) error {
	q.mu.Lock()
	q.queue.push(task)
	q.mu.Unlock()
	return nil
}

// GetTask get task
func (q *HeapQueue) GetTask() (task TaskInterface, err error) {
	q.mu.Lock()
	if len(q.queue) > 0 {
		task = q.queue.pop()
	}
	q.mu.Unlock()
	return
}

type queue []TaskInterface

func (q *queue) push(t TaskInterface) {
	*q = append(*q, t)
	q.up()
}

func (q *queue) pop() TaskInterface {
	q.swap(0, len(*q)-1)
	q.down()

	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid leak memheap
	*q = old[0 : n-1]
	return item
}

func (q queue) less(i, j int) bool {
	return q[i].Priority() < q[j].Priority()
}

func (q queue) swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

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