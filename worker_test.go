package taskmanager

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	q := NewQueue()

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

	worker := NewWorkerPool(q)

	go worker.Run()

	// wait when workers got all tasks
	time.Sleep(time.Millisecond * 300)

	if err := worker.Shutdown(context.Background()); err != nil {
		t.Errorf(`unexpected error - %v`, err)
	}

	if workCounter != int64(countTasks) {
		t.Errorf(`different count: adding - %d executing - %d`, countTasks, workCounter)
	}
}

func TestWorkerPool_Shutdown(t *testing.T) {
	q := NewQueue()

	testTask := NewTask(HighestPriority, func() error {
		time.Sleep(time.Second * 2)
		return nil
	})
	q.AddTask(testTask)
	workerPool := NewWorkerPool(q)
	go workerPool.Run()
	// wait when workers got all tasks
	time.Sleep(time.Millisecond * 300)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := workerPool.Shutdown(ctx); err == nil {
		t.Error(`expected timeout error`)
	}
}

func TestWorkerErrors(t *testing.T) {
	q := NewQueue()

	oops := fmt.Errorf("oops")
	testTask := NewTask(HighestPriority, func() error {
		return oops
	})
	q.AddTask(testTask)
	workerPool := NewWorkerPool(q, WithErrors())
	go workerPool.Run()

	// wait when workers got all tasks
	time.Sleep(time.Millisecond * 100)

	err := <-workerPool.Errors
	if !errors.Is(err, oops) {
		t.Errorf(`expected error: %v got: %v`, oops, err)
	}

	if err := workerPool.Shutdown(context.Background()); err != nil {
		t.Errorf(`unexpected shutdown error - %v`, err)
	}
}

func TestTask_Attempts(t *testing.T) {
	q := NewQueue()

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
	workerPool := NewWorkerPool(q, WithPollTaskInterval(50*time.Millisecond))
	go workerPool.Run()

	// wait when workers got all tasks
	time.Sleep(time.Millisecond * 300)

	if err := workerPool.Shutdown(context.Background()); err != nil {
		t.Errorf(`unexpected shutdown error - %v`, err)
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
		NewQueue(),
		WithPollTaskInterval(300*time.Millisecond),
		WithWorkers(20),
		WithErrors())
	if workerPool.pollTaskInterval != 300*time.Millisecond {
		t.Error("unexpected pollTaskInterval")
	}
	if workerPool.countWorkers != 20 {
		t.Error("unexpected count of workers")
	}
	if workerPool.returnErr != true {
		t.Error("unexpected returnErr")
	}
}
