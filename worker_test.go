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
	var countTasks = 10 // ==  count workers

	for i := 0; i < countTasks; i++ {
		q.AddTask(NewTask(HighestPriority, calcTask))
	}

	worker := NewWorkerPool(q, false)
	if err := worker.SetMinWorkers(10); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	worker.SetPollTaskDuration(20 * time.Millisecond)

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
	workerPool := NewWorkerPool(q, true)
	workerPool.SetPollTaskDuration(10 * time.Millisecond)
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
	workerPool := NewWorkerPool(q, true)
	workerPool.SetPollTaskDuration(10 * time.Millisecond)
	go workerPool.Run()

	// 10ms * 1task = 10
	// wait when worker got task > 10
	time.Sleep(time.Millisecond * 1000)

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
	workerPool := NewWorkerPool(q, false)
	workerPool.SetPollTaskDuration(10 * time.Millisecond)
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

func TestSetOptions(t *testing.T) {
	w := NewWorkerPool(NewQueue(), false)
	go w.Run()

	_ = w.SetMinWorkers(15)
	if w.GetMinWorkers() != 15 {
		t.Errorf("unexpected minWorkers - %d", w.minWorkers)
	}

	_ = w.SetMaxWorkers(100)
	if w.GetMaxWorkers() != 100 {
		t.Errorf("unexpected maxWorkers - %d", w.minWorkers)
	}
}

func TestAddAndRemoveWorkers(t *testing.T) {
	q := NewQueue()
	w := NewWorkerPool(q, false)

	// for fastest tests
	w.SetPollTaskDuration(10 * time.Millisecond)

	if err := w.SetMinWorkers(2); err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if err := w.SetMaxWorkers(5); err != nil {
		t.Errorf("unexpected error %v", err)
	}

	go w.Run()

	// wait running
	time.Sleep(100 * time.Millisecond)

	// check workers - 2
	current := w.GetCountWorkers()
	if current != 2 {
		t.Errorf(`unexpected count of current workers - %d `, current)
	}

	// add 5 hard tasks - 50ms
	for i := 0; i < 10; i++ {
		q.AddTask(NewTask(HighestPriority, func() error {
			time.Sleep(60 * time.Millisecond)
			return nil
		}))
	}

	// wait 58ms. expected 5 workers (after 20 ms every 10ms add 1 worker, max count - 5)
	time.Sleep(58 * time.Millisecond)
	current = w.GetCountWorkers()
	if current != 5 {
		t.Errorf(`unexpected count of current workers - %d `, current)
	}

	w.SetGCDuration(1 * time.Millisecond)

	// wait 100ms. expected 2 workers
	time.Sleep(120 * time.Millisecond)
	current = w.GetCountWorkers()
	if current != 2 {
		t.Errorf(`unexpected count of current workers - %d `, current)
	}

	if err := w.Shutdown(context.Background()); err != nil {
		t.Errorf(`unexpected shutdown error - %v`, err)
	}
}

func TestWorkerPool_setMaxWorkers(t *testing.T) {
	wp1 := NewWorkerPool(NewQueue(), false)
	wp1.minWorkers = 10
	wp2 := NewWorkerPool(NewQueue(), false)

	type args struct {
		count uint32
	}
	tests := []struct {
		name    string
		w       *WorkerPool
		args    args
		wantErr bool
	}{
		{
			name: " < min workers",
			w:    wp1,
			args: args{
				count: 9,
			},
			wantErr: true,
		},
		{
			name: "good",
			w:    wp2,
			args: args{
				count: 30,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.w.SetMaxWorkers(tt.args.count); (err != nil) != tt.wantErr {
				t.Errorf("setMaxWorkers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkerPool_setMinWorkers(t *testing.T) {
	wp1 := NewWorkerPool(NewQueue(), false)
	wp2 := NewWorkerPool(NewQueue(), false)
	type args struct {
		count uint32
	}
	tests := []struct {
		name    string
		w       *WorkerPool
		args    args
		wantErr bool
	}{
		{
			name: "<2",
			w:    wp1,
			args: args{
				count: 1,
			},
			wantErr: true,
		},
		{
			name: ">default max 25",
			w:    wp2,
			args: args{
				count: 26,
			},
			wantErr: true,
		},
		{
			name: "good",
			w:    wp1,
			args: args{
				count: 5,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.w.SetMinWorkers(tt.args.count); (err != nil) != tt.wantErr {
				t.Errorf("setMinWorkers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
