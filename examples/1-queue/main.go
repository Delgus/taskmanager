package main

import (
	"fmt"

	"github.com/delgus/taskmanager"
	"github.com/delgus/taskmanager/memory"
)

func main() {
	tq := new(memory.Queue)

	lTask := memory.NewTask(taskmanager.LowestPriority, func() error {
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

	hTask := memory.NewTask(taskmanager.HighestPriority, func() error {
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

	mTask := memory.NewTask(taskmanager.MiddlePriority, func() error {
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

	bTask := memory.NewTask(taskmanager.HighestPriority, func() error {
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
		task := tq.GetTask()
		if task == nil {
			break
		}
		task.Exec()
	}
}
