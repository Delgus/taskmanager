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

	worker, err := NewWorkerPool(q, WithPollTaskInterval(20*time.Millisecond))
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}

	go worker.Run()

	// 20ms * 10tasks = 200ms
	// wait when workers got all tasks > 200ms
	time.Sleep(time.Millisecond * 201)

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
	workerPool, err := NewWorkerPool(q, WithPollTaskInterval(10*time.Millisecond))
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	go workerPool.Run()
	// 10ms * 1task = 10
	// wait when worker got task > 10
	time.Sleep(time.Millisecond * 11)

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
	workerPool, err := NewWorkerPool(q, WithErrors(), WithPollTaskInterval(10*time.Millisecond))
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	go workerPool.Run()

	// 10ms * 1task = 10
	// wait when worker got task > 10
	time.Sleep(time.Millisecond * 1000)

	err = <-workerPool.Errors
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
	workerPool, err := NewWorkerPool(q, WithPollTaskInterval(10*time.Millisecond))
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	go workerPool.Run()

	// 10ms * (1+5)task = 60
	// wait when worker got task > 60
	time.Sleep(time.Millisecond * 61)

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
	workerPool, err := NewWorkerPool(
		NewQueue(),
		WithPollTaskInterval(300*time.Millisecond),
		WithMinWorkers(20),
		WithMaxWorkers(40),
		WithErrors(),
	)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if workerPool.pollTaskInterval != 300*time.Millisecond {
		t.Errorf("unexpected pollTaskInterval - %d", workerPool.pollTaskInterval)
	}
	if workerPool.minWorkers != 20 {
		t.Errorf("unexpected minWorkers - %d", workerPool.minWorkers)
	}
	if workerPool.returnErr != true {
		t.Errorf("unexpected returnErr - %v", workerPool.returnErr)
	}
	if workerPool.maxWorkers != 40 {
		t.Errorf("unexpected maxWorkers - %d", workerPool.maxWorkers)
	}

	_, err = NewWorkerPool(nil)
	if err == nil {
		t.Errorf("expected err. queue nil!")
	}

	_, err = NewWorkerPool(NewQueue(), WithMaxWorkers(3), WithMinWorkers(5))
	if err == nil {
		t.Errorf("expected err. minWorkers > maxWorkers")
	}

	_, err = NewWorkerPool(NewQueue(), WithMinWorkers(1))
	if err == nil {
		t.Errorf("expected err. minWorkers < 2 ")
	}
}

func TestAddAndRemoveWorkers(t *testing.T) {
	q := NewQueue()
	workerPool, err := NewWorkerPool(
		q,
		WithPollTaskInterval(10*time.Millisecond),
		WithMinWorkers(2),
		WithMaxWorkers(5),
	)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	taskCreate := func(ms time.Duration) *Task {
		return NewTask(HighestPriority, func() error {
			time.Sleep(ms)
			return nil
		})
	}
	testTime := 60 * time.Millisecond
	for i := 0; i < 10; i++ {
		q.AddTask(taskCreate(testTime))
		testTime -= 10 * time.Millisecond
	}

	go workerPool.Run()
	// wait 30+ ms (add 3 workers)
	// and wait more 30 ms
	time.Sleep(time.Millisecond * 61)

	workerPool.Lock()
	if len(workerPool.workers) != 5 {
		t.Errorf(`unexpected count of current workers - %d `, len(workerPool.workers))
	}
	workerPool.Unlock()

	// wait minimize count
	time.Sleep(time.Millisecond * 100)
	workerPool.Lock()
	if len(workerPool.workers) != 2 {
		t.Errorf(`unexpected count of current workers - %d `, len(workerPool.workers))
	}
	workerPool.Unlock()

	if err := workerPool.Shutdown(context.Background()); err != nil {
		t.Errorf(`unexpected shutdown error - %v`, err)
	}
}
