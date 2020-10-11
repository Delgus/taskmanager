package taskmanager

import (
	"fmt"
	"sync"
	"time"
)

// Logger interface
type Logger interface {
	Error(...interface{})
}

// WorkerPool of workers
type WorkerPool struct {
	logger         Logger
	queue          QueueInterface
	countWorkers   uint32        // count of workers
	periodicity    time.Duration // period for check task in queue
	quit           chan struct{}
	errors         chan error
	closeWorkersCh chan struct{}
	sync.Mutex
	sync.WaitGroup
}

type Option func(*WorkerPool)

func WithPeriodicity(duration time.Duration) Option {
	return func(w *WorkerPool) {
		w.periodicity = duration
	}
}

func WithWorkers(count uint32) Option {
	return func(w *WorkerPool) {
		w.countWorkers = count
	}
}

// NewWorkerPool constructor for create WorkerPool
func NewWorkerPool(queue QueueInterface, logger Logger, opts ...Option) *WorkerPool {
	wp := WorkerPool{
		logger:         logger,
		queue:          queue,
		countWorkers:   10,
		periodicity:    100 * time.Millisecond,
		closeWorkersCh: make(chan struct{}),
		quit:           make(chan struct{}),
		errors:         make(chan error),
	}

	for _, opt := range opts {
		opt(&wp)
	}

	return &wp
}

// Run worker pool
func (w *WorkerPool) Run() {
	go func() {
		for err := range w.errors {
			w.logger.Error(err)
		}
	}()

	w.Add(int(w.countWorkers))
	for i := 0; i < int(w.countWorkers); i++ {
		go w.work()
	}
	<-w.quit
}

func (w *WorkerPool) work() {
	ticker := time.NewTicker(w.periodicity)
loop:
	for {
		select {
		case <-ticker.C:
			task := w.queue.GetTask()
			if task == nil {
				continue
			}
			if err := task.Exec(); err != nil {
				w.errors <- err
				if task.Attempts() != 0 {
					w.queue.AddTask(task)
				}
			}

		case <-w.closeWorkersCh:
			break loop
		}
	}
	w.Done()
}

// Shutdown - the worker will not stop until he has completed all the unfinished tasks
// or the timeout does not expire
func (w *WorkerPool) Shutdown(timeout time.Duration) error {
	ok := make(chan struct{})
	go func() {
		for i := 0; i < int(w.countWorkers); i++ {
			w.closeWorkersCh <- struct{}{}
		}
		w.Wait()
		ok <- struct{}{}
	}()

	select {
	case <-ok:
		close(w.quit)
		return nil
	case <-time.After(timeout):
		close(w.quit)
		return fmt.Errorf(`taskmanager: timeout error`)
	}
}
