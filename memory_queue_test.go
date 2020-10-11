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

	q := NewMemoryQueue()
	for _, task := range tasks {
		if err := q.AddTask(task); err != nil {
			t.Errorf(`unexpected error: %s`, err.Error())
		}
	}

	highest, _ := q.GetTask()
	priority := highest.Priority()
	if priority != HighestPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, HighestPriority, priority)
	}

	high, _ := q.GetTask()
	priority = high.Priority()
	if priority != HighPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, HighPriority, priority)
	}

	middle, _ := q.GetTask()
	priority = middle.Priority()
	if priority != MiddlePriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, MiddlePriority, priority)
	}

	low, _ := q.GetTask()
	priority = low.Priority()
	if priority != LowPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, LowPriority, priority)
	}

	lowest, _ := q.GetTask()
	priority = lowest.Priority()
	if priority != LowestPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, LowestPriority, priority)
	}
}

func TestGetTaskFromMemoryQueue(t *testing.T) {
	q := NewMemoryQueue()
	if task, _ := q.GetTask(); task != nil {
		t.Error(`unexpected TaskInterface, expect nil`)
	}
	testTask := NewTask(HighestPriority, func() error { return nil })
	if err := q.AddTask(testTask); err != nil {
		t.Errorf(`unexpected error: %s`, err.Error())
	}
	taskFromQueue, _ := q.GetTask()
	if testTask != taskFromQueue {
		t.Error(`one task from queue is not equal task put in queue `)
	}
}

func TestCountTasksForMemoryQueue(t *testing.T) {
	q := NewMemoryQueue()

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

		if err := q.AddTask(tasks[0]); err != nil {
			t.Errorf(`unexpected error: %s`, err.Error())
		}
	}
	var tasksOut int
	for {
		task, _ := q.GetTask()
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
	q := NewMemoryQueue()

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
		_ = q.AddTask(t)
	}

	for i := 1; i <= taskIn; i++ {
		task, _ := q.GetTask()
		_ = task.Exec()
		if i != number {
			t.Error("not right order")
		}
	}
}

func TestRaceConditionForMemoryQueue(t *testing.T) {
	q := NewMemoryQueue()
	go func() {
		if err := q.AddTask(NewTask(HighestPriority, func() error { return nil })); err != nil {
			t.Errorf(`unexpected error: %s`, err.Error())
		}
	}()
	_, _ = q.GetTask()
}

func BenchmarkMemoryQueue_AddTask(b *testing.B) {
	queue := NewMemoryQueue()
	for n := 0; n < b.N; n++ {
		_ = queue.AddTask(NewTask(HighPriority, func() error {
			return nil
		}))
		_ = queue.AddTask(NewTask(LowPriority, func() error {
			return nil
		}))
		_ = queue.AddTask(NewTask(MiddlePriority, func() error {
			return nil
		}))
		_ = queue.AddTask(NewTask(LowestPriority, func() error {
			return nil
		}))
		_ = queue.AddTask(NewTask(HighestPriority, func() error {
			return nil
		}))
	}
}

func BenchmarkSliceQueue_GetTask(b *testing.B) {
	queue := NewMemoryQueue()
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		for i := 0; i < 200; i++ {
			_ = queue.AddTask(NewTask(HighPriority, func() error {
				return nil
			}))
			_ = queue.AddTask(NewTask(LowPriority, func() error {
				return nil
			}))
			_ = queue.AddTask(NewTask(MiddlePriority, func() error {
				return nil
			}))
			_ = queue.AddTask(NewTask(LowestPriority, func() error {
				return nil
			}))
			_ = queue.AddTask(NewTask(HighestPriority, func() error {
				return nil
			}))
		}
		b.StartTimer()
		for i := 0; i < 200; i++ {
			_, _ = queue.GetTask()
			_, _ = queue.GetTask()
			_, _ = queue.GetTask()
			_, _ = queue.GetTask()
			_, _ = queue.GetTask()
		}
	}
}
