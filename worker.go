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
	queue              QueueInterface
	minWorkers         uint32 // minimal count of workers
	maxWorkers         uint32 // maximal count of workers
	minWorkingPercent  uint32
	maxWorkingPercent  uint32
	pollTaskTicker     *time.Ticker
	scheduleTaskTicker *time.Ticker
	tasks              chan TaskInterface
	workers            map[*worker]struct{}
	quit               chan struct{}
	returnErr          bool
	Errors             chan error
	sync.Mutex
}

// NewWorkerPool constructor for create WorkerPool
func NewWorkerPool(queue QueueInterface, returnErr bool) *WorkerPool {
	return &WorkerPool{
		queue:              queue,
		minWorkers:         2,
		maxWorkers:         25,
		minWorkingPercent:  50,
		maxWorkingPercent:  99,
		pollTaskTicker:     time.NewTicker(200 * time.Millisecond),
		quit:               make(chan struct{}),
		Errors:             make(chan error),
		tasks:              make(chan TaskInterface),
		workers:            make(map[*worker]struct{}),
		scheduleTaskTicker: time.NewTicker(1 * time.Second),
		returnErr:          returnErr,
	}
}

func (wp *WorkerPool) SetPollTaskTicker(ticker *time.Ticker) {
	wp.Lock()
	wp.pollTaskTicker.Stop()
	wp.pollTaskTicker = ticker
	wp.Unlock()
}

func (wp *WorkerPool) SetScheduleTicker(ticker *time.Ticker) {
	wp.Lock()
	wp.scheduleTaskTicker.Stop()
	wp.scheduleTaskTicker = ticker
	wp.Unlock()
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

func (wp *WorkerPool) addWorker() {
	w := newWorker()
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
			w.close()
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
				w.close()
				delete(wp.workers, w)
			}
		}
	}
	wp.Unlock()
}

func (wp *WorkerPool) getStats() (current, inWork int) {
	wp.Lock()
	defer wp.Unlock()
	return wp.getCountWorkers(), wp.getCountWorkingWorkers()
}

func (wp *WorkerPool) getCountWorkers() int {
	return len(wp.workers)
}

func (wp *WorkerPool) getCountWorkingWorkers() (inWork int) {
	for w := range wp.workers {
		if w.inWork() {
			inWork++
		}
	}
	return
}

func (wp *WorkerPool) GetCountWorkers() int {
	wp.Lock()
	defer wp.Unlock()
	return len(wp.workers)
}

func (wp *WorkerPool) GetCountWorkingWorkers() int {
	wp.Lock()
	defer wp.Unlock()
	return wp.getCountWorkingWorkers()
}

// run worker pool
func (wp *WorkerPool) Run() {
	defer func() {
		wp.pollTaskTicker.Stop()
	}()

	defer func() {
		wp.scheduleTaskTicker.Stop()
	}()

	for i := 0; i < int(wp.GetMinWorkers()); i++ {
		wp.addWorker()
	}
	go func() {
		for {
			select {
			case <-wp.scheduleTaskTicker.C:
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
			case <-wp.pollTaskTicker.C:
				task := wp.queue.GetTask()
				if task != nil {
					wp.tasks <- task
				}
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
	StateListen int32 = 0
	StateInWork int32 = 1
)

func (w *worker) setState(state int32) {
	atomic.StoreInt32(&w.state, state)
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
