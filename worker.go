package taskmanager

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	validMinWorkers = 2
)

// WorkerPool of workers
type WorkerPool struct {
	queue               QueueInterface
	minWorkers          uint32 // minimal count of workers
	maxWorkers          uint32 // maximal count of workers
	countWorkers        int64
	countWorkingWorkers int64
	pollTaskTicker      *time.Ticker
	gcTicker            *time.Ticker
	tasks               chan TaskInterface
	workers             map[*worker]struct{}
	quit                chan struct{}
	returnErr           bool
	Errors              chan error
	sync.Mutex
}

// NewWorkerPool constructor for create WorkerPool
func NewWorkerPool(queue QueueInterface, returnErr bool) *WorkerPool {
	// default values
	var minWorkers, maxWorkers uint32 = 2, 25
	var pollDuration, scheduleDuration = 200 * time.Millisecond, 1 * time.Second
	return &WorkerPool{
		queue:          queue,
		minWorkers:     minWorkers,
		maxWorkers:     maxWorkers,
		pollTaskTicker: time.NewTicker(pollDuration),
		quit:           make(chan struct{}),
		Errors:         make(chan error),
		tasks:          make(chan TaskInterface),
		workers:        make(map[*worker]struct{}),
		gcTicker:       time.NewTicker(scheduleDuration),
		returnErr:      returnErr,
	}
}

// Set duration for polling tasks.
func (wp *WorkerPool) SetPollTaskDuration(duration time.Duration) {
	wp.Lock()
	wp.pollTaskTicker.Reset(duration)
	wp.Unlock()
}

// Set duration for check count of workers for adding new or removing
func (wp *WorkerPool) SetGCDuration(duration time.Duration) {
	wp.Lock()
	wp.gcTicker.Reset(duration)
	wp.Unlock()
}

// Set min count running workers in pool
func (wp *WorkerPool) SetMinWorkers(count uint32) error {
	if count < validMinWorkers {
		return fmt.Errorf("minWorkers can not be < 2")
	}
	if wp.maxWorkers < count {
		return fmt.Errorf("maxWorkers can not be < minWorkers")
	}
	atomic.StoreUint32(&wp.minWorkers, count)
	return nil
}

func (wp *WorkerPool) GetMinWorkers() uint32 {
	return atomic.LoadUint32(&wp.minWorkers)
}

// Set max count running workers in pool
func (wp *WorkerPool) SetMaxWorkers(count uint32) error {
	if count < wp.minWorkers {
		return fmt.Errorf("maxWorkers can not be < minWorkers")
	}
	atomic.StoreUint32(&wp.maxWorkers, count)
	return nil
}

func (wp *WorkerPool) GetMaxWorkers() uint32 {
	return atomic.LoadUint32(&wp.maxWorkers)
}

func (wp *WorkerPool) addWorker() {
	w := newWorker()
	w.pool = wp
	wp.Lock()
	wp.workers[w] = struct{}{}
	atomic.AddInt64(&wp.countWorkers, 1)
	wp.Unlock()
	go w.run()
}

func (wp *WorkerPool) removeWorker() {
	wp.Lock()
	defer wp.Unlock()
	for w := range wp.workers {
		if !w.inWork() {
			w.close()
			delete(wp.workers, w)
			atomic.AddInt64(&wp.countWorkers, -1)
			break
		}
	}
}

func (wp *WorkerPool) removeAllWorkers() {
	wp.Lock()
	for len(wp.workers) != 0 {
		for w := range wp.workers {
			if !w.inWork() {
				w.close()
				delete(wp.workers, w)
				atomic.AddInt64(&wp.countWorkers, -1)
			}
		}
	}
	wp.Unlock()
}

func (wp *WorkerPool) GetCountWorkers() int64 {
	return atomic.LoadInt64(&wp.countWorkers)
}

func (wp *WorkerPool) GetCountWorkingWorkers() int64 {
	return atomic.LoadInt64(&wp.countWorkingWorkers)
}

// run worker pool
func (wp *WorkerPool) Run() {
	defer func() {
		wp.pollTaskTicker.Stop()
	}()

	defer func() {
		wp.gcTicker.Stop()
	}()

	for i := 0; i < int(wp.GetMinWorkers()); i++ {
		wp.addWorker()
	}
	go func() {
		for range wp.pollTaskTicker.C {
			// check ready workers
			all := wp.GetCountWorkers()
			isMax := all == int64(wp.maxWorkers)
			isAllInWork := wp.GetCountWorkingWorkers() == all
			if isMax && isAllInWork {
				continue
			}

			// check tasks
			task := wp.queue.GetTask()
			if task == nil {
				continue
			}

			// add worker if needed
			if isAllInWork {
				wp.addWorker()
			}

			wp.tasks <- task
		}
	}()

	// garbage collector
	go func() {
		for range wp.gcTicker.C {
			inWork := wp.GetCountWorkingWorkers()
			all := wp.GetCountWorkers()
			if inWork < all && uint32(all) > wp.minWorkers {
				wp.removeWorker()
			}
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
	state   int32
	pool    *WorkerPool
	closeCh chan struct{}
}

func newWorker() *worker {
	return &worker{
		closeCh: make(chan struct{}),
	}
}

const (
	StateListen int32 = iota
	StateInWork
)

func (w *worker) setState(state int32) {
	atomic.StoreInt32(&w.state, state)
	if state == StateInWork {
		atomic.AddInt64(&w.pool.countWorkingWorkers, 1)
	} else {
		atomic.AddInt64(&w.pool.countWorkingWorkers, -1)
	}
}

func (w *worker) inWork() bool {
	return atomic.LoadInt32(&w.state) == StateInWork
}

func (w *worker) close() {
	w.closeCh <- struct{}{}
}

func (w *worker) run() {
	for {
		select {
		case task := <-w.pool.tasks:
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
		case <-w.closeCh:
			w.pool = nil
			close(w.closeCh)
			return
		}
	}
}
