package memory

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/delgus/taskmanager"
)

func TestPriority(t *testing.T) {
	tasks := []*Task{
		NewTask(taskmanager.HighestPriority, func() error { return nil }),
		NewTask(taskmanager.HighPriority, func() error { return nil }),
		NewTask(taskmanager.MiddlePriority, func() error { return nil }),
		NewTask(taskmanager.LowPriority, func() error { return nil }),
		NewTask(taskmanager.LowestPriority, func() error { return nil }),
	}

	// shuffle tasks
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tasks), func(i, j int) { tasks[i], tasks[j] = tasks[j], tasks[i] })

	q := new(Queue)
	for _, t := range tasks {
		q.AddTask(t)
	}

	highest := q.GetTask()
	priority := highest.Priority()
	if priority != taskmanager.HighestPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, taskmanager.HighestPriority, priority)
	}

	high := q.GetTask()
	priority = high.Priority()
	if priority != taskmanager.HighPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, taskmanager.HighPriority, priority)
	}

	middle := q.GetTask()
	priority = middle.Priority()
	if priority != taskmanager.MiddlePriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, taskmanager.MiddlePriority, priority)
	}

	low := q.GetTask()
	priority = low.Priority()
	if priority != taskmanager.LowPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, taskmanager.LowPriority, priority)
	}

	lowest := q.GetTask()
	priority = lowest.Priority()
	if priority != taskmanager.LowestPriority {
		t.Errorf(`unexpected priority: expect %d get %d"`, taskmanager.LowestPriority, priority)
	}
}

func TestGetTask(t *testing.T) {
	q := new(Queue)
	if q.GetTask() != nil {
		t.Error(`unexpected TaskInterface, expect nil`)
	}
	testTask := NewTask(taskmanager.HighestPriority, func() error { return nil })
	q.AddTask(testTask)
	taskFromQueue := q.GetTask()
	if testTask != taskFromQueue {
		t.Error(`one task from queue is not equal task put in queue `)
	}
}

func TestCountTasks(t *testing.T) {
	q := new(Queue)

	tasksIn := 64
	tasks := []*Task{
		NewTask(taskmanager.HighestPriority, func() error { return nil }),
		NewTask(taskmanager.HighPriority, func() error { return nil }),
		NewTask(taskmanager.MiddlePriority, func() error { return nil }),
		NewTask(taskmanager.LowPriority, func() error { return nil }),
		NewTask(taskmanager.LowestPriority, func() error { return nil }),
	}

	for i := 0; i < tasksIn; i++ {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(tasks), func(i, j int) { tasks[i], tasks[j] = tasks[j], tasks[i] })

		q.AddTask(tasks[0])
	}
	var tasksOut int
	for q.GetTask() != nil {
		tasksOut++
	}

	if tasksIn != tasksOut {
		t.Errorf(`unexpected out tasks - %d, expect - %d`, tasksOut, tasksIn)
	}
}

func TestErrorTask(t *testing.T) {
	failFlag := false

	task := NewTask(taskmanager.HighestPriority, func() error { return fmt.Errorf(`test error`) })
	task.OnEvent(taskmanager.FailedEvent, func() {
		failFlag = true
	})
	err := task.Exec()

	if err == nil {
		t.Errorf(`expected error!`)
	}

	if !failFlag {
		t.Errorf(`expected execution of handler for failed event!`)
	}
}

func TestRaceCondition(t *testing.T) {
	q := new(Queue)
	go func() {
		q.AddTask(NewTask(taskmanager.HighestPriority, func() error { return nil }))
	}()
	q.GetTask()
}

func TestOnEvent(t *testing.T) {
	ed := NewTask(taskmanager.HighestPriority, func() error { return nil })

	eventFlag := false
	ed.OnEvent(taskmanager.BeforeExecEvent, func() {
		eventFlag = true
	})

	ed.EmitEvent(taskmanager.AfterExecEvent)

	if eventFlag {
		t.Errorf(`unexpected execution of handler`)
	}

	ed.EmitEvent(taskmanager.BeforeExecEvent)

	if !eventFlag {
		t.Errorf(`handler not execute`)
	}
}
