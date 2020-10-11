package taskmanager

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeLogger struct {
	buffer interface{}
	sync.Mutex
}

func (f *fakeLogger) Error(info ...interface{}) {
	f.Lock()
	f.buffer = info[0]
	f.Unlock()
}

func (f *fakeLogger) getBuffer() interface{} {
	f.Lock()
	defer f.Unlock()
	return f.buffer
}

type brokenQueue struct {
	err error
}

func (b *brokenQueue) AddTask(task TaskInterface) error {
	return nil
}

func (b *brokenQueue) GetTask() (TaskInterface, error) {
	return nil, b.err
}

func (b *brokenQueue) addError(err error) {
	b.err = err
}

func TestWorkerPool(t *testing.T) {
	q := NewMemoryQueue()

	var workCounter int64

	var calcTask = func() error {
		atomic.AddInt64(&workCounter, 1)
		time.Sleep(time.Second * 2)
		return nil
	}
	var countTasks = 5

	for i := 0; i < countTasks; i++ {
		if err := q.AddTask(NewTask(HighestPriority, calcTask)); err != nil {
			t.Errorf(`unexpected error: %s`, err.Error())
		}
	}

	worker := NewWorkerPool(q, 10, time.Millisecond, new(fakeLogger))

	go worker.Run()

	// wait pool
	time.Sleep(time.Second * 1)

	if err := worker.Shutdown(time.Second * 3); err != nil {
		t.Error(err)
	}

	if workCounter != int64(countTasks) {
		t.Errorf(`different count: adding - %d executing - %d`, countTasks, workCounter)
	}
}

func TestWorkerPool_Shutdown(t *testing.T) {
	q := NewMemoryQueue()

	testTask := NewTask(HighestPriority, func() error {
		time.Sleep(time.Second * 10)
		return nil
	})
	if err := q.AddTask(testTask); err != nil {
		t.Errorf(`unexpected error: %s`, err.Error())
	}
	workerPool := NewWorkerPool(q, 2, time.Millisecond, new(fakeLogger))
	go workerPool.Run()
	time.Sleep(time.Second)
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
	if err := q.AddTask(testTask); err != nil {
		t.Errorf(`unexpected error: %s`, err.Error())
	}
	fakeLogger := new(fakeLogger)
	workerPool := NewWorkerPool(q, 2, time.Millisecond, fakeLogger)
	go workerPool.Run()
	time.Sleep(time.Second)
	if err := workerPool.Shutdown(time.Second); err != nil {
		t.Error(`unexpected timeout error`)
	}
	if errFromLogger, ok := fakeLogger.getBuffer().(error); ok {
		if !errors.Is(errFromLogger, oops) {
			t.Errorf(`expected logger text: %v got %v`, fakeLogger.buffer, oops)
		}
	} else {
		t.Errorf(`expected error. got %v`, fakeLogger.buffer)
	}
}

func TestWorkerLogErrorByGetTask(t *testing.T) {
	q := new(brokenQueue)
	oops := fmt.Errorf("oops-oops")
	q.addError(oops)
	fakeLogger := new(fakeLogger)
	workerPool := NewWorkerPool(q, 2, time.Millisecond, fakeLogger)
	go workerPool.Run()
	time.Sleep(time.Second)
	if err := workerPool.Shutdown(time.Second); err != nil {
		t.Error(`unexpected timeout error`)
	}
	if errFromLogger, ok := fakeLogger.getBuffer().(error); ok {
		if !errors.Is(errFromLogger, oops) {
			t.Errorf(`expected logger text: %v got %v`, fakeLogger.buffer, oops)
		}
	} else {
		t.Errorf(`expected error. got %v`, fakeLogger.buffer)
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

	if err := q.AddTask(task1); err != nil {
		t.Errorf(`unexpected error: %s`, err.Error())
	}
	if err := q.AddTask(task2); err != nil {
		t.Errorf(`unexpected error: %s`, err.Error())
	}
	fakeLogger := new(fakeLogger)
	workerPool := NewWorkerPool(q, 2, time.Millisecond, fakeLogger)
	go workerPool.Run()
	time.Sleep(time.Second)
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
