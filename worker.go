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
	queue             QueueInterface
	minWorkers        uint32 // minimal count of workers
	maxWorkers        uint32 // maximal count of workers
	minWorkingPercent uint32
	maxWorkingPercent uint32
	pollTaskInterval  time.Duration // period for check task in queue
	scheduleInterval  time.Duration
	tasks             chan TaskInterface
	workers           map[*worker]struct{}
	quit              chan struct{}
	returnErr         bool
	Errors            chan error
	sync.Mutex
}

type Option func(*WorkerPool)

// WithPollTaskInterval set duration for polling tasks from queue
func WithPollTaskInterval(duration time.Duration) Option {
	return func(w *WorkerPool) {
		w.pollTaskInterval = duration
	}
}

// WithScheduleInterval set duration for check workers and schedule
func WithScheduleInterval(duration time.Duration) Option {
	return func(w *WorkerPool) {
		w.scheduleInterval = duration
	}
}

func WithErrors() Option {
	return func(w *WorkerPool) {
		w.returnErr = true
	}
}

func (wp *WorkerPool) SetMinWorkers(count uint32) error {
	return wp.setMinWorkers(count)
}

func (wp *WorkerPool) GetMinWorkers() uint32 {
	return atomic.LoadUint32(&wp.minWorkers)
}

func (wp *WorkerPool) setMinWorkers(count uint32) error {
	if count < 2 {
		return fmt.Errorf("minWorkers can not be < 2")
	}
	if wp.maxWorkers < count {
		return fmt.Errorf("maxWorkers can not be < minWorkers")
	}
	atomic.StoreUint32(&wp.minWorkers, count)
	return nil
}

func (wp *WorkerPool) SetMaxWorkers(count uint32) error {
	return wp.setMaxWorkers(count)
}

func (wp *WorkerPool) GetMaxWorkers() uint32 {
	return atomic.LoadUint32(&wp.maxWorkers)
}

func (wp *WorkerPool) setMaxWorkers(count uint32) error {
	if count < wp.minWorkers {
		return fmt.Errorf("maxWorkers can not be < minWorkers")
	}
	atomic.StoreUint32(&wp.maxWorkers, count)
	return nil
}

func (wp *WorkerPool) SetMinWorkingPercent(per uint32) error {
	return wp.setMinWorkingPercent(per)
}

func (wp *WorkerPool) GetMinWorkingPercent() uint32 {
	return atomic.LoadUint32(&wp.minWorkingPercent)
}

func (wp *WorkerPool) setMinWorkingPercent(per uint32) error {
	if per > 99 {
		return fmt.Errorf("minWorkingPercent  can not be > 99")
	}
	if per > wp.maxWorkingPercent {
		return fmt.Errorf("minWorkingPercent  can not be > maxWorkingPercent")
	}
	atomic.StoreUint32(&wp.minWorkingPercent, per)
	return nil
}

func (wp *WorkerPool) SetMaxWorkingPercent(per uint32) error {
	return wp.setMaxWorkingPercent(per)
}

func (wp *WorkerPool) GetMaxWorkingPercent() uint32 {
	return atomic.LoadUint32(&wp.maxWorkingPercent)
}

func (wp *WorkerPool) setMaxWorkingPercent(per uint32) error {
	if per > 100 {
		return fmt.Errorf("maxWorkingPercent  can not be > 100")
	}
	if per < wp.minWorkingPercent {
		return fmt.Errorf("maxWorkingPercent  can not be < minWorkingPercent")
	}
	atomic.StoreUint32(&wp.maxWorkingPercent, per)
	return nil
}

// NewWorkerPool constructor for create WorkerPool
func NewWorkerPool(queue QueueInterface, opts ...Option) (*WorkerPool, error) {
	if queue == nil {
		return nil, fmt.Errorf("queue can not be nil")
	}

	wp := WorkerPool{
		queue:             queue,
		minWorkers:        2,
		maxWorkers:        25,
		minWorkingPercent: 50,
		maxWorkingPercent: 99,
		pollTaskInterval:  200 * time.Millisecond,
		quit:              make(chan struct{}),
		Errors:            make(chan error),
		tasks:             make(chan TaskInterface),
		workers:           make(map[*worker]struct{}),
		scheduleInterval:  time.Second * 1,
	}

	for _, opt := range opts {
		opt(&wp)
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
		if !w.inWork() {
			delete(wp.workers, w)
			break
		}
	}
}

func (wp *WorkerPool) removeAllWorkers() {
	wp.Lock()
	for len(wp.workers) != 0 {
		for w := range wp.workers {
			if !w.inWork() {
				delete(wp.workers, w)
			}
		}
	}
	wp.Unlock()
}

func (wp *WorkerPool) getStats() (current, inWork int) {
	wp.Lock()
	defer wp.Unlock()
	current = len(wp.workers)
	inWork = 0
	for w := range wp.workers {
		if w.inWork() {
			inWork++
		}
	}
	return
}

// run worker pool
func (wp *WorkerPool) Run() {
	pollTicker := time.NewTicker(wp.pollTaskInterval)
	defer pollTicker.Stop()

	scheduleTicker := time.NewTicker(wp.scheduleInterval)
	defer scheduleTicker.Stop()

	for i := 0; i < int(wp.GetMinWorkers()); i++ {
		wp.addWorker()
	}
	go func() {
		for {
			select {
			case <-scheduleTicker.C:
				current, inWork := wp.getStats()
				if current == 0 {
					return
				}
				workingPercent := uint32(inWork * 100 / current)
				if workingPercent < wp.GetMinWorkingPercent() && uint32(current) > wp.minWorkers {
					wp.removeWorker()
				}
				if workingPercent > wp.GetMaxWorkingPercent() && uint32(current) < wp.maxWorkers {
					wp.addWorker()
				}
			case <-pollTicker.C:
				task := wp.queue.GetTask()
				if task == nil {
					continue
				}

				wp.tasks <- task
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

func (w *worker) inWork() bool {
	return atomic.LoadInt32(&w.state) == StateInWork
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
