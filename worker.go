package taskmanager

import (
	"context"
	"sync"
	"time"
)

// WorkerPool of workers
type WorkerPool struct {
	queue            QueueInterface
	countWorkers     uint32        // count of workers
	pollTaskInterval time.Duration // period for check task in queue
	quit             chan struct{}
	closeWorkersCh   chan struct{}
	returnErr        bool
	Errors           chan error
	sync.WaitGroup
}

type Option func(*WorkerPool)

func WithPollTaskInterval(duration time.Duration) Option {
	return func(w *WorkerPool) {
		w.pollTaskInterval = duration
	}
}

func WithWorkers(count uint32) Option {
	return func(w *WorkerPool) {
		w.countWorkers = count
	}
}

func WithErrors() Option {
	return func(w *WorkerPool) {
		w.returnErr = true
	}
}

// NewWorkerPool constructor for create WorkerPool
func NewWorkerPool(queue QueueInterface, opts ...Option) *WorkerPool {
	wp := WorkerPool{
		queue:            queue,
		countWorkers:     10,
		pollTaskInterval: 200 * time.Millisecond,
		closeWorkersCh:   make(chan struct{}),
		quit:             make(chan struct{}),
		Errors:           make(chan error),
	}

	for _, opt := range opts {
		opt(&wp)
	}

	return &wp
}

// Run worker pool
func (w *WorkerPool) Run() {
	w.Add(int(w.countWorkers))
	for i := 0; i < int(w.countWorkers); i++ {
		go w.work()
	}
	<-w.quit
}

func (w *WorkerPool) work() {
	defer w.Done()
	ticker := time.NewTicker(w.pollTaskInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			task := w.queue.GetTask()
			if task == nil {
				continue
			}
			if err := task.Exec(); err != nil {
				if w.returnErr {
					w.Errors <- err
				}
				if task.Attempts() != 0 {
					w.queue.AddTask(task)
				}
			}
		case <-w.closeWorkersCh:
			return
		}
	}
}

// Shutdown - the worker will not stop until he has completed all the unfinished tasks
// or the timeout does not expire
func (w *WorkerPool) Shutdown(ctx context.Context) error {
	defer close(w.quit)

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
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
