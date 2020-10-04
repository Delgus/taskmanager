package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/delgus/taskmanager"
	"github.com/sirupsen/logrus"
)

func main() {
	q := new(taskmanager.HeapQueue)
	// async adding tasks
	stopAddTasks := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond * 100)
		for {
			select {
			case <-ticker.C:
				task := taskmanager.NewTask(taskmanager.HighestPriority, func() error {
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
	workerPool := taskmanager.NewWorkerPool(q, 10, time.Millisecond*50, logrus.New())

	// press CTRL + C to stop the worker
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		stopAddTasks <- struct{}{}
		if err := workerPool.Shutdown(time.Second * 5); err != nil {
			fmt.Println(`error by stopping:` + err.Error())
		}
		fmt.Println(`stopping worker pool`)
	}()

	workerPool.Run()
}
