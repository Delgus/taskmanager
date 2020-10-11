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
	logger            Logger
	queue             QueueInterface
	wg                sync.WaitGroup
	maxWorkers        int                // count of workers
	periodicityTicker *time.Ticker       // period for check task in queue
	closeTaskCh       chan struct{}      // channel for stopped getting of tasks
	taskCh            chan TaskInterface // channel for tasks
	quit              chan struct{}
}

// NewWorkerPool constructor for create WorkerPool
// maxWorkers - max count of workers
// periodicity - period for check task in queue
func NewWorkerPool(queue QueueInterface, maxWorkers int, periodicity time.Duration, logger Logger) *WorkerPool {
	return &WorkerPool{
		logger:            logger,
		queue:             queue,
		maxWorkers:        maxWorkers,
		periodicityTicker: time.NewTicker(periodicity),
		closeTaskCh:       make(chan struct{}),
		taskCh:            make(chan TaskInterface),
		quit:              make(chan struct{}),
	}
}

// Run worker pool
func (w *WorkerPool) Run() {
	go func() {
		for {
			select {
			case <-w.closeTaskCh:
				close(w.taskCh)
				w.periodicityTicker.Stop()
			case <-w.periodicityTicker.C:
				task, err := w.queue.GetTask()
				if err != nil {
					w.logger.Error(fmt.Errorf(`can not get task from queue: %w`, err))
				}
				if task != nil {
					w.taskCh <- task
				}
			}
		}
	}()

	w.wg.Add(w.maxWorkers)
	for i := 0; i < w.maxWorkers; i++ {
		go w.work()
	}
	<-w.quit
}

func (w *WorkerPool) work() {
	for task := range w.taskCh {
		for task.Attempts() != 0 {
			if err := task.Exec(); err != nil {
				w.logger.Error(err)
			}
		}
	}
	w.wg.Done()
}

// Shutdown - the worker will not stop until he has completed all the unfinished tasks
// or the timeout does not expire
func (w *WorkerPool) Shutdown(timeout time.Duration) error {
	w.closeTaskCh <- struct{}{}

	ok := make(chan struct{})
	go func() {
		w.wg.Wait()
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
