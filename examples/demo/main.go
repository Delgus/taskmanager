package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/delgus/taskmanager"
)

func main() {
	q := taskmanager.NewQueue()

	// async adding tasks
	go func() {
		tickerLight := time.NewTicker(200 * time.Millisecond)
		tickerHard := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-tickerLight.C:
				q.AddTask(taskmanager.NewTask(taskmanager.MiddlePriority, func() error { return nil }))
				q.AddTask(taskmanager.NewTask(taskmanager.LowestPriority, func() error { return nil }))
			case <-tickerHard.C:
				for i := 0; i < 20; i++ {
					q.AddTask(taskmanager.NewTask(taskmanager.HighestPriority, func() error { time.Sleep(500 * time.Millisecond); return nil }))
				}
			}
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// if need in log errors use option WithErrors and read from channel Errors
	wp := taskmanager.NewWorkerPool(q, true)
	go func() {
		for err := range wp.Errors {
			log.Println("err:", err)
		}
	}()
	wp.SetPollTaskDuration(200 * time.Millisecond)
	wp.SetGCDuration(5 * time.Second)
	go wp.Run()
	log.Print("worker started")

	//writer := uilive.New()
	// start listening for updates and render
	//writer.Start()

	/*go func() {
		for range time.NewTicker(500 * time.Millisecond).C {
			fmt.Fprintf(writer, "workers - in work - %d / all - %d \n",
				wp.GetCountWorkingWorkers(),
				wp.GetCountWorkers())
		}
	}()*/

	<-done
	// writer.Stop()
	log.Print("worker stopped")

	if err := wp.Shutdown(context.Background()); err != nil {
		log.Fatalf("worker shutdown failed: %v", err)
	}
	log.Print("worker exited properly")
}
