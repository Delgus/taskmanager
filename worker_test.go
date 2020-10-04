package taskmanager

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

type fakeLogger struct {
	buffer interface{}
}

func (f *fakeLogger) Error(info ...interface{}) {
	f.buffer = info[0]
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
	q := new(HeapQueue)

	var workCounter int64

	var countTasks = 5

	testTask := NewTask(HighestPriority, func() error {
		atomic.AddInt64(&workCounter, 1)
		time.Sleep(time.Second * 2)
		return nil
	})

	for i := 0; i < countTasks; i++ {
		if err := q.AddTask(testTask); err != nil {
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
		t.Error(`not all tasks completed`)
	}
}

func TestWorkerPool_Shutdown(t *testing.T) {
	q := new(HeapQueue)

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
	q := new(HeapQueue)

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
	if fakeLogger.buffer != oops {
		t.Errorf(`expected logger text: %v got %v`, fakeLogger.buffer, oops)
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
	if errFromLogger, ok := fakeLogger.buffer.(error); ok {
		if !errors.Is(errFromLogger, oops) {
			t.Errorf(`expected logger text: %v got %v`, fakeLogger.buffer, oops)
		}
	} else {
		t.Errorf(`expected error. got %v`, fakeLogger.buffer)
	}
}
