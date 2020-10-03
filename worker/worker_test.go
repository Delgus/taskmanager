package worker

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/delgus/taskmanager"
	"github.com/delgus/taskmanager/memheap"
)

func TestWorkerPool(t *testing.T) {
	q := new(memheap.Queue)

	var workCounter int64

	var countTasks = 5

	testTask := taskmanager.NewTask(taskmanager.HighestPriority, func() error {
		atomic.AddInt64(&workCounter, 1)
		time.Sleep(time.Second * 2)
		return nil
	})

	for i := 0; i < countTasks; i++ {
		q.AddTask(testTask)
	}

	worker := NewPool(q, 10, time.Millisecond)

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
	q := new(memheap.Queue)

	testTask := taskmanager.NewTask(taskmanager.HighestPriority, func() error {
		time.Sleep(time.Second * 10)
		return nil
	})
	q.AddTask(testTask)
	workerPool := NewPool(q, 2, time.Millisecond)
	go workerPool.Run()
	time.Sleep(time.Second)
	if err := workerPool.Shutdown(time.Second); err == nil {
		t.Error(`expected timeout error`)
	}
}
