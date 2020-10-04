package main

import (
	"fmt"

	"github.com/delgus/taskmanager"
)

func main() {
	tq := new(taskmanager.HeapQueue)

	lTask := taskmanager.NewTask(taskmanager.LowestPriority, func() error {
		fmt.Println("i lowest! good work!")
		return nil
	})
	lTask.OnEvent(taskmanager.BeforeExecEvent, func() {
		fmt.Println("i lowest! i before execution!")
	})
	lTask.OnEvent(taskmanager.AfterExecEvent, func() {
		fmt.Println("i lowest! i after execution!")
	})
	tq.AddTask(lTask)

	hTask := taskmanager.NewTask(taskmanager.HighestPriority, func() error {
		fmt.Println("i highest! good work!")
		return nil
	})
	hTask.OnEvent(taskmanager.BeforeExecEvent, func() {
		fmt.Println("i highest! i before execution!")
	})
	hTask.OnEvent(taskmanager.AfterExecEvent, func() {
		fmt.Println("i highest! i after execution!")
	})
	tq.AddTask(hTask)

	mTask := taskmanager.NewTask(taskmanager.MiddlePriority, func() error {
		fmt.Println("i middle! good work!")
		return nil
	})
	mTask.OnEvent(taskmanager.BeforeExecEvent, func() {
		fmt.Println("i middle! i before execution!")
	})
	mTask.OnEvent(taskmanager.AfterExecEvent, func() {
		fmt.Println("i middle! i after execution!")
	})
	tq.AddTask(mTask)

	bTask := taskmanager.NewTask(taskmanager.HighestPriority, func() error {
		return fmt.Errorf("i broken! sorry(")
	})
	bTask.OnEvent(taskmanager.FailedEvent, func() {
		fmt.Println("i broke! sorry")
	})
	bTask.OnEvent(taskmanager.BeforeExecEvent, func() {
		fmt.Println("i highest! i before execution!")
	})
	bTask.OnEvent(taskmanager.AfterExecEvent, func() {
		fmt.Println("i highest! i after execution!")
	})
	tq.AddTask(bTask)

	for {
		task, _ := tq.GetTask()
		if task == nil {
			break
		}
		task.Exec()
	}
}
