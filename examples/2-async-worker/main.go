package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/delgus/taskmanager"
	"github.com/delgus/taskmanager/memory"
	"github.com/delgus/taskmanager/worker"
)

func main() {
	q := new(memory.Queue)
	// async adding tasks
	stopAddTasks := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond * 100)
		for {
			select {
			case <-ticker.C:
				task := memory.NewTask(taskmanager.HighestPriority, func() error {
					time.Sleep(time.Second * 5)
					fmt.Println("i highest! good work!")
					return nil
				})
				task.OnEvent(taskmanager.BeforeExecEvent, func() {
					fmt.Println("i highest! i before execution!")
				})
				task.OnEvent(taskmanager.AfterExecEvent, func() {
					fmt.Println("i highest! i after execution!")
				})
				q.AddTask(task)
			case <-stopAddTasks:
				return
			}
		}
	}()

	// process tasks in 10 goroutine
	worker := worker.NewPool(q, 10, time.Millisecond*50)

	// press CTRL + C to stop the worker
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		stopAddTasks <- struct{}{}
		if err := worker.Shutdown(time.Second * 5); err != nil {
			fmt.Println(`error by stopping:` + err.Error())
		}
		fmt.Println(`stopping worker pool`)
	}()

	worker.Run()

}
