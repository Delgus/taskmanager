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
	worker.SetPollTaskTicker(time.NewTicker(20 * time.Millisecond))

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
	workerPool.SetPollTaskTicker(time.NewTicker(10 * time.Millisecond))
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
	workerPool.SetPollTaskTicker(time.NewTicker(10 * time.Millisecond))
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
	workerPool.SetPollTaskTicker(time.NewTicker(10 * time.Millisecond))
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

	_ = w.SetMinWorkingPercent(33)
	if w.GetMinWorkingPercent() != 33 {
		t.Errorf("unexpected minWorkingPercent- %d", w.minWorkingPercent)
	}

	_ = w.SetMaxWorkingPercent(70)
	if w.GetMaxWorkingPercent() != 70 {
		t.Errorf("unexpected minWorkers - %d", w.maxWorkingPercent)
	}

	newPollTicker := time.NewTicker(33 * time.Second)
	w.SetPollTaskTicker(newPollTicker)
	if w.pollTaskTicker != newPollTicker {
		t.Error("unexpected poll task ticker")
	}

	newScheduleTicker := time.NewTicker(22 * time.Second)
	w.SetScheduleTicker(newScheduleTicker)
	if w.scheduleTaskTicker != newScheduleTicker {
		t.Error("unexpected new schedule ticker")
	}
}

func TestAddAndRemoveWorkers(t *testing.T) {
	q := NewQueue()
	w := NewWorkerPool(q, false)
	w.SetPollTaskTicker(time.NewTicker(10 * time.Millisecond))
	w.SetScheduleTicker(time.NewTicker(5 * time.Millisecond)) // for fastest tests
	if err := w.SetMinWorkers(2); err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if err := w.SetMaxWorkers(5); err != nil {
		t.Errorf("unexpected error %v", err)
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

	go w.Run()
	// wait 30+ ms (add 3 workers)
	// and wait more 30 ms
	time.Sleep(time.Millisecond * 61)

	current, _ := w.getStats()
	if current != 5 {
		t.Errorf(`unexpected count of current workers - %d `, current)
	}

	// wait minimize count
	time.Sleep(time.Millisecond * 100)

	current, _ = w.getStats()
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
			if err := tt.w.setMaxWorkers(tt.args.count); (err != nil) != tt.wantErr {
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
			if err := tt.w.setMinWorkers(tt.args.count); (err != nil) != tt.wantErr {
				t.Errorf("setMinWorkers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkerPool_setMinWorkingPercent(t *testing.T) {
	wp1 := NewWorkerPool(NewQueue(), false)
	wp2 := NewWorkerPool(NewQueue(), false)
	wp2.maxWorkingPercent = 70
	type args struct {
		per uint32
	}
	tests := []struct {
		name    string
		w       *WorkerPool
		args    args
		wantErr bool
	}{
		{
			name: " > 100",
			w:    wp1,
			args: args{
				per: 101,
			},
			wantErr: true,
		},
		{
			name: " > max percent",
			w:    wp2,
			args: args{
				per: 71,
			},
			wantErr: true,
		},
		{
			name: "good",
			w:    wp1,
			args: args{
				per: 45,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.w.setMinWorkingPercent(tt.args.per); (err != nil) != tt.wantErr {
				t.Errorf("setMinWorkingPercent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkerPool_SetMaxWorkingPercent(t *testing.T) {
	wp1 := NewWorkerPool(NewQueue(), false)
	wp2 := NewWorkerPool(NewQueue(), false)
	type args struct {
		per uint32
	}
	tests := []struct {
		name    string
		w       *WorkerPool
		args    args
		wantErr bool
	}{
		{
			name: " > 100",
			w:    wp1,
			args: args{
				per: 101,
			},
			wantErr: true,
		},
		{
			name: " < min percent",
			w:    wp2,
			args: args{
				per: 49,
			},
			wantErr: true,
		},
		{
			name: "good",
			w:    wp1,
			args: args{
				per: 90,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.w.setMaxWorkingPercent(tt.args.per); (err != nil) != tt.wantErr {
				t.Errorf("setMaxWorkingPercent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
