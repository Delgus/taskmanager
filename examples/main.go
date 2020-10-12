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

type logger struct{}

func (l *logger) Error(interfaces ...interface{}) {
	for _, i := range interfaces {
		log.Println(i)
	}
}

func main() {
	q := taskmanager.NewMemoryQueue()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			<-ticker.C
			q.AddTask(taskmanager.NewTask(taskmanager.HighestPriority, func() error { time.Sleep(3 * time.Second); return nil }))
			q.AddTask(taskmanager.NewTask(taskmanager.MiddlePriority, func() error { return nil }))
			q.AddTask(taskmanager.NewTask(taskmanager.LowestPriority, func() error { return nil }))
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wp := taskmanager.NewWorkerPool(q, new(logger))
	go wp.Run()
	log.Print("worker started")

	<-done
	log.Print("worker stopped")

	if err := wp.Shutdown(context.Background()); err != nil {
		log.Fatalf("worker shutdown failed:%+v", err)
	}
	log.Print("worker exited properly")
}
