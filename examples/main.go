package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/delgus/taskmanager"
)

func main() {
	q := taskmanager.NewQueue()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			<-ticker.C
			q.AddTask(taskmanager.NewTask(taskmanager.HighestPriority, func() error { time.Sleep(1 * time.Second); return nil }))
			q.AddTask(taskmanager.NewTask(taskmanager.MiddlePriority, func() error { return fmt.Errorf("oops") }))
			q.AddTask(taskmanager.NewTask(taskmanager.LowestPriority, func() error { return nil }))
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wp, _ := taskmanager.NewWorkerPool(q, taskmanager.WithErrors())
	go wp.Run()
	go func() {
		for err := range wp.Errors {
			log.Println("err:", err)
		}
	}()
	log.Print("worker started")

	<-done
	log.Print("worker stopped")

	if err := wp.Shutdown(context.Background()); err != nil {
		log.Fatalf("worker shutdown failed:%+v", err)
	}
	log.Print("worker exited properly")
}
