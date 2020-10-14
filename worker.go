package taskmanager

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// WorkerPool of workers
type WorkerPool struct {
	queue            QueueInterface
	minWorkers       uint32        // minimal count of workers
	maxWorkers       uint32        // maximal count of workers
	pollTaskInterval time.Duration // period for check task in queue
	tasks            chan TaskInterface
	workers          map[*worker]struct{}
	quit             chan struct{}
	returnErr        bool
	Errors           chan error
	sync.Mutex
}

type Option func(*WorkerPool)

func WithPollTaskInterval(duration time.Duration) Option {
	return func(w *WorkerPool) {
		w.pollTaskInterval = duration
	}
}

func WithMinWorkers(count uint32) Option {
	return func(w *WorkerPool) {
		w.minWorkers = count
	}
}

func WithMaxWorkers(count uint32) Option {
	return func(w *WorkerPool) {
		w.maxWorkers = count
	}
}

func WithErrors() Option {
	return func(w *WorkerPool) {
		w.returnErr = true
	}
}

// NewWorkerPool constructor for create WorkerPool
func NewWorkerPool(queue QueueInterface, opts ...Option) (*WorkerPool, error) {
	if queue == nil {
		return nil, fmt.Errorf("queue can not be nil")
	}

	wp := WorkerPool{
		queue:            queue,
		minWorkers:       10,
		maxWorkers:       25,
		pollTaskInterval: 200 * time.Millisecond,
		quit:             make(chan struct{}),
		Errors:           make(chan error),
		tasks:            make(chan TaskInterface),
		workers:          make(map[*worker]struct{}),
	}

	for _, opt := range opts {
		opt(&wp)
	}

	if wp.minWorkers < 2 {
		return nil, fmt.Errorf("minWorkers can not be < 2")
	}
	if wp.maxWorkers < wp.minWorkers {
		return nil, fmt.Errorf("maxWorkers can not be < minWorkers")
	}

	return &wp, nil
}

func (wp *WorkerPool) addWorker() {
	w := new(worker)
	w.pool = wp
	wp.Lock()
	wp.workers[w] = struct{}{}
	wp.Unlock()
	go w.run()
}

func (wp *WorkerPool) removeWorker() {
	wp.Lock()
	defer wp.Unlock()
	for w := range wp.workers {
		if w.getState() != StateInWork {
			delete(wp.workers, w)
			break
		}
	}
}

func (wp *WorkerPool) removeAllWorkers() {
	wp.Lock()
	for len(wp.workers) != 0 {
		for w := range wp.workers {
			if w.getState() != StateInWork {
				delete(wp.workers, w)
			}
		}
	}
	wp.Unlock()
}

func (wp *WorkerPool) getStatistic() (current, inWork int) {
	wp.Lock()
	defer wp.Unlock()
	current = len(wp.workers)
	inWork = 0
	for w := range wp.workers {
		if w.getState() == StateInWork {
			inWork++
		}
	}
	return
}

// run worker pool
func (wp *WorkerPool) Run() {
	ticker := time.NewTicker(wp.pollTaskInterval)
	defer ticker.Stop()
	for i := 0; i < int(wp.minWorkers); i++ {
		wp.addWorker()
	}
	go func() {
		for range ticker.C {
			current, inWork := wp.getStatistic()
			if current < 1 {
				continue
			}
			task := wp.queue.GetTask()
			if task == nil {
				// if < 50% do nothing - remove worker
				if inWork/current*100 < 50 && uint32(current) > wp.minWorkers {
					wp.removeWorker()
				}
				continue
			}
			// if all workers in work - add worker
			if inWork == current && uint32(current) < wp.maxWorkers {
				wp.addWorker()
			}
			wp.tasks <- task
		}
	}()
	<-wp.quit
}

// Shutdown - the worker will not stop until he has completed all the unfinished tasks
// or the timeout does not expire
func (wp *WorkerPool) Shutdown(ctx context.Context) error {
	defer close(wp.quit)

	ok := make(chan struct{})
	go func() {
		wp.removeAllWorkers()
		ok <- struct{}{}
	}()

	select {
	case <-ok:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type worker struct {
	state int32
	pool  *WorkerPool
}

const (
	StateListen int32 = 0
	StateInWork int32 = 1
)

func (w *worker) setState(state int32) {
	atomic.StoreInt32(&w.state, state)
}

func (w *worker) getState() int32 {
	return atomic.LoadInt32(&w.state)
}

func (w *worker) run() {
	for {
		task := <-w.pool.tasks
		w.setState(StateInWork)
		if err := task.Exec(); err != nil {
			if w.pool.returnErr {
				w.pool.Errors <- err
			}
			if task.Attempts() != 0 {
				w.pool.tasks <- task
			}
		}
		w.setState(StateListen)
	}
}
