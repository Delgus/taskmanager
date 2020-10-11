package taskmanager

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// logger mock
type fakeLogger struct{ err chan error }

func newFakeLogger() *fakeLogger {
	return &fakeLogger{
		err: make(chan error, 100),
	}
}

func (f *fakeLogger) Error(info ...interface{}) {
	for _, i := range info {
		if v, ok := i.(error); ok {
			f.err <- v
		}
	}
}

func (f *fakeLogger) getError() error {
	if len(f.err) > 0 {
		return <-f.err
	}
	return nil
}

func TestWorkerPool(t *testing.T) {
	q := NewMemoryQueue()

	var workCounter int64

	var calcTask = func() error {
		atomic.AddInt64(&workCounter, 1)
		time.Sleep(time.Second)
		return nil
	}
	var countTasks = 10 // == default count workers

	for i := 0; i < countTasks; i++ {
		q.AddTask(NewTask(HighestPriority, calcTask))
	}

	worker := NewWorkerPool(q, newFakeLogger())

	go worker.Run()

	// wait when workers got all tasks
	time.Sleep(time.Millisecond * 300)

	if err := worker.Shutdown(time.Second); err != nil {
		t.Errorf(`unexpected error - %v`, err)
	}

	if workCounter != int64(countTasks) {
		t.Errorf(`different count: adding - %d executing - %d`, countTasks, workCounter)
	}
}

func TestWorkerPool_Shutdown(t *testing.T) {
	q := NewMemoryQueue()

	testTask := NewTask(HighestPriority, func() error {
		time.Sleep(time.Second * 2)
		return nil
	})
	q.AddTask(testTask)
	workerPool := NewWorkerPool(q, newFakeLogger())
	go workerPool.Run()
	// wait when workers got all tasks
	time.Sleep(time.Millisecond * 300)
	if err := workerPool.Shutdown(time.Second); err == nil {
		t.Error(`expected timeout error`)
	}
}

func TestWorkerLogTaskError(t *testing.T) {
	q := NewMemoryQueue()

	oops := fmt.Errorf("oops")
	testTask := NewTask(HighestPriority, func() error {
		return oops
	})
	q.AddTask(testTask)
	fakeLogger := newFakeLogger()
	workerPool := NewWorkerPool(q, fakeLogger)
	go workerPool.Run()

	// wait when workers got all tasks
	time.Sleep(time.Millisecond * 300)

	if err := workerPool.Shutdown(time.Millisecond * 200); err != nil {
		t.Error(`unexpected timeout error`)
	}
	if err := fakeLogger.getError(); !errors.Is(err, oops) {
		t.Errorf(`expected error: %v got: %v`, oops, err)
	}
}

func TestTask_Attempts(t *testing.T) {
	q := NewMemoryQueue()

	var task1Counter int64
	task1 := NewTask(HighestPriority, func() error {
		atomic.AddInt64(&task1Counter, 1)
		return nil
	})

	var task2Counter int64
	var attempts uint32 = 5
	task2 := NewTask(HighestPriority, func() error {
		atomic.AddInt64(&task2Counter, 1)
		return fmt.Errorf("oops")
	})
	task2.SetAttempts(attempts)

	q.AddTask(task1)
	q.AddTask(task2)
	workerPool := NewWorkerPool(q, newFakeLogger())
	go workerPool.Run()

	// wait when workers got all tasks
	time.Sleep(time.Millisecond * 300)

	if err := workerPool.Shutdown(time.Second); err != nil {
		t.Error(`unexpected timeout error`)
	}

	if task1Counter != 1 {
		t.Errorf("wrong count of execution. expect - 1, got - %d", task1Counter)
	}
	if task2Counter != int64(attempts) {
		t.Errorf("wrong count of execution. expect - %d, got - %d", attempts, task2Counter)
	}
}

func TestWorkerOptions(t *testing.T) {
	workerPool := NewWorkerPool(
		NewMemoryQueue(),
		newFakeLogger(),
		WithPeriodicity(300*time.Millisecond),
		WithWorkers(20))
	if workerPool.periodicity != 300*time.Millisecond {
		t.Error("unexpected periodicity")
	}
	if workerPool.countWorkers != 20 {
		t.Error("unexpected count of workers")
	}
}
