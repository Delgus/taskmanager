package taskmanager

import (
	"math/rand"
	"testing"
	"time"
)

func TestPriorityInMemoryQueue(t *testing.T) {
	tasks := []*Task{
		NewTask(HighestPriority, func() error { return nil }),
		NewTask(HighPriority, func() error { return nil }),
		NewTask(MiddlePriority, func() error { return nil }),
		NewTask(LowPriority, func() error { return nil }),
		NewTask(LowestPriority, func() error { return nil }),
	}

	// shuffle tasks
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tasks), func(i, j int) { tasks[i], tasks[j] = tasks[j], tasks[i] })

	q := NewQueue()
	for _, task := range tasks {
		q.AddTask(task)
	}

	highest := q.GetTask()
	priority := highest.Priority()
	if priority != HighestPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, HighestPriority, priority)
	}

	high := q.GetTask()
	priority = high.Priority()
	if priority != HighPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, HighPriority, priority)
	}

	middle := q.GetTask()
	priority = middle.Priority()
	if priority != MiddlePriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, MiddlePriority, priority)
	}

	low := q.GetTask()
	priority = low.Priority()
	if priority != LowPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, LowPriority, priority)
	}

	lowest := q.GetTask()
	priority = lowest.Priority()
	if priority != LowestPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, LowestPriority, priority)
	}
}

func TestGetTaskFromMemoryQueue(t *testing.T) {
	q := NewQueue()
	if task := q.GetTask(); task != nil {
		t.Error(`unexpected TaskInterface, expect nil`)
	}
	testTask := NewTask(HighestPriority, func() error { return nil })
	q.AddTask(testTask)
	taskFromQueue := q.GetTask()
	if testTask != taskFromQueue {
		t.Error(`one task from queue is not equal task put in queue `)
	}
}

func TestCountTasksForMemoryQueue(t *testing.T) {
	q := NewQueue()

	tasksIn := 64
	tasks := []*Task{
		NewTask(HighestPriority, func() error { return nil }),
		NewTask(HighPriority, func() error { return nil }),
		NewTask(MiddlePriority, func() error { return nil }),
		NewTask(LowPriority, func() error { return nil }),
		NewTask(LowestPriority, func() error { return nil }),
	}

	for i := 0; i < tasksIn; i++ {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(tasks), func(i, j int) { tasks[i], tasks[j] = tasks[j], tasks[i] })

		q.AddTask(tasks[0])
	}
	var tasksOut int
	for {
		task := q.GetTask()
		if task == nil {
			break
		}
		tasksOut++
	}

	if tasksIn != tasksOut {
		t.Errorf(`unexpected out tasks - %d, expect - %d`, tasksOut, tasksIn)
	}
}

func TestFIFOForMemoryQueue(t *testing.T) {
	q := NewQueue()

	var number int
	taskIn := 5
	tasks := []*Task{
		NewTask(HighestPriority, func() error { number = 1; return nil }),
		NewTask(HighestPriority, func() error { number = 2; return nil }),
		NewTask(HighestPriority, func() error { number = 3; return nil }),
		NewTask(HighestPriority, func() error { number = 4; return nil }),
		NewTask(HighestPriority, func() error { number = 5; return nil }),
	}
	for _, t := range tasks {
		q.AddTask(t)
	}

	for i := 1; i <= taskIn; i++ {
		task := q.GetTask()
		_ = task.Exec()
		if i != number {
			t.Error("not right order")
		}
	}
}

func TestRaceConditionForMemoryQueue(t *testing.T) {
	q := NewQueue()
	go func() {
		q.AddTask(NewTask(HighestPriority, func() error { return nil }))
	}()
	q.GetTask()
}

func BenchmarkMemoryQueue_AddTask(b *testing.B) {
	queue := NewQueue()
	task := NewTask(LowestPriority, func() error {
		return nil
	})
	for n := 0; n < b.N; n++ {
		queue.AddTask(task)
	}
}

func BenchmarkSliceQueue_GetTask(b *testing.B) {
	queue := NewQueue()
	task := NewTask(LowestPriority, func() error {
		return nil
	})
	for i := 0; i < 1000; i++ {
		queue.AddTask(task)
	}
	b.ResetTimer()
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		queue.GetTask()
	}
}
