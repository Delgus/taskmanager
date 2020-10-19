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

	// async adding tasks
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		for {
			<-ticker.C
			q.AddTask(taskmanager.NewTask(taskmanager.HighestPriority, func() error { time.Sleep(1 * time.Second); return nil }))
			q.AddTask(taskmanager.NewTask(taskmanager.MiddlePriority, func() error { return fmt.Errorf("oops") }))
			q.AddTask(taskmanager.NewTask(taskmanager.LowestPriority, func() error { return nil }))
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// if need in log errors use option WithErrors and read from channel Errors
	wp, _ := taskmanager.NewWorkerPool(q, taskmanager.WithErrors())
	go func() {
		for err := range wp.Errors {
			log.Println("err:", err)
		}
	}()
	go wp.Run()
	log.Print("worker started")

	<-done
	log.Print("worker stopped")

	if err := wp.Shutdown(context.Background()); err != nil {
		log.Fatalf("worker shutdown failed: %v", err)
	}
	log.Print("worker exited properly")
}
