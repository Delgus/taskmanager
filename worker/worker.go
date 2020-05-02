package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/delgus/taskmanager"
)

// Logger interface
type Logger interface {
	Error(interface{})
}

// Pool of workers
type Pool struct {
	logger            Logger
	queue             taskmanager.Queue
	wg                sync.WaitGroup
	maxWorkers        int                   // count of workers
	periodicityTicker *time.Ticker          // period for check task in queue
	closeTaskCh       chan struct{}         // channel for stopped getting of tasks
	taskCh            chan taskmanager.Task // channel for tasks
	quit              chan struct{}
}

// NewPool constructor for create Pool
// maxWorkers - max count of workers
// periodicity - period for check task in queue
func NewPool(queue taskmanager.Queue, maxWorkers int, periodicity time.Duration) *Pool {
	return &Pool{
		queue:             queue,
		maxWorkers:        maxWorkers,
		periodicityTicker: time.NewTicker(periodicity),
		closeTaskCh:       make(chan struct{}),
		taskCh:            make(chan taskmanager.Task),
		quit:              make(chan struct{}),
	}
}

// Run worker pool
func (w *Pool) Run() {
	go func() {
		for {
			select {
			case <-w.closeTaskCh:
				close(w.taskCh)
				return
			case <-w.periodicityTicker.C:
				if task := w.queue.GetTask(); task != nil {
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

func (w *Pool) work() {
	for task := range w.taskCh {
		if err := task.Exec(); err != nil {
			w.logger.Error(err)
		}
	}
	w.wg.Done()
}

// Shutdown - the worker will not stop until he has completed all the unfinished tasks
// or the timeout does not expire
func (w *Pool) Shutdown(timeout time.Duration) error {
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
